package config

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

// Hook defines a single command to be executed at a specific lifecycle stage.
type Hook struct {
	Name    string `toml:"name"`
	Type    string `toml:"type"`
	Command string `toml:"command"`
}

// Hooks contains lists of hooks for different deployment stages.
type Hooks struct {
	PreDeploy  []Hook `toml:"pre_deploy"`
	Migrate    []Hook `toml:"migrate"`
	PostDeploy []Hook `toml:"post_deploy"`
	// Future hooks like pre_start, post_start can be added here.
}

// Config stores the full configuration loaded from shipyard.toml
type Config struct {
	App           string                 `toml:"app"`
	Domains       []string               `toml:"domains"`        // support multiple domains
	PrimaryDomain string                 `toml:"primary_domain"` // primary domain (optional)
	Runtime       string                 `toml:"runtime"`        // phoenix|node|golang, can be empty for auto-detection
	Env           map[string]interface{} `toml:"env"`
	Hooks         Hooks                  `toml:"hooks"`
	KeepReleases  int                    `toml:"keep_releases"` // number of old releases to keep, default 3
}

var AppConfig Config

// ConfigPath stores the path to the config file in use; can be overridden by --config
var ConfigPath = "shipyard.toml"

// LoadConfig loads the configuration file at the given path and sets defaults.
// configPath: path to the configuration file, defaults to "shipyard.toml"
func LoadConfig(appName string, configPath string) {
	if configPath == "" {
		configPath = "shipyard.toml"
	}

	if _, err := toml.DecodeFile(configPath, &AppConfig); err != nil {
		log.Printf("Warning: %s not found or failed to parse: %v. Using default configuration.", configPath, err)
	}

	// if app is not set in toml, use provided parameter or leave empty
	if AppConfig.App == "" {
		AppConfig.App = appName
	}

	// Backwards compatibility: if PHX_HOST is provided in env and domains not set, read it
	if len(AppConfig.Domains) == 0 {
		if phxHost, ok := AppConfig.Env["PHX_HOST"]; ok {
			if hostStr, ok := phxHost.(string); ok && hostStr != "" {
				AppConfig.Domains = []string{hostStr}
				log.Printf("Read domain from PHX_HOST environment variable: %s", hostStr)
			}
		}
	}

	// If primary_domain is set, ensure it's present in domains list
	if AppConfig.PrimaryDomain != "" {
		found := false
		for _, domain := range AppConfig.Domains {
			if domain == AppConfig.PrimaryDomain {
				found = true
				break
			}
		}
		if !found {
			log.Printf("Warning: primary_domain '%s' is not present in domains list", AppConfig.PrimaryDomain)
		}
	}

	// If primary_domain is not set, use the first domain in the domains list
	if AppConfig.PrimaryDomain == "" && len(AppConfig.Domains) > 0 {
		AppConfig.PrimaryDomain = AppConfig.Domains[0]
	}

	if AppConfig.Runtime == "" {
		log.Println("Runtime not explicitly configured; will auto-detect during deployment (phoenix|node|golang|static)")
	} else {
		log.Printf("runtime=%s (from %s)", AppConfig.Runtime, configPath)
	}

	// set default KeepReleases if not configured
	if AppConfig.KeepReleases == 0 {
		AppConfig.KeepReleases = 3 // default keep 3 old releases
		log.Printf("keep_releases not configured, using default value %d.", AppConfig.KeepReleases)
	}

	log.Printf("Configuration loaded (from %s).", configPath)
}

// GetRemoteReleasesDir helper function to get the remote releases directory
func GetRemoteReleasesDir() string {

	// fallback to default path based on app name if old config missing
	if AppConfig.App == "" {
		log.Fatal("Error: application name is not configured in shipyard.toml.")
	}
	return fmt.Sprintf("/var/www/%s/releases", AppConfig.App)
}

// ReadConfigFile reads the configuration file from the given path and returns a Config object
func ReadConfigFile(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "shipyard.toml"
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read or parse %s: %w", configPath, err)
	}

	return &cfg, nil
}
