package deploy

import (
	"youfun/shipyard/internal/config"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

// prepareEnvVars merges configuration, domains, and secrets into a single map.
func (d *Deployer) prepareEnvVars(secrets map[string]string) map[string]interface{} {
	envs := make(map[string]interface{})

	// 1. Load from config (shipyard.toml)
	if len(config.AppConfig.Env) > 0 {
		log.Printf("Loading %d non-sensitive environment variables from shipyard.toml...", len(config.AppConfig.Env))
		for key, value := range config.AppConfig.Env {
			envs[key] = value
		}
	}

	// 2. Auto-generate PHX_HOST from domains
	if len(config.AppConfig.Domains) > 0 {
		phxHost := strings.Join(config.AppConfig.Domains, ",")
		envs["PHX_HOST"] = phxHost
		log.Printf("Automatically setting PHX_HOST environment variable: %s", phxHost)
	}

	// 3. Load Secrets (Override config)
	if len(secrets) > 0 {
		log.Printf("Loading %d secrets...", len(secrets))
		for key, value := range secrets {
			envs[key] = value
		}
	}

	return envs
}

// formatEnvValue converts a value to a string suitable for environment variables.
// Arrays are converted to comma-separated strings instead of Go's [a b c] format.
func formatEnvValue(value interface{}) string {
	switch v := value.(type) {
	case []interface{}:
		// Convert array to comma-separated string
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(v, ",")
	default:
		return fmt.Sprintf("%v", value)
	}
}

// injectEnvVars writes the environment variables to the remote server.
func (d *Deployer) injectEnvVars(envs map[string]interface{}) error {
	if len(envs) > 0 {
		var envContent strings.Builder
		for key, value := range envs {
			envContent.WriteString(fmt.Sprintf("%s=%s\n", key, formatEnvValue(value)))
		}
		b64 := base64.StdEncoding.EncodeToString([]byte(envContent.String()))
		remoteCmd := fmt.Sprintf("set -e; mkdir -p /etc/%s; echo '%s' | base64 -d | tee /etc/%s/env >/dev/null; (chown root:phoenix /etc/%s/env || true); (chmod 0640 /etc/%s/env || true)", d.AppName, b64, d.AppName, d.AppName, d.AppName)
		if err := d.executeRemoteCommand(remoteCmd, true); err != nil {
			return err
		}
		log.Printf("✅ %d environment variables written to /etc/%s/env", len(envs), d.AppName)
	}
	return nil
}
