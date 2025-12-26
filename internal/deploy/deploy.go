package deploy

import (
	"bytes"
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/crypto"
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"youfun/shipyard/pkg/types"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Deployer encapsulates all deployment logic and state
type Deployer struct {
	AppName            string
	HostName           string
	useBuild           string
	Instance           *models.ApplicationInstance
	Application        *models.Application
	Runtime            string // phoenix|node|golang
	Host               *models.SSHHost
	History            *models.DeploymentHistory
	SSHClient          *ssh.Client
	LogBuffer          bytes.Buffer
	tarballPath        string
	md5Hash            string
	Version            string // mix.exs version
	GitCommitSHA       string // Git commit hash
	CurrentReleasePath string // The remote path for the current release
	caddySvc           *caddy.Service
	APIClient          client.APIClient
	Domains            []string // Domains for deployment
	IsLocalhost        bool     // Whether it is a local deployment
	DeploymentID       string   // Friendly ID from API
	HostKeyCallback    ssh.HostKeyCallback
}

// Run executes the deployment process (legacy mode using direct DB).
func Run(appName, hostName, useBuild string, hostKeyCallback ssh.HostKeyCallback) {
	d := &Deployer{
		AppName:         appName,
		HostName:        hostName,
		useBuild:        useBuild,
		HostKeyCallback: hostKeyCallback,
	}

	// Capture all logs during deployment
	log.SetOutput(io.MultiWriter(os.Stdout, &d.LogBuffer))

	var err error
	defer func() {
		if err != nil {
			log.Printf("❌ Deployment failed: %v", err)
			if d.History != nil {
				_ = database.UpdateDeploymentHistoryStatus(d.History.ID, models.DeploymentStatusFailed, d.LogBuffer.String())
			}
			os.Exit(1)
		}
	}()

	err = d.setup()
	if err != nil {
		return
	}

	// Sync domain config to database
	if err = d.SyncDomainsForDeployment(); err != nil {
		log.Printf("⚠️  Warning: Failed to sync domain config: %v", err)
		// Do not abort deployment, only warn
	}

	err = d.execute()
	if err != nil {
		return
	}
}

// RunWithAPIClient executes the deployment using the API Client (Client-Server mode).
func RunWithAPIClient(apiClient client.APIClient, appName, hostName, useBuild string, hostKeyCallback ssh.HostKeyCallback) {
	// Check if deploying to localhost/server
	isLocalhost := hostName == "localhost" || hostName == "127.0.0.1" || hostName == "local"

	d := &Deployer{
		AppName:         appName,
		HostName:        hostName,
		useBuild:        useBuild,
		APIClient:       apiClient,
		IsLocalhost:     isLocalhost,
		HostKeyCallback: hostKeyCallback,
	}

	// Capture logs
	log.SetOutput(io.MultiWriter(os.Stdout, &d.LogBuffer))

	var err error
	// Defer error handling and status update via API
	defer func() {
		if err != nil {
			log.Printf("❌ Deployment failed: %v", err)
			// Update status via API if we have a deployment ID
			if d.DeploymentID != "" {
				_ = apiClient.UpdateDeploymentStatus(d.DeploymentID, "failed", 0, "", "")
				_ = apiClient.UploadDeploymentLogs(d.DeploymentID, d.LogBuffer.String())
			}
			os.Exit(1)
		}
	}()

	log.Println("---", "1. [CLI] Fetching remote config", "---")
	// Fetch config from API
	conf, err := apiClient.GetDeployConfig(appName, hostName)
	if err != nil {
		return // err set
	}

	d.Domains = conf.Domains // Store domains for later use

	log.Printf("Config fetched: App=%s, Host=%s", conf.App.Name, conf.Host.Name)

	// Convert DTOs to internal models
	d.Application, err = convertAppDTOToModel(&conf.App)
	if err != nil {
		return
	}
	d.Host, err = convertHostDTOToModel(&conf.Host)
	if err != nil {
		return
	}
	d.Instance, err = convertInstanceDTOToModel(&conf.Instance)
	if err != nil {
		return
	}

	// Set runtime: prioritize shipyard.toml, otherwise auto-detect
	config.LoadConfig(d.AppName, config.ConfigPath)
	d.Runtime = config.AppConfig.Runtime
	if d.Runtime == "" {
		d.Runtime = d.detectRuntime()
	}

	// Display domain info
	domains := config.AppConfig.Domains
	log.Printf("App: %s, Host: %s, Domains: %v, runtime=%s", d.Application.Name, d.Host.Name, domains, d.Runtime)

	// Check for missing domains and warn user
	if len(domains) == 0 && len(d.Domains) == 0 {
		log.Println("⚠️  Warning: No domain configuration detected!")
		log.Println("    Application will not be accessible via domain after deployment, and Caddy reverse proxy will not be configured.")
		log.Println("    Please configure 'domains' in 'shipyard.toml', e.g.:")
		log.Println("      domains = [\"example.com\", \"www.example.com\"]")
		log.Println("    Or add domain using CLI:")
		log.Printf("      deployer-cli domain add --app %s --domain example.com", d.AppName)
		// We don't abort, as user might want to deploy purely for testing or internal port usage
	}

	// Check if this is a server-side deployment (localhost = server machine)
	if isLocalhost {
		log.Println("---", "2. [Server-Side Deployment Mode] Deploying to server machine", "---")
		log.Println("📦 This will deploy the application to the server machine itself (not via SSH)")
		err = d.executeServerSideDeployment(apiClient, conf.Secrets)
		return
	}

	// Regular SSH-based deployment
	log.Println("---", "2. Connecting to remote host", "---")
	if err = d.connectSSHWithAPIConfig(); err != nil {
		return
	}

	// After SSH connection, prepare Caddy environment
	log.Println("---", "2.5. Preparing Caddy environment", "---")
	d.caddySvc = caddy.NewService(d.SSHClient)

	// Check Caddy availability immediately
	if err = d.caddySvc.CheckAvailability(); err != nil {
		return
	}

	// Sync domains from config
	if err = d.SyncDomainsForDeployment(); err != nil {
		log.Printf("⚠️  Warning: Failed to sync domain config: %v", err)
		// Don't abort deployment, just warn
	} else {
		// If sync success, update d.Domains from local config because that's what we just synced
		// This ensures we use the latest domains for traffic switching
		if len(config.AppConfig.Domains) > 0 {
			d.Domains = config.AppConfig.Domains
		}
	}

	// Execute deployment with API client (use the same execute logic as legacy mode)
	err = d.executeWithAPIClient(apiClient, conf.Secrets)
}

// connectSSHWithAPIConfig establishes an SSH connection using API-provided host config.
func (d *Deployer) connectSSHWithAPIConfig() error {
	sshConfig, err := sshutil.NewClientConfig(d.Host, d.HostKeyCallback)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", d.Host.Addr, d.Host.Port)
	d.SSHClient, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to remote host: %w", err)
	}
	log.Println("✅ Successfully connected to remote host.")
	return nil
}

// executeWithAPIClient executes the deployment using API-provided secrets
func (d *Deployer) executeWithAPIClient(apiClient client.APIClient, secrets map[string]string) error {
	defer d.SSHClient.Close()
	var err error

	// --- 3. Process build artifact (New build or reuse) ---
	if err = d.ProcessArtifact(); err != nil {
		return err
	}

	log.Println("🚀 Preparing release...")
	releasePath := fmt.Sprintf("%s/%s-%d", config.GetRemoteReleasesDir(), d.Version, time.Now().Unix())
	d.CurrentReleasePath = releasePath
	if err := d.executeRemoteCommand(fmt.Sprintf("mkdir -p %s", releasePath), false); err != nil {
		return err
	}

	// Create deployment history via API
	deployReq := &types.CreateDeploymentRequest{
		AppName:  d.AppName,
		HostName: d.HostName,
		Version:  d.Version,
	}
	historyDTO, err := apiClient.CreateDeployment(deployReq)
	if err != nil {
		return fmt.Errorf("failed to create deployment record: %w", err)
	}
	// Store the friendly ID for API calls
	d.DeploymentID = historyDTO.ID

	log.Println("📤 Uploading files...")
	if err := d.uploadTarFile(d.tarballPath, releasePath); err != nil {
		return err
	}

	// Set executable permissions for non-static deployments
	if err := d.ensurePermissions(releasePath); err != nil {
		return err
	}

	// --- 5b. Execute pre_deploy hooks ---
	if err := d.runHooks("pre_deploy", config.AppConfig.Hooks.PreDeploy); err != nil {
		return fmt.Errorf("pre_deploy failed: %w", err)
	}

	// --- 5c. Check path-related variables in .env file ---
	// d.checkAndWarnForPathsInEnvFile(releasePath) // Silence check

	log.Println("🔧 Configuring environment...")

	// 4. Merge all environment variables
	envs := d.prepareEnvVars(secrets)

	// For Phoenix apps, ensure SECRET_KEY_BASE exists
	if d.Runtime == "phoenix" {
		if _, exists := envs["SECRET_KEY_BASE"]; !exists {
			newSecret, err := crypto.GeneratePhoenixSecret()
			if err != nil {
				return fmt.Errorf("failed to generate secret: %w", err)
			}
			envs["SECRET_KEY_BASE"] = newSecret
			// Note: In API client mode, we don't persist the generated secret back to the server here?
			// Ideally the API should have provided it, or we should send it back.
			// For now, we keep the behavior of generating it ephemerally for this deployment if missing.
		}
	}

	// 5. Write final environment variables to remote server
	if err := d.injectEnvVars(envs); err != nil {
		return err
	}

	// 5.5 Ensure paths in environment variables have proper permissions
	if err := d.ensurePathPermissions(envs); err != nil {
		return fmt.Errorf("failed to ensure path permissions: %w", err)
	}

	if err := d.runHooks("migrate", config.AppConfig.Hooks.Migrate); err != nil {
		return fmt.Errorf("migrate failed: %w", err)
	}

	// 8. Start new version
	run, err := d.startNewVersion(releasePath)
	if err != nil {
		return err
	}
	greenPort := run.Port

	time.Sleep(2 * time.Second)

	log.Println("💓 Health check...")
	if err := d.performHealthCheck(greenPort); err != nil {
		// Rollback logic (stop new version)
		instancesDir := fmt.Sprintf("/var/www/%s/instances", d.AppName)
		d.executeRemoteCommand(fmt.Sprintf("systemctl stop %s@%d || true", d.AppName, greenPort), false)
		d.executeRemoteCommand(fmt.Sprintf("rm -f %s/%d || true", instancesDir, greenPort), false)

		// Update failed status via API
		_ = apiClient.UpdateDeploymentStatus(d.DeploymentID, "failed", 0, "", "")
		return fmt.Errorf("health check failed: %w", err)
	}

	// 10. Switch traffic
	if err := d.switchTraffic(greenPort, d.Domains); err != nil {
		return err
	}

	// Update active status (via API if possible, or implicitly done by switch traffic success)
	_ = apiClient.UpdateDeploymentStatus(d.DeploymentID, "success", greenPort, releasePath, d.GitCommitSHA)

	// --- 10b. Execute post_deploy hooks ---
	d.runHooks("post_deploy", config.AppConfig.Hooks.PostDeploy)

	oldPort := 0
	if d.Instance.ActivePort.Valid && d.Instance.ActivePort.Int64 > 0 {
		oldPort = int(d.Instance.ActivePort.Int64)
	}

	if oldPort > 0 {
		log.Printf("🛑 Stopping old version (:%d)...", oldPort)
		time.Sleep(3 * time.Second)
		d.executeRemoteCommand(fmt.Sprintf("systemctl disable %s@%d", d.AppName, oldPort), false)
		d.executeRemoteCommand(fmt.Sprintf("systemctl stop %s@%d", d.AppName, oldPort), false)
	}

	log.Println("🎉 Deployment successful!")

	// Update deployment status via API
	if err := apiClient.UpdateDeploymentStatus(d.DeploymentID, "success", greenPort, releasePath, d.GitCommitSHA); err != nil {
		// Log quietly
	}
	if err := apiClient.UploadDeploymentLogs(d.DeploymentID, d.LogBuffer.String()); err != nil {
		// Log quietly
	}

	return nil
}

// execute executes the core logic of deployment
func (d *Deployer) execute() error {
	defer d.SSHClient.Close()
	var err error

	// --- 3. Process build artifact (New build or reuse) ---
	if err = d.ProcessArtifact(); err != nil {
		return err
	}

	log.Println("---", "4. Preparing remote environment", "---")
	releasePath := fmt.Sprintf("%s/%s-%d", config.GetRemoteReleasesDir(), d.Version, time.Now().Unix())
	d.CurrentReleasePath = releasePath // Store for hook variable substitution
	if err := d.executeRemoteCommand(fmt.Sprintf("mkdir -p %s", releasePath), false); err != nil {
		return err
	}

	d.History, err = database.CreateDeploymentHistory(d.Instance.ID, d.Version, releasePath)
	if err != nil {
		return fmt.Errorf("failed to create deployment history record: %w", err)
	}

	log.Println("---", "5. Upload and extract files (streaming)", "---")
	// Call new function to complete upload, extract and progress display in one step
	if err := d.uploadTarFile(d.tarballPath, releasePath); err != nil {
		// If error occurs, function internal has contained all error info (e.g., "failed to execute remote streaming extraction: ...")
		// Your "green" environment directory (releasePath) on remote might be incomplete
		return err
	}

	// Only set executable permissions for non-static deployments (Phoenix/Node/Golang)
	if err := d.ensurePermissions(releasePath); err != nil {
		return err
	}

	// --- 5b. Execute pre_deploy hooks ---
	if err := d.runHooks("pre_deploy", config.AppConfig.Hooks.PreDeploy); err != nil {
		return fmt.Errorf("pre_deploy hook execution failed: %w", err)
	}

	// --- 5c. Check path-related variables in .env file ---
	d.checkAndWarnForPathsInEnvFile(releasePath)

	log.Println("---", "6. Inject environment variables (env and secrets)", "---")

	// 1. First get all existing secrets
	secrets, err := database.GetSecretsForApp(d.Application.ID)
	if err != nil {
		return fmt.Errorf("failed to get application secrets: %w", err)
	}

	// 2. If it is a Phoenix app, ensure SECRET_KEY_BASE exists
	if d.Runtime == "phoenix" {
		if _, exists := secrets["SECRET_KEY_BASE"]; !exists {
			log.Println("⚠️ SECRET_KEY_BASE not found, generating one for you...")
			newSecret, err := crypto.GeneratePhoenixSecret()
			if err != nil {
				return fmt.Errorf("failed to generate SECRET_KEY_BASE: %w", err)
			}
			if err := database.SetSecret(d.Application.ID, "SECRET_KEY_BASE", newSecret); err != nil {
				return fmt.Errorf("failed to save new SECRET_KEY_BASE: %w", err)
			}
			log.Println("✅ New SECRET_KEY_BASE generated and saved.")
			// 3. Add newly generated secret to in-memory map for use in current deployment
			secrets["SECRET_KEY_BASE"] = newSecret
		}
	}

	// 4. Merge all environment variables
	envs := d.prepareEnvVars(secrets)

	// 5. Write final environment variables to remote server
	if err := d.injectEnvVars(envs); err != nil {
		return err
	}

	log.Println("---", "6.5. Ensure path permissions", "---")
	// Ensure paths in environment variables have proper permissions
	if err := d.ensurePathPermissions(envs); err != nil {
		return fmt.Errorf("failed to ensure path permissions: %w", err)
	}

	log.Println("---", "7. Execute migrate hooks", "---")
	if err := d.runHooks("migrate", config.AppConfig.Hooks.Migrate); err != nil {
		// Migrate hooks are critical, abort on failure
		return fmt.Errorf("migrate hook execution failed: %w", err)
	}

	// 8. Start new version
	run, err := d.startNewVersion(releasePath)
	if err != nil {
		return err
	}
	// In Server mode, immediately record new instance to database
	if err := database.AddDeploymentInstance(run); err != nil {
		return fmt.Errorf("failed to record instance run info: %w", err)
	}
	greenPort := run.Port

	time.Sleep(3 * time.Second)

	log.Println("---", "9. Health check", "---")
	if err := d.performHealthCheck(greenPort); err != nil {
		// Rollback logic
		instancesDir := fmt.Sprintf("/var/www/%s/instances", d.AppName)
		d.executeRemoteCommand(fmt.Sprintf("systemctl stop %s@%d || true", d.AppName, greenPort), true)
		d.executeRemoteCommand(fmt.Sprintf("rm -f %s/%d || true", instancesDir, greenPort), false)

		st := time.Now()
		_ = database.UpdateDeploymentInstanceStatus(run.ID, "failed", &st)
		return fmt.Errorf("new version health check failed: %w", err)
	}
	log.Println("✅ New version health status is good")

	log.Println("---", "10. Switch traffic", "---")
	// Get all domains and update Caddy config
	domains, err := GetDomainsForDeploy(d.Instance.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain list: %w", err)
	}

	// Note: we no longer fallback to Application.Domain; rely on domains table or config only.
	if err := d.switchTraffic(greenPort, domains); err != nil {
		return err
	}

	_ = database.UpdateDeploymentInstanceStatus(run.ID, "active", nil)

	// --- 10b. Execute post_deploy hooks ---
	if err := d.runHooks("post_deploy", config.AppConfig.Hooks.PostDeploy); err != nil {
		// Post-deploy hook failure usually shouldn't cause deployment failure, just log warning
		log.Printf("⚠️ post_deploy hook failed: %v", err)
	}

	// 11. Handle old version
	if err := d.stopOldVersion(greenPort); err != nil {
		return err
	}

	log.Println("---", "12. Clean up stale instances", "---")
	if err := d.cleanupStaleInstances(greenPort); err != nil {
		log.Printf("⚠️ Error cleaning up stale instances: %v", err)
	}

	log.Println("🎉 Deployment successfully completed!")
	return database.UpdateDeploymentHistoryStatus(d.History.ID, models.DeploymentStatusSuccess, d.LogBuffer.String())
}

// ensurePermissions sets executable permissions for non-static deployments.
func (d *Deployer) ensurePermissions(releasePath string) error {
	if d.Runtime == "static" {
		return nil
	}

	log.Printf("General permission correction: append executable permission for bin/* and erts-*/bin/* (if needed)")

	cmds := []string{
		fmt.Sprintf("cd %s && if [ -d bin ]; then chmod +x bin/* || true; fi", releasePath),
		fmt.Sprintf("cd %s && if ls -d erts-*/bin >/dev/null 2>&1; then find erts-*/bin -type f -exec chmod +x {} +; fi", releasePath),
		fmt.Sprintf("cd %s && if [ -d releases ]; then find releases -name elixir -type f -exec chmod +x {} +; fi", releasePath),
	}

	for _, cmd := range cmds {
		if err := d.executeRemoteCommand(cmd, true); err != nil {
			return err
		}
	}

	return nil
}

// switchTraffic updates the Caddy reverse proxy and enables the new service.
func (d *Deployer) switchTraffic(port int, domains []string) error {
	log.Println("🔀 Switching traffic...")

	if len(domains) > 0 {
		if err := d.caddySvc.UpdateReverseProxyMultiDomain(domains, port); err != nil {
			return fmt.Errorf("failed to update Caddy config: %w", err)
		}
		log.Printf("✅ Caddy traffic switched to port %d (Domains: %v)", port, domains)
	} else {
		log.Println("⚠️  Warning: No domain configured, skipping Caddy config")
	}

	// Enable auto-start for new version
	enableCmd := fmt.Sprintf("systemctl enable %s@%d", d.AppName, port)
	if err := d.executeRemoteCommand(enableCmd, true); err != nil {
		log.Printf("⚠️ Warning: Failed to set new version auto-start: %v", err)
		// Note: Usually this shouldn't block successful deployment, so just log warning
	} else {
		log.Printf("✅ Enabled auto-start for new version (Port %d)", port)
	}

	return nil
}

// executeServerSideDeployment handles server-side deployment (localhost = server machine)
// In this mode, CLI builds and uploads artifact to server, then server executes deployment locally
func (d *Deployer) executeServerSideDeployment(apiClient client.APIClient, secrets map[string]string) error {
	var err error

	// --- 3. Process build artifact (New build or reuse) ---
	log.Println("---", "3. [CLI] Building or reusing artifact", "---")
	if err = d.ProcessArtifact(); err != nil {
		return err
	}

	log.Printf("📦 Artifact ready: %s (version: %s)", d.tarballPath, d.Version)

	// --- 4. Create deployment record ---
	log.Println("---", "4. [CLI] Creating deployment record", "---")
	deployReq := &types.CreateDeploymentRequest{
		AppName:  d.AppName,
		HostName: d.HostName,
		Version:  d.Version,
	}
	historyDTO, err := apiClient.CreateDeployment(deployReq)
	if err != nil {
		return fmt.Errorf("failed to create deployment record: %w", err)
	}
	d.DeploymentID = historyDTO.ID
	log.Printf("✅ Deployment record created: %s", d.DeploymentID)

	// --- 5. Upload artifact to server ---
	log.Println("---", "5. [CLI] Uploading artifact to server", "---")
	if err := apiClient.UploadDeploymentArtifact(d.DeploymentID, d.tarballPath); err != nil {
		return fmt.Errorf("failed to upload artifact to server: %w", err)
	}
	log.Println("✅ Artifact uploaded to server successfully")

	// --- 6. Trigger server-side execution ---
	log.Println("---", "6. [CLI] Triggering server-side deployment execution", "---")
	if err := apiClient.ExecuteServerDeployment(d.DeploymentID, d.Version, d.GitCommitSHA, d.md5Hash); err != nil {
		return fmt.Errorf("failed to trigger server-side deployment: %w", err)
	}
	log.Println("✅ Server-side deployment triggered successfully")
	log.Println("📝 Note: Deployment is now running on the server. Check deployment logs for progress.")

	// Upload current logs
	if err := apiClient.UploadDeploymentLogs(d.DeploymentID, d.LogBuffer.String()); err != nil {
		// Log quietly
		log.Printf("⚠️  Warning: Failed to upload logs: %v", err)
	}

	return nil
}
