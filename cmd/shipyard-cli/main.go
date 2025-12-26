package main

import (
	"youfun/shipyard/cmd/shipyard-cli/commands"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/config"
	"flag"
	"fmt"
	"log"
)

var version = "dev"

func main() {
	// 1. Load saved CLI config (if exists) to set defaults
	cliConfig, _ := commands.LoadCLIConfig()
	defaultServerURL := "http://localhost:8080"
	var savedToken string

	if cliConfig != nil {
		if cliConfig.Endpoint != "" {
			defaultServerURL = cliConfig.Endpoint
		}
		savedToken = cliConfig.AccessToken
	}

	// 2. Parse global flags
	configPath := flag.String("config", "shipyard.toml", "Config file path (default: shipyard.toml)")
	serverURL := flag.String("server", defaultServerURL, "Deployer Server URL")
	flag.Parse()

	// Set the global config path
	config.ConfigPath = *configPath

	// Initialize API Client
	apiClient := client.NewClient(*serverURL)
	if savedToken != "" {
		apiClient.Token = savedToken
	}

	args := flag.Args()
	if len(args) < 1 {
		commands.PrintUsage()
		return
	}

	command := args[0]

	// Check for auth requirement
	if savedToken == "" && command != "login" && command != "version" && command != "help" && command != "logout" {
		log.Printf("⚠️  Warning: Login session not found. If subsequent operations fail (401 Unauthorized), please run 'shipyard-cli login' first.")
	}

	switch command {
	case "deploy":
		commands.DeployApp(apiClient)
	case "launch":
		commands.LaunchCommand(apiClient)
	case "login":
		commands.Login()
	case "logout":
		commands.Logout()
	case "vars":
		commands.VarsCommand(apiClient)
	case "logs":
		commands.LogsCommand(apiClient)
	case "app":
		commands.AppCommand(apiClient)
	case "build":
		commands.BuildCommand(apiClient)
	case "domain":
		commands.DomainCommand(apiClient)
	case "status", "info":
		commands.StatusCommand(apiClient)
	case "version":
		fmt.Println(version)
	case "help", "--help", "-h":
		commands.PrintUsage()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		commands.PrintUsage()
	}
}
