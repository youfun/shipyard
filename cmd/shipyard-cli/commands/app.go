package commands

import (
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/cliutils"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

// Constants for health check configuration
const (
	cliHealthCheckMaxRetries    = 10
	cliHealthCheckRetryInterval = 2 * time.Second
)

// appCommand handles the 'app' command with subcommands: restart, stop, status
func AppCommand(apiClient *client.Client) {
	if len(os.Args) < 3 {
		printAppUsage()
		os.Exit(1)
	}

	subCommand := os.Args[2]
	switch subCommand {
	case "restart":
		appRestartCommand(apiClient)
	case "stop":
		appStopCommand(apiClient)
	case "status":
		appStatusCommand(apiClient)
	case "help", "--help", "-h":
		printAppUsage()
	default:
		fmt.Printf("Unknown app subcommand: %s\n", subCommand)
		printAppUsage()
		os.Exit(1)
	}
}

func printAppUsage() {
	fmt.Print(`
Usage: shipyard-cli app <subcommand> [options]

Subcommands:
  restart     Restart application (stop then start)
  stop        Stop application
  status      View application status

Options:
  --app       Application name (optional, defaults to shipyard.toml)
  --host      Host name (optional, defaults to interactive selection)

Example:
  shipyard-cli app restart
  shipyard-cli app restart --app my-app --host prod
  shipyard-cli app stop --app my-app
  shipyard-cli app status
`)
}

// resolveAppAndHostFromAPI resolves the app name and host using API
func resolveAppAndHostFromAPI(apiClient *client.Client, appFlag, hostFlag string) (appName, hostName string, instanceInfo *client.InstanceInfo, host *models.SSHHost, err error) {
	// Resolve app name
	if appFlag != "" {
		appName = appFlag
	} else {
		var projConf config.Config
		if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err == nil {
			appName = projConf.App
		}
		if appName == "" {
			return "", "", nil, nil, fmt.Errorf("could not determine application name. Please use --app or run in project directory (shipyard.toml required)")
		}
	}

	// Resolve host name
	if hostFlag != "" {
		hostName = hostFlag
	} else {
		// Interactive host selection via API
		hosts, listErr := apiClient.ListHosts()
		if listErr != nil {
			return "", "", nil, nil, fmt.Errorf("failed to list hosts: %w", listErr)
		}
		if len(hosts) == 0 {
			return "", "", nil, nil, fmt.Errorf("no available hosts on server")
		}
		if len(hosts) == 1 {
			hostName = hosts[0].Name
			log.Printf("Automatically selected unique host: %s", hostName)
		} else {
			var items []string
			for _, h := range hosts {
				items = append(items, fmt.Sprintf("%s (%s:%d)", h.Name, h.Addr, h.Port))
			}
			selectedIndex, selectErr := cliutils.PromptForSelection("\n--- Please select a host ---", items, -1)
			if selectErr != nil {
				return "", "", nil, nil, fmt.Errorf("error selecting host: %w", selectErr)
			}
			hostName = hosts[selectedIndex].Name
		}
	}

	// Get instance info from API
	instanceInfo, err = apiClient.GetInstance(appName, hostName)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to get instance info: %w", err)
	}

	// Convert to models.SSHHost for SSH operations
	host = &models.SSHHost{
		Name: instanceInfo.Host.Name,
		Addr: instanceInfo.Host.Addr,
		Port: instanceInfo.Host.Port,
		User: instanceInfo.Host.User,
	}
	if instanceInfo.Host.Password != nil {
		host.Password = instanceInfo.Host.Password
	}
	if instanceInfo.Host.PrivateKey != nil {
		host.PrivateKey = instanceInfo.Host.PrivateKey
	}

	return appName, hostName, instanceInfo, host, nil
}

// appRestartCommand handles the 'app restart' command
func appRestartCommand(apiClient *client.Client) {
	cmd := flag.NewFlagSet("app restart", flag.ExitOnError)
	appFlag := cmd.String("app", "", "Application name (optional)")
	hostFlag := cmd.String("host", "", "Host name (optional)")
	cmd.Parse(os.Args[3:])

	appName, hostName, instanceInfo, host, err := resolveAppAndHostFromAPI(apiClient, *appFlag, *hostFlag)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	log.Printf("--- Restarting app '%s' (Host: %s) ---", appName, hostName)

	// Check if there's an active port
	if instanceInfo.Instance.ActivePort == 0 {
		log.Fatalf("❌ Cannot restart: App has no active instance (active_port not set). Please deploy first.")
	}

	activePort := int(instanceInfo.Instance.ActivePort)
	log.Printf("Current active port: %d", activePort)

	// Connect to remote host
	sshClient, err := connectToHostCLI(host)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer sshClient.Close()

	// Step 1: Stop the service
	log.Printf("--- Stopping service (Port %d) ---", activePort)
	stopCmd := fmt.Sprintf("systemctl stop %s@%d", appName, activePort)
	if _, err := executeRemoteCommandCLI(sshClient, stopCmd); err != nil {
		log.Fatalf("❌ Failed to stop service: %v", err)
	}
	log.Printf("✅ Service stopped")

	// Brief pause to ensure service is fully stopped
	time.Sleep(1 * time.Second)

	// Step 2: Start the service
	log.Printf("--- Starting service (Port %d) ---", activePort)
	startCmd := fmt.Sprintf("systemctl start %s@%d", appName, activePort)
	if _, err := executeRemoteCommandCLI(sshClient, startCmd); err != nil {
		log.Fatalf("❌ Failed to start service: %v", err)
	}

	// Step 3: Health check
	log.Println("--- Executing health check ---")
	if err := healthCheckWithRetryCLI(sshClient, appName, activePort, cliHealthCheckMaxRetries); err != nil {
		log.Printf("❌ Health check failed: %v", err)
		log.Println("⚠️ Service may not have started correctly, please check logs")
		os.Exit(1)
	}

	log.Printf("✅ App '%s' restarted successfully (Host: %s, Port: %d)", appName, hostName, activePort)
}

// appStopCommand handles the 'app stop' command
func appStopCommand(apiClient *client.Client) {
	cmd := flag.NewFlagSet("app stop", flag.ExitOnError)
	appFlag := cmd.String("app", "", "Application name (optional)")
	hostFlag := cmd.String("host", "", "Host name (optional)")
	cmd.Parse(os.Args[3:])

	appName, hostName, instanceInfo, host, err := resolveAppAndHostFromAPI(apiClient, *appFlag, *hostFlag)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	log.Printf("--- Stopping app '%s' (Host: %s) ---", appName, hostName)

	// Check if there's an active port
	if instanceInfo.Instance.ActivePort == 0 {
		log.Printf("⚠️ App has no active instance (active_port not set).")
		return
	}

	activePort := int(instanceInfo.Instance.ActivePort)
	log.Printf("Current active port: %d", activePort)

	// Connect to remote host
	sshClient, err := connectToHostCLI(host)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer sshClient.Close()

	// Stop the service
	stopCmd := fmt.Sprintf("systemctl stop %s@%d", appName, activePort)
	if _, err := executeRemoteCommandCLI(sshClient, stopCmd); err != nil {
		log.Fatalf("❌ Failed to stop service: %v", err)
	}

	log.Printf("✅ App '%s' stopped successfully (Host: %s, Port: %d)", appName, hostName, activePort)
}

// appStatusCommand handles the 'app status' command
func appStatusCommand(apiClient *client.Client) {
	cmd := flag.NewFlagSet("app status", flag.ExitOnError)
	appFlag := cmd.String("app", "", "Application name (optional)")
	hostFlag := cmd.String("host", "", "Host name (optional)")
	cmd.Parse(os.Args[3:])

	appName, hostName, instanceInfo, host, err := resolveAppAndHostFromAPI(apiClient, *appFlag, *hostFlag)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	fmt.Printf("\n--- App Status: %s ---\n", appName)
	fmt.Printf("Host: %s (%s:%d)\n", hostName, host.Addr, host.Port)

	// Check active port
	if instanceInfo.Instance.ActivePort == 0 {
		fmt.Println("Status: Not deployed (no active instance)")
		return
	}

	activePort := int(instanceInfo.Instance.ActivePort)
	fmt.Printf("Active Port: %d\n", activePort)

	// Connect to remote host to check systemd status
	sshClient, err := connectToHostCLI(host)
	if err != nil {
		fmt.Printf("⚠️ Failed to connect to host to check service status: %v\n", err)
		return
	}
	defer sshClient.Close()

	// Check systemd service status
	statusCmd := fmt.Sprintf("systemctl is-active %s@%d", appName, activePort)
	output, statusErr := executeRemoteCommandCLI(sshClient, statusCmd)
	output = strings.TrimSpace(output)

	if statusErr != nil {
		if strings.Contains(output, "inactive") {
			fmt.Printf("Service Status: ⏹️  Stopped (inactive)\n")
		} else if strings.Contains(output, "failed") {
			fmt.Printf("Service Status: ❌ Failed (failed)\n")
		} else {
			fmt.Printf("Service Status: ⚠️  Unknown (%s)\n", output)
		}
	} else {
		fmt.Printf("Service Status: ✅ Running (active)\n")
	}

	// Show previous active port if available (for rollback info)
	if instanceInfo.Instance.PreviousActivePort > 0 {
		prevPort := int(instanceInfo.Instance.PreviousActivePort)
		fmt.Printf("\nRollback Port: %d\n", prevPort)
	}

	fmt.Println()
}

// connectToHostCLI establishes an SSH connection to the host
func connectToHostCLI(host *models.SSHHost) (*ssh.Client, error) {
	sshConfig, err := sshutil.NewClientConfig(host, ssh.InsecureIgnoreHostKey())
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH config: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", host.Addr, host.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote host: %w", err)
	}
	log.Println("✅ Successfully connected to remote host.")
	return sshClient, nil
}

// executeRemoteCommandCLI executes a remote command and logs the output
func executeRemoteCommandCLI(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		log.Printf("Failed to create SSH session: %v", err)
		return "", err
	}
	defer session.Close()

	log.Printf("🚀 Executing remote command: %s", command)
	output, err := session.CombinedOutput(command)
	outputStr := strings.TrimSpace(string(output))

	if len(outputStr) > 0 {
		log.Println(outputStr)
	}

	if err != nil {
		return outputStr, err
	}
	return outputStr, nil
}

// healthCheckWithRetryCLI performs health check with retry
func healthCheckWithRetryCLI(client *ssh.Client, appName string, port int, maxRetries int) error {
	healthCheckCmd := fmt.Sprintf("systemctl is-active --quiet %s@%d", appName, port)
	for i := 0; i < maxRetries; i++ {
		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create SSH session: %w", err)
		}
		err = session.Run(healthCheckCmd)
		session.Close()

		if err == nil {
			log.Println("✅ Health check passed")
			return nil
		}
		log.Printf("Health check attempt %d/%d failed, retrying in %v...", i+1, maxRetries, cliHealthCheckRetryInterval)
		time.Sleep(cliHealthCheckRetryInterval)
	}

	return fmt.Errorf("service failed to enter 'active' state after multiple attempts")
}
