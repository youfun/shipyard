package depsinstall

import (
	"bytes"
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/static"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// InitializeCaddyConfig initializes Caddy with basic configuration structure
// This should be called after Caddy is installed and running
// Note: The install script already creates /etc/caddy/caddy.json with the apps node,
// so we only need to call SetupCaddy() to ensure the HTTP app structure is ready
func InitializeCaddyConfig() error {
	log.Println("🔧 Initializing Caddy configuration...")

	// The install script creates a default caddy.json with apps node already,
	// so we can directly setup the HTTP app structure
	caddySvc := caddy.NewLocalService()
	return caddySvc.SetupCaddy()
}

// CheckCaddyInstalled checks if Caddy is installed locally and returns its version
func CheckCaddyInstalled() (bool, string, error) {
	cmd := exec.Command("caddy", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", fmt.Errorf("caddy is not installed or not executable: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return true, version, nil
}

// CheckCaddyRunning checks if Caddy server is running locally
func CheckCaddyRunning() (bool, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://localhost:2019/config/")
	if err != nil {
		return false, nil // Caddy is not running
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// InstallCaddy installs Caddy on the local machine (non-Windows)
func InstallCaddy() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("automatic installation is not supported on Windows, please install manually: https://caddyserver.com/docs/install")
	}

	log.Println("Preparing to install Caddy...")

	// Write the install script to a temporary file
	tmpFile, err := os.CreateTemp("", "install_caddy_*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temporary install script: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(static.InstallCaddyScript)); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write install script: %w", err)
	}
	tmpFile.Close()

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to set script permissions: %w", err)
	}

	log.Println("Executing install script (this may take a few minutes)...")
	cmd := exec.Command("sudo", "bash", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Caddy installation failed: %w", err)
	}

	log.Println("✅ Caddy installed successfully!")
	return nil
}

// StartCaddy attempts to start the Caddy server locally
func StartCaddy() error {
	log.Println("🚀 Starting Caddy server...")

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// On Windows, start Caddy in the background
		cmd = exec.Command("cmd", "/C", "start", "/B", "caddy", "run")
		if err := cmd.Run(); err != nil {
			return err
		}
	case "linux", "darwin":
		// On Linux/Mac, try systemd first
		systemdCmd := exec.Command("systemctl", "start", "caddy")
		if err := systemdCmd.Run(); err == nil {
			log.Println("✅ Caddy service started via systemd")
			// Wait a bit for Caddy to start
			time.Sleep(2 * time.Second)
			if running, _ := CheckCaddyRunning(); running {
				log.Println("✅ Caddy service is running")
				return nil
			}
			return fmt.Errorf("failed to connect to Caddy Admin API after systemd start")
		}

		// Fallback to direct command
		cmd = exec.Command("caddy", "run")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start Caddy: %w", err)
		}

		// Detach from parent
		go cmd.Wait()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Wait a bit for Caddy to start
	time.Sleep(2 * time.Second)

	// Verify it's running
	if running, _ := CheckCaddyRunning(); running {
		log.Println("✅ Caddy service is running")
		return nil
	}

	return fmt.Errorf("failed to connect to Caddy Admin API after start")
}

// EnsureCaddyRunning checks if Caddy is installed and running, installs/starts it if needed
// If autoInstall is false, will return error instead of installing
// If autoStart is false, will return error instead of starting
func EnsureCaddyRunning(autoInstall, autoStart bool) error {
	// Check if Caddy is installed
	installed, version, err := CheckCaddyInstalled()
	if !installed {
		log.Println("❌ Caddy is not installed.")

		if !autoInstall {
			return fmt.Errorf("Caddy is required but not installed. Please install manually: https://caddyserver.com/docs/install")
		}

		if err := InstallCaddy(); err != nil {
			return fmt.Errorf("failed to install Caddy: %w", err)
		}

		// Verify installation
		installed, version, err = CheckCaddyInstalled()
		if !installed {
			return fmt.Errorf("Caddy installation verification failed")
		}
		log.Printf("✅ Caddy installed: %s", version)
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

		if !autoStart {
			return fmt.Errorf("Caddy service must be running")
		}

		if err := StartCaddy(); err != nil {
			return err
		}
	} else {
		log.Println("✅ Caddy service is running")
	}

	// Initialize Caddy configuration after ensuring it's running
	if err := InitializeCaddyConfig(); err != nil {
		log.Printf("⚠️  Failed to initialize Caddy configuration: %v", err)
		// Don't fail here as it might already be configured
	}

	return nil
}

// CheckAndInstallCaddyRemote checks for Caddy on remote host and installs it if missing
func CheckAndInstallCaddyRemote(client *ssh.Client, autoInstall bool) error {
	log.Println("🔍 Checking if Caddy is installed on the remote server...")

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stderrBuf bytes.Buffer
	session.Stderr = &stderrBuf

	err = session.Run("command -v caddy")
	if err == nil {
		log.Println("✅ Caddy is already installed on remote server.")
		return nil
	}

	log.Println("📦 Caddy not found on remote server.")

	if !autoInstall {
		return fmt.Errorf("Caddy is not installed on remote server and automatic installation is disabled")
	}

	log.Println("Preparing to install Caddy on remote server...")

	// Use embedded Caddy install script
	caddyScriptBytes := []byte(static.InstallCaddyScript)

	// Convert CRLF to LF for Unix-like systems
	caddyScriptBytes = bytes.ReplaceAll(caddyScriptBytes, []byte("\r\n"), []byte("\n"))
	caddyB64 := base64.StdEncoding.EncodeToString(caddyScriptBytes)

	// Execute Caddy installation script
	installSess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer installSess.Close()

	installCmd := fmt.Sprintf("bash -c \"mkdir -p /tmp; echo '%s' | base64 -d > /tmp/install_caddy_github.sh; chmod +x /tmp/install_caddy_github.sh; bash /tmp/install_caddy_github.sh\"", caddyB64)
	log.Println("🚀 Executing Caddy install script on remote...")

	installSess.Stdout = os.Stdout
	installSess.Stderr = os.Stderr

	if err := installSess.Run(installCmd); err != nil {
		return fmt.Errorf("Caddy install script execution failed: %w", err)
	}

	log.Println("✅ Caddy installation completed on remote server.")
	return nil
}
