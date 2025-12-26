package utils

import (
	"youfun/shipyard/internal/depsinstall"
	"fmt"
	"log"
	"strings"
)

// IsLocalhost checks if the given hostname/address is localhost
func IsLocalhost(host string) bool {
	if host == "" {
		return false
	}

	localNames := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
	}

	host = strings.ToLower(strings.TrimSpace(host))
	for _, local := range localNames {
		if host == local {
			return true
		}
	}

	return false
}

// CheckCaddyInstalled checks if Caddy is installed and returns its version
func CheckCaddyInstalled() (bool, string, error) {
	return depsinstall.CheckCaddyInstalled()
}

// CheckCaddyRunning checks if Caddy server is running
func CheckCaddyRunning() (bool, error) {
	return depsinstall.CheckCaddyRunning()
}

// InstallCaddyLocal installs Caddy on the local machine (non-Windows)
func InstallCaddyLocal() error {
	return depsinstall.InstallCaddy()
}

// StartCaddy attempts to start the Caddy server
func StartCaddy() error {
	return depsinstall.StartCaddy()
}

// EnsureCaddyRunning checks if Caddy is installed and running, starts it if needed
func EnsureCaddyRunning() error {
	// Check if Caddy is installed
	installed, version, err := CheckCaddyInstalled()
	if !installed {
		log.Println("❌ Caddy is not installed.")

		// Ask user if they want to install Caddy
		if PromptForConfirmation("Install Caddy automatically?", true) {
			if err := InstallCaddyLocal(); err != nil {
				return fmt.Errorf("failed to install Caddy: %w", err)
			}
			// Verify installation
			installed, version, err = CheckCaddyInstalled()
			if !installed {
				return fmt.Errorf("Caddy installation verification failed")
			}
			log.Printf("✅ Caddy installed: %s", version)
		} else {
			return fmt.Errorf("Caddy is required for localhost deployment. Please install manually: https://caddyserver.com/docs/install")
		}
	} else {
		log.Printf("✅ Caddy detected: %s", version)
	}

	// Check if Caddy is running
	running, err := CheckCaddyRunning()
	if err != nil {
		return fmt.Errorf("failed to check Caddy status: %w", err)
	}

	if !running {
		log.Println("⚠️  Caddy service is not running")

		// Ask user if they want to start Caddy
		if PromptForConfirmation("Start Caddy service?", true) {
			if err := StartCaddy(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Caddy service must be running for localhost deployment")
		}
	} else {
		log.Println("✅ Caddy service is running")
	}

	return nil
}

// GetLocalhostInfo returns information about deploying to localhost
func GetLocalhostInfo() string {
	return "Server Machine (localhost) - Deploy to the server where shipyard-server is running"
}
