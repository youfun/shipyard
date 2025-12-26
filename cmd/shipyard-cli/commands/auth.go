package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// CLIConfig holds the CLI configuration
type CLIConfig struct {
	Endpoint    string `json:"endpoint"`
	AccessToken string `json:"access_token"`
	DeviceName  string `json:"device_name"`
}

// GetCLIConfigPath returns the path to the CLI config file
func GetCLIConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".shipyard", "config.json")
}

// LoadCLIConfig loads the CLI configuration
func LoadCLIConfig() (*CLIConfig, error) {
	configPath := GetCLIConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config CLIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveCLIConfig saves the CLI configuration
func SaveCLIConfig(config *CLIConfig) error {
	configPath := GetCLIConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// Login handles the CLI login process using device flow
func Login() {
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	endpoint := loginCmd.String("endpoint", "", "Server endpoint (e.g., http://localhost:8080)")
	loginCmd.Parse(os.Args[2:])

	// Load saved endpoint if not provided via flag
	if *endpoint == "" {
		cliConfig, _ := LoadCLIConfig()
		if cliConfig != nil && cliConfig.Endpoint != "" {
			*endpoint = cliConfig.Endpoint
			fmt.Printf("Using saved endpoint: %s\n", *endpoint)
			fmt.Print("Press Enter to continue or type a new endpoint: ")
			var newEndpoint string
			fmt.Scanln(&newEndpoint)
			if newEndpoint != "" {
				*endpoint = newEndpoint
			}
		} else {
			fmt.Print("Enter server endpoint (e.g., http://localhost:8080): ")
			fmt.Scanln(endpoint)
		}
	}

	if *endpoint == "" {
		fmt.Println("Error: endpoint is required")
		os.Exit(1)
	}

	// Get device name
	deviceName, _ := os.Hostname()
	if deviceName == "" {
		deviceName = "unknown"
	}

	// Request device code
	reqBody, _ := json.Marshal(map[string]string{
		"os":          runtime.GOOS,
		"device_name": deviceName,
	})

	resp, err := http.Post(*endpoint+"/api/auth/device/code", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Response envelope structure
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			SessionID       string `json:"session_id"`
			UserCode        string `json:"user_code"`
			VerificationURI string `json:"verification_uri"`
			ExpiresIn       int    `json:"expires_in"`
			Interval        int    `json:"interval"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	codeResp := envelope.Data

	// Build verification URL
	verificationURL := *endpoint + codeResp.VerificationURI + "?session_id=" + codeResp.SessionID

	fmt.Println("\n=== CLI Authorization ===")
	fmt.Printf("User Code: %s\n", codeResp.UserCode)
	fmt.Printf("\nPlease open the following URL in your browser to authorize:\n")
	fmt.Printf("  %s\n\n", verificationURL)

	// Try to open browser automatically
	openBrowser(verificationURL)

	fmt.Println("Waiting for authorization...")

	// Poll for token
	interval := time.Duration(codeResp.Interval) * time.Second
	timeout := time.Now().Add(time.Duration(codeResp.ExpiresIn) * time.Second)

	for time.Now().Before(timeout) {
		time.Sleep(interval)

		tokenResp, err := http.Get(*endpoint + "/api/auth/device/token?session_id=" + codeResp.SessionID)
		if err != nil {
			continue
		}

		switch tokenResp.StatusCode {
		case http.StatusOK:
			var tokenEnvelope struct {
				Success bool `json:"success"`
				Data    struct {
					AccessToken string `json:"access_token"`
				} `json:"data"`
			}
			if err := json.NewDecoder(tokenResp.Body).Decode(&tokenEnvelope); err != nil {
				tokenResp.Body.Close()
				continue
			}
			tokenResp.Body.Close()
			tokenData := tokenEnvelope.Data

			// Save config
			config := &CLIConfig{
				Endpoint:    *endpoint,
				AccessToken: tokenData.AccessToken,
				DeviceName:  deviceName,
			}
			if err := SaveCLIConfig(config); err != nil {
				fmt.Printf("Warning: Could not save config: %v\n", err)
			}

			fmt.Println("\n✅ Login successful!")
			fmt.Printf("Configuration saved to %s\n", GetCLIConfigPath())
			return

		case http.StatusAccepted:
			// Authorization pending, continue polling
			tokenResp.Body.Close()
			continue

		case http.StatusForbidden:
			tokenResp.Body.Close()
			fmt.Println("\n❌ Authorization denied")
			os.Exit(1)

		case http.StatusGone:
			tokenResp.Body.Close()
			fmt.Println("\n❌ Session expired")
			os.Exit(1)

		default:
			tokenResp.Body.Close()
		}
	}

	fmt.Println("\n❌ Authorization timed out")
	os.Exit(1)
}

// openBrowser tries to open the URL in the default browser
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

// Logout handles the CLI logout process
func Logout() {
	configPath := GetCLIConfigPath()
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Logged out successfully")
}
