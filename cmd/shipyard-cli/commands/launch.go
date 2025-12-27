package commands

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"youfun/shipyard/cmd/utils"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/cliutils"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/deploy"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/pkg/types"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

// launchCommand handles the 'launch' command for the CLI
func LaunchCommand(apiClient *client.Client) {
	cmd := flag.NewFlagSet("launch", flag.ExitOnError)
	appNameFlag := cmd.String("app", "", "Application name (optional, defaults to shipyard.toml or current directory name)")
	hostNameFlag := cmd.String("host", "", "Host to deploy to (optional, defaults to interactive selection)")
	cmd.Parse(os.Args[2:])

	log.Println("--- 🚀 Starting shipyard launch process (CLI Mode) ---")

	// 1. Handle shipyard.toml file (Local)
	_, appName := handleShipyardToml(*appNameFlag)

	// 2. Check and add application (API)
	if err := apiClient.CreateApp(appName); err != nil {
		// Ignore error if it's just "already exists", but print for now
		log.Printf("Attempting to register app '%s' (ignoring if already exists): %v", appName, err)
	} else {
		log.Printf("✅ App '%s' registered.", appName)
	}

	// 3. Check available hosts and select (API + Interactive)
	hostDTO := selectHost(apiClient, *hostNameFlag, "")

	// 4. Check and link app instance (API)
	log.Printf("Linking app '%s' to host '%s'...", appName, hostDTO.Name)
	if err := apiClient.LinkApp(appName, hostDTO.Name); err != nil {
		// Might already be linked, continue but warn
		log.Printf("Link request returned: %v (might be already linked)", err)
	} else {
		log.Println("✅ App instance linked successfully.")
	}

	// 5. Initialize remote host (SSH via Local)
	log.Println("--- ⚙️ Initializing remote host ---")
	runtime := detectRuntime()

	// For Phoenix projects, ensure 'mix phx.gen.release' has been run
	// Note: This only applies to Phoenix projects, not pure Elixir projects
	if runtime == "phoenix" {
		releaseExPath := filepath.Join("lib", appName, "release.ex")
		_, relDirErr := os.Stat("rel")
		_, releaseExFileErr := os.Stat(releaseExPath)

		if os.IsNotExist(relDirErr) || os.IsNotExist(releaseExFileErr) {
			log.Println("Detected Phoenix project missing release config or Release module, running 'mix phx.gen.release'...")
			mixCmd := exec.Command("mix", "phx.gen.release")
			mixCmd.Stdout = os.Stdout
			mixCmd.Stderr = os.Stderr
			if err := mixCmd.Run(); err != nil {
				log.Fatalf("❌ 'mix phx.gen.release' failed: %v", err)
			}
			log.Println("✅ 'mix phx.gen.release' completed.")
		}
	} else if runtime == "elixir" {
		log.Println("Detected pure Elixir project (no Phoenix), skipping 'mix phx.gen.release'.")
	}

	modelHost := &models.SSHHost{
		ID:         parseUUID(hostDTO.ID),
		Name:       hostDTO.Name,
		Addr:       hostDTO.Addr,
		Port:       hostDTO.Port,
		User:       hostDTO.User,
		Password:   hostDTO.Password,
		PrivateKey: hostDTO.PrivateKey,
	}

	// Use InsecureIgnoreHostKey for now in CLI mode as we don't have DB access
	if err := deploy.InitializeHost(modelHost, appName, runtime, "", "phoenix", ssh.InsecureIgnoreHostKey()); err != nil {
		log.Fatalf("Failed to initialize remote host: %v", err)
	}
	log.Println("✅ Remote host initialization completed.")

	// 6. Execute deployment
	log.Println("--- 🚀 Executing first deployment ---")
	deploy.RunWithAPIClient(apiClient, appName, hostDTO.Name, "", ssh.InsecureIgnoreHostKey())
	log.Println("✅ Application deployed successfully!")
	log.Println("--- 🎉 shipyard launch process completed ---")
}

func handleShipyardToml(appNameFlag string) (config.Config, string) {
	return cliutils.HandleShipyardToml(appNameFlag)
}

func selectHost(apiClient *client.Client, hostNameFlag string, defaultHostName string) *types.SSHHostDTO {
	hosts, err := apiClient.ListHosts()
	if err != nil {
		log.Fatalf("Failed to get host list: %v", err)
	}

	// TODO: Re-enable localhost deployment after fixing database instance creation issue
	// Special handling for localhost flag
	// if hostNameFlag == "localhost" || hostNameFlag == "local" {
	// 	return createLocalhostHost()
	// }

	if hostNameFlag != "" {
		for _, h := range hosts {
			if h.Name == hostNameFlag {
				log.Printf("Using specified host: '%s'", h.Name)
				return &h
			}
		}
		log.Fatalf("Error: Specified host '%s' not found", hostNameFlag)
	}

	// TODO: Re-enable localhost deployment after fixing database instance creation issue
	// Add localhost option to the list
	var items []string
	// items = append(items, utils.GetLocalhostInfo())

	defaultIndex := -1

	for i, host := range hosts {
		item := fmt.Sprintf("%s (%s:%d)", host.Name, host.Addr, host.Port)
		if defaultHostName != "" && host.Name == defaultHostName {
			item += " (last deployed)"
			defaultIndex = i // No +1 because localhost option is removed
		}
		items = append(items, item)
	}

	promptMsg := "\n--- Please select a deployment host ---"
	if defaultIndex != -1 {
		promptMsg += fmt.Sprintf(" (Last deployed: %d)", defaultIndex+1) // Display index is 1-based
	}

	selectedIndex, err := utils.PromptForSelection(promptMsg, items, defaultIndex)
	if err != nil {
		log.Fatalf("Error selecting host: %v", err)
	}

	// TODO: Re-enable localhost deployment after fixing database instance creation issue
	// If localhost is selected (index 0)
	// if selectedIndex == 0 {
	// 	return createLocalhostHost()
	// }

	// No need to adjust index since localhost option is removed
	selectedHost := &hosts[selectedIndex]
	log.Printf("Selected host: '%s'", selectedHost.Name)
	return selectedHost
}

// createLocalhostHost creates a special SSHHostDTO for localhost deployment
func createLocalhostHost() *types.SSHHostDTO {
	log.Println("📍 Selected Server Machine (localhost)")
	log.Println("ℹ️  Note: This will deploy the application to the server where shipyard-server is running")
	log.Println("         The CLI will build the app locally, upload it to the server, and the server will execute deployment")

	// Check Caddy environment
	if err := utils.EnsureCaddyRunning(); err != nil {
		log.Fatalf("❌ Local environment check failed: %v", err)
	}

	return &types.SSHHostDTO{
		Name: "localhost",
		Addr: "127.0.0.1",
		Port: 22, // Not actually used for localhost
		User: "", // Not used for localhost
	}
}

func detectRuntime() string {
	return cliutils.DetectRuntime()
}

func parseUUID(idStr string) uuid.UUID {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil
	}
	return id
}
