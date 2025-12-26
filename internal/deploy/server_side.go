package deploy

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/database"
	"time"

	"github.com/google/uuid"
)

// ExecuteServerSideDeployment executes deployment on the server itself (no SSH)
// This is called by the API handler when localhost deployment is triggered
func ExecuteServerSideDeployment(deploymentIDStr, appName, version string) error {
	log.Printf("🚀 [Server] Starting server-side deployment for %s (deployment: %s, version: %s)", appName, deploymentIDStr, version)

	// Parse deployment ID
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid deployment ID: %w", err)
	}

	// Get deployment history
	history, err := database.GetDeploymentHistoryByID(deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get deployment history: %w", err)
	}

	// Get instance
	instance, err := database.GetApplicationInstanceByID(history.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to get application instance: %w", err)
	}

	// Get application
	app, err := database.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	// Verify artifact exists
	artifactPath := fmt.Sprintf("/var/lib/shipyard/artifacts/%s.tar.gz", deploymentIDStr)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		return fmt.Errorf("artifact file not found: %s", artifactPath)
	}

	// Ensure phoenix user exists
	if err := ensureServerUser("phoenix"); err != nil {
		log.Printf("⚠️  Warning: Failed to ensure 'phoenix' user exists: %v. Deployment might fail if permissions are incorrect.", err)
	}

	log.Printf("📦 [Server] Artifact found: %s", artifactPath)

	// Load app config
	config.LoadConfig(appName, config.ConfigPath)

	// Prepare release directory
	releasePath := fmt.Sprintf("%s/%s-%d", config.GetRemoteReleasesDir(), version, time.Now().Unix())
	log.Printf("📂 [Server] Creating release directory: %s", releasePath)
	if err := os.MkdirAll(releasePath, 0755); err != nil {
		return fmt.Errorf("failed to create release directory: %w", err)
	}

	// Extract artifact to release path
	log.Printf("📤 [Server] Extracting artifact to %s", releasePath)
	if err := extractTarGz(artifactPath, releasePath); err != nil {
		return fmt.Errorf("failed to extract artifact: %w", err)
	}

	// Set permissions
	log.Printf("🔐 [Server] Setting permissions")
	if err := setExecutablePermissions(releasePath, app.Name); err != nil {
		log.Printf("⚠️  Warning: Failed to set permissions: %v", err)
	}

	// Execute pre_deploy hooks (if any)
	if len(config.AppConfig.Hooks.PreDeploy) > 0 {
		log.Printf("🪝 [Server] Executing pre_deploy hooks")
		if err := executeLocalHooks(releasePath, "pre_deploy", config.AppConfig.Hooks.PreDeploy); err != nil {
			return fmt.Errorf("pre_deploy hook failed: %w", err)
		}
	}

	// Inject environment variables
	log.Printf("🔧 [Server] Injecting environment variables")
	secrets, err := database.GetAllSecrets(app.ID)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to get secrets: %v", err)
		secrets = make(map[string]string)
	}
	if err := injectEnvVarsLocally(releasePath, secrets); err != nil {
		return fmt.Errorf("failed to inject environment variables: %w", err)
	}

	// Start new version
	log.Printf("🌱 [Server] Starting new version")
	port, err := findFreePortLocally()
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}
	log.Printf("Found free port: %d", port)

	if err := startLocalInstance(app.Name, port, releasePath); err != nil {
		return fmt.Errorf("failed to start new version: %w", err)
	}

	// Update deployment history status
	if err := database.RecordSuccessfulDeployment(deploymentID, port, releasePath, ""); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	// Switch traffic via Caddy
	log.Printf("🔄 [Server] Switching traffic to port %d", port)
	caddySvc := caddy.NewLocalService()
	
	// Get domains from config
	domains := config.AppConfig.Domains
	if len(domains) == 0 {
		log.Println("⚠️  Warning: No domains configured in shipyard.toml")
	}

	if len(domains) > 0 {
		if err := caddySvc.UpdateReverseProxyMultiDomain(domains, port); err != nil {
			log.Printf("⚠️  Warning: Failed to update Caddy routes: %v", err)
		} else {
			log.Printf("✅ [Server] Updated Caddy routes for %d domains to port %d", len(domains), port)
		}
	} else {
		log.Println("⚠️  Warning: No domains configured, skipping traffic switching")
	}

	// Handle old version cleanup
	if instance.ActivePort.Valid && instance.ActivePort.Int64 > 0 {
		oldPort := int(instance.ActivePort.Int64)
		log.Printf("🔄 [Server] Stopping old version on port %d", oldPort)
		if err := stopLocalInstance(app.Name, oldPort); err != nil {
			log.Printf("⚠️  Warning: Failed to stop old version: %v", err)
		}
	}

	log.Printf("✅ [Server] Server-side deployment completed successfully")
	return nil
}

// extractTarGz extracts a tar.gz file to the target directory
func extractTarGz(tarPath, targetDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

// setExecutablePermissions sets execute permissions for binaries
func setExecutablePermissions(releasePath, appName string) error {
	// For Phoenix/Elixir apps
	binPath := filepath.Join(releasePath, "bin", appName)
	if _, err := os.Stat(binPath); err == nil {
		if err := os.Chmod(binPath, 0755); err != nil {
			return err
		}
		log.Printf("✅ Set executable permission: %s", binPath)
	}

	// For compiled binaries in root
	if _, err := os.Stat(filepath.Join(releasePath, appName)); err == nil {
		if err := os.Chmod(filepath.Join(releasePath, appName), 0755); err != nil {
			return err
		}
	}

	return nil
}

// executeLocalHooks executes hooks locally
func executeLocalHooks(releasePath, hookName string, hooks []config.Hook) error {
	for i, hook := range hooks {
		cmdStr := hook.Command
		log.Printf("🪝 Executing %s hook [%d/%d]: %s", hookName, i+1, len(hooks), cmdStr)
		cmd := exec.Command("bash", "-c", cmdStr)
		cmd.Dir = releasePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("hook command failed: %w", err)
		}
	}
	return nil
}

// injectEnvVarsLocally injects environment variables into .env file
func injectEnvVarsLocally(releasePath string, secrets map[string]string) error {
	// Similar to the SSH version but local file operations
	envPath := filepath.Join(releasePath, ".env")
	
	// Read existing .env if it exists
	existing := make(map[string]string)
	if content, err := os.ReadFile(envPath); err == nil {
		// Parse existing vars
		for _, line := range splitLines(string(content)) {
			if len(line) > 0 && line[0] != '#' {
				parts := splitOnce(line, "=")
				if len(parts) == 2 {
					existing[parts[0]] = parts[1]
				}
			}
		}
	}

	// Merge with secrets
	for k, v := range secrets {
		existing[k] = v
	}

	// Write back
	f, err := os.Create(envPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range existing {
		fmt.Fprintf(f, "%s=%s\n", k, v)
	}

	return nil
}

// findFreePortLocally finds a free port on the local machine
func findFreePortLocally() (int, error) {
	// Simple implementation - check systemd services
	basePort := 4000
	for port := basePort; port < basePort+100; port++ {
		// Check if port is in use by checking systemd service status
		cmd := exec.Command("systemctl", "is-active", fmt.Sprintf("*@%d", port))
		if err := cmd.Run(); err != nil {
			// Service not active, port is free
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free ports available in range %d-%d", basePort, basePort+99)
}

// startLocalInstance starts a local systemd service instance
func startLocalInstance(appName string, port int, releasePath string) error {
	instancesDir := fmt.Sprintf("/var/www/%s/instances", appName)
	
	// Create instances directory
	if err := os.MkdirAll(instancesDir, 0755); err != nil {
		return err
	}

	// Create symlink
	symlinkPath := filepath.Join(instancesDir, fmt.Sprintf("%d", port))
	os.Remove(symlinkPath) // Remove if exists
	if err := os.Symlink(releasePath, symlinkPath); err != nil {
		return err
	}

	// Fix ownership (optional, may require sudo)
	cmd := exec.Command("chown", "-R", "phoenix:phoenix", releasePath)
	if err := cmd.Run(); err != nil {
		log.Printf("⚠️  Warning: Failed to change ownership: %v", err)
	}

	// Start systemd service
	serviceName := fmt.Sprintf("%s@%d", appName, port)
	cmd = exec.Command("systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}

	log.Printf("✅ Started service: %s", serviceName)

	// Enable auto-start
	cmd = exec.Command("systemctl", "enable", serviceName)
	if err := cmd.Run(); err != nil {
		log.Printf("⚠️  Warning: Failed to enable auto-start: %v", err)
	}

	return nil
}

// stopLocalInstance stops a local systemd service instance
func stopLocalInstance(appName string, port int) error {
	serviceName := fmt.Sprintf("%s@%d", appName, port)
	cmd := exec.Command("systemctl", "stop", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}
	log.Printf("✅ Stopped service: %s", serviceName)
	return nil
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	var line string
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, line)
			line = ""
		} else if c != '\r' {
			line += string(c)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func splitOnce(s, sep string) []string {
	idx := 0
	for i := range s {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			idx = i
			return []string{s[:idx], s[idx+len(sep):]}
		}
	}
	return []string{s}
}

// ensureServerUser checks if a user exists and creates it if not
func ensureServerUser(username string) error {
	// Check if user exists
	if err := exec.Command("id", "-u", username).Run(); err == nil {
		return nil // User exists
	}

	log.Printf("👤 [Server] User '%s' not found, creating...", username)

	// Create user with home directory and bash shell
	// useradd -m -s /bin/bash <username>
	cmd := exec.Command("useradd", "-m", "-s", "/bin/bash", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create user %s: %w (output: %s)", username, err, string(output))
	}

	log.Printf("✅ [Server] Created user '%s'", username)
	return nil
}
