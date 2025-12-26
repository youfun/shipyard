package deploy

import (
	"youfun/shipyard/internal/config"
	"fmt"
	"strings"
	"testing"
)

// TestPrepareEnvVars verifies environment variable merging logic
// This test corresponds to the prepareEnvVars function to be extracted from execute
func TestPrepareEnvVars(t *testing.T) {
	// Save original config to restore later
	originalEnv := config.AppConfig.Env
	originalDomains := config.AppConfig.Domains
	defer func() {
		config.AppConfig.Env = originalEnv
		config.AppConfig.Domains = originalDomains
	}()

	// Simulate shipyard.toml config
	config.AppConfig.Env = map[string]interface{}{
		"PORT":      "4000",
		"LOG_LEVEL": "info",
	}
	config.AppConfig.Domains = []string{"example.com", "www.example.com"}

	tests := []struct {
		name           string
		runtime        string
		secrets        map[string]string
		expectedValues map[string]string
	}{
		{
			name:    "Phoenix App with Secrets",
			runtime: "phoenix",
			secrets: map[string]string{
				"DATABASE_URL":    "postgres://user:pass@localhost/db",
				"SECRET_KEY_BASE": "existing_secret_key",
			},
			expectedValues: map[string]string{
				"PORT":            "4000",
				"LOG_LEVEL":       "info",
				"PHX_HOST":        "example.com,www.example.com",
				"DATABASE_URL":    "postgres://user:pass@localhost/db",
				"SECRET_KEY_BASE": "existing_secret_key",
			},
		},
		{
			name:    "Node App without Secrets",
			runtime: "node",
			secrets: map[string]string{},
			expectedValues: map[string]string{
				"PORT":      "4000",
				"LOG_LEVEL": "info",
				"PHX_HOST":  "example.com,www.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Simulate logic to be extracted Start ---
			// This logic will become func (d *Deployer) prepareEnvVars(secrets map[string]string) map[string]string
			envs := make(map[string]interface{})

			// 1. Load from config (shipyard.toml)
			for k, v := range config.AppConfig.Env {
				envs[k] = v
			}

			// 2. Auto-generate PHX_HOST
			if len(config.AppConfig.Domains) > 0 {
				envs["PHX_HOST"] = strings.Join(config.AppConfig.Domains, ",")
			}

			// 3. Load Secrets (Override config)
			for k, v := range tt.secrets {
				envs[k] = v
			}
			// --- Simulate logic to be extracted End ---

			// Verify
			for k, want := range tt.expectedValues {
				gotVal, exists := envs[k]
				if !exists {
					t.Errorf("Expected env var %s to exist, but it was missing", k)
					continue
				}

				// envs values are interface{}, cast to string for comparison if possible,
				// or just compare string representation
				gotStr, ok := gotVal.(string)
				if !ok {
					t.Errorf("Env var %s value is not a string", k)
					continue
				}

				if gotStr != want {
					t.Errorf("Env var %s = %q; want %q", k, gotStr, want)
				}
			}
		})
	}
}

// TestGetPermissionCommands verifies permission command generation logic
func TestGetPermissionCommands(t *testing.T) {
	tests := []struct {
		name        string
		runtime     string
		releasePath string
		wantEmpty   bool
		wantCmds    []string
	}{
		{
			name:        "Static App",
			runtime:     "static",
			releasePath: "/var/www/app/releases/v1",
			wantEmpty:   true,
		},
		{
			name:        "Phoenix App",
			runtime:     "phoenix",
			releasePath: "/var/www/app/releases/v1",
			wantEmpty:   false,
			wantCmds: []string{
				"cd /var/www/app/releases/v1 && if [ -d bin ]; then chmod +x bin/* || true; fi",
				"cd /var/www/app/releases/v1 && if ls -d erts-*/bin >/dev/null 2>&1; then find erts-*/bin -type f -exec chmod +x {} +; fi",
				"cd /var/www/app/releases/v1 && if [ -d releases ]; then find releases -name elixir -type f -exec chmod +x {} +; fi",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Simulate logic to be extracted Start ---
			// func (d *Deployer) getPermissionCommands(releasePath string) []string
			var cmds []string
			if tt.runtime != "static" {
				cmds = append(cmds, fmt.Sprintf("cd %s && if [ -d bin ]; then chmod +x bin/* || true; fi", tt.releasePath))
				cmds = append(cmds, fmt.Sprintf("cd %s && if ls -d erts-*/bin >/dev/null 2>&1; then find erts-*/bin -type f -exec chmod +x {} +; fi", tt.releasePath))
				cmds = append(cmds, fmt.Sprintf("cd %s && if [ -d releases ]; then find releases -name elixir -type f -exec chmod +x {} +; fi", tt.releasePath))
			}
			// --- Simulate logic to be extracted End ---

			if tt.wantEmpty {
				if len(cmds) > 0 {
					t.Errorf("Expected no commands for runtime %s, got %d", tt.runtime, len(cmds))
				}
			} else {
				if len(cmds) != len(tt.wantCmds) {
					t.Errorf("Expected %d commands, got %d", len(tt.wantCmds), len(cmds))
				}
			}
		})
	}
}

// TestGetSystemdEnableCommand verifies Systemd enable command generation logic
func TestGetSystemdEnableCommand(t *testing.T) {
	appName := "my-app"
	port := 4000
	expected := "systemctl enable my-app@4000"

	// Simulate logic to be extracted
	cmd := fmt.Sprintf("systemctl enable %s@%d", appName, port)

	if cmd != expected {
		t.Errorf("Expected command %q, got %q", expected, cmd)
	}
}
