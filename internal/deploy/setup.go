package deploy

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

// setup prepares for the deployment by loading configuration, connecting to the host, and preparing Caddy.
func (d *Deployer) setup() error {
	log.Println("--- 1. Setting up deployment environment ---")
	var err error

	// If app name is not provided, read it from shipyard.toml
	if d.AppName == "" {
		var projConf struct {
			App string `toml:"app"`
		}
		if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err != nil {
			return fmt.Errorf("failed to read %s, please ensure the file exists or use the --app flag to specify the application", config.ConfigPath)
		}
		d.AppName = projConf.App
	}

	// Load and set configuration
	config.LoadConfig(d.AppName, config.ConfigPath)
	d.AppName = config.AppConfig.App

	// Get instance, application, and host details from the database
	d.Instance, d.Application, d.Host, err = database.GetInstance(d.AppName, d.HostName)
	if err != nil {
		return d.handleInstanceLoadingError(err)
	}

	// Set runtime: prioritize shipyard.toml, otherwise auto-detect
	d.Runtime = config.AppConfig.Runtime
	if d.Runtime == "" {
		d.Runtime = d.detectRuntime()
	}

	// Display domain information
	domains := config.AppConfig.Domains
	log.Printf("App: %s, Host: %s, Domains: %v, runtime=%s", d.Application.Name, d.Host.Name, domains, d.Runtime)

	// Check if the host is initialized and initialize if necessary
	if d.Host.InitializedAt.Time == nil {
		if err := d.initializeHostInteractively(); err != nil {
			return err
		}
	}

	// Skip SSH connection for localhost deployment
	if !d.IsLocalhost {
		log.Println("--- 2. Connecting to remote host ---")
		if err := d.connectSSH(); err != nil {
			return err
		}

		// After SSH connection, prepare Caddy environment
		log.Println("--- 2.5. Preparing Caddy environment ---")
		d.caddySvc = caddy.NewService(d.SSHClient)

		// Check Caddy availability immediately
		if err := d.caddySvc.CheckAvailability(); err != nil {
			return err
		}
	} else {
		log.Println("--- 2. Local deployment mode (skipping SSH connection) ---")

		// For localhost, use local Caddy service
		log.Println("--- 2.5. Preparing local Caddy environment ---")
		d.caddySvc = caddy.NewLocalService()

		// Check Caddy availability
		if err := d.caddySvc.CheckAvailability(); err != nil {
			return err
		}
	}

	return nil
}

// handleInstanceLoadingError provides more specific guidance when loading an instance fails.
func (d *Deployer) handleInstanceLoadingError(err error) error {
	if errors.Is(err, database.ErrAppNotFound) {
		if _, statErr := os.Stat(config.ConfigPath); statErr == nil {
			return fmt.Errorf("Application '%s' is not registered in the database. Please run 'shipyard init' to sync it from '%s' to the database", d.AppName, config.ConfigPath)
		}
	} else if errors.Is(err, database.ErrInstanceNotFound) {
		fmt.Printf("Application '%s' is not linked to host '%s'. Link it now? (y/n): ", d.AppName, d.HostName)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "y" || input == "yes" {
			log.Printf("Linking application '%s' to host '%s'...", d.AppName, d.HostName)
			newInstance := &models.ApplicationInstance{
				ApplicationID: d.Application.ID,
				HostID:        d.Host.ID,
				Status:        "linked",
			}
			if linkErr := database.LinkApplicationToHost(newInstance); linkErr != nil {
				return fmt.Errorf("automatic linking failed: %w", linkErr)
			}
			log.Println("✅ Linking successful.")

			// Re-fetch instance info
			var fetchErr error
			d.Instance, d.Application, d.Host, fetchErr = database.GetInstance(d.AppName, d.HostName)
			if fetchErr != nil {
				return fmt.Errorf("failed to re-fetch instance info after linking: %w", fetchErr)
			}
			return nil // Success
		}
		return fmt.Errorf("Deployment aborted: Application '%s' is not linked to host '%s'.", d.AppName, d.HostName)
	}
	return err // Return other errors as-is
}

// initializeHostInteractively prompts the user to initialize an uninitialized host.
func (d *Deployer) initializeHostInteractively() error {
	fmt.Printf("Host '%s' is not initialized. Initialize it now? (y/n): ", d.HostName)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "y" || input == "yes" {
		log.Printf("Initializing host '%s'...", d.HostName)
		if err := InitializeHost(d.Host, d.Application.Name, d.detectRuntime(), "", "phoenix", d.HostKeyCallback); err != nil {
			return fmt.Errorf("automatic initialization failed: %w", err)
		}
		// Refresh host info to get the initialization timestamp
		refreshedHost, err := database.GetSSHHostByName(d.HostName)
		if err != nil {
			return fmt.Errorf("failed to reload host info: %w", err)
		}
		d.Host = refreshedHost
		return nil
	}
	return fmt.Errorf("Deployment aborted: Host '%s' is not initialized.", d.HostName)
}

// detectRuntime heuristically detects the runtime based on project characteristics.
func (d *Deployer) detectRuntime() string {
	if _, err := os.Stat("mix.exs"); err == nil {
		content, _ := os.ReadFile("mix.exs")
		if strings.Contains(string(content), ":phoenix") {
			return "phoenix"
		}
		// Pure Elixir project (no Phoenix)
		return "elixir"
	}
	if _, err := os.Stat("bin/server"); err == nil {
		return "phoenix"
	}
	if _, err := os.Stat("package.json"); err == nil {
		return "node"
	}
	if _, err := os.Stat("go.mod"); err == nil {
		return "golang"
	}
	// Check for static HTML files (index.html or a dist/build directory with index.html)
	if isStaticProject() {
		return "static"
	}
	return "elixir" // Default fallback for backward compatibility
}

// isStaticProject checks if the current directory is a static HTML project.
// It looks for:
// 1. index.html in the root directory
// 2. dist/index.html (common for Vue, React, etc.)
// 3. build/index.html (common for create-react-app)
// 4. public/index.html (common for some frameworks)
func isStaticProject() bool {
	staticPaths := []string{
		"index.html",
		"dist/index.html",
		"build/index.html",
		"public/index.html",
	}
	for _, p := range staticPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// connectSSH establishes an SSH connection to the remote host.
func (d *Deployer) connectSSH() error {
	var err error
	sshConfig, err := sshutil.NewClientConfig(d.Host, d.HostKeyCallback)
	if err != nil {
		return fmt.Errorf("failed to create SSH configuration: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", d.Host.Addr, d.Host.Port)
	d.SSHClient, err = ssh.Dial("tcp", addr, sshConfig)
	return err
}
