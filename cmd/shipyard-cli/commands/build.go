package commands

import (
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/config"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// BuildCommand handles the 'build' command with subcommands
func BuildCommand(apiClient *client.Client) {
	if len(os.Args) < 3 {
		printBuildUsage()
		os.Exit(1)
	}

	subCommand := os.Args[2]
	switch subCommand {
	case "list":
		buildListCommand(apiClient)
	case "help", "--help", "-h":
		printBuildUsage()
	default:
		fmt.Printf("Unknown build subcommand: %s\n", subCommand)
		printBuildUsage()
		os.Exit(1)
	}
}

func printBuildUsage() {
	fmt.Print(`
Usage: shipyard-cli build <subcommand> [options]

Subcommands:
  list        List build artifacts for an application

Options:
  --app       Application name (optional, defaults to shipyard.toml)

Example:
  shipyard-cli build list
  shipyard-cli build list --app my-app
`)
}

// buildListCommand handles the 'build list' command
func buildListCommand(apiClient *client.Client) {
	cmd := flag.NewFlagSet("build list", flag.ExitOnError)
	appFlag := cmd.String("app", "", "Application name (optional)")
	cmd.Parse(os.Args[3:])

	// Resolve app name
	appName := *appFlag
	if appName == "" {
		var projConf config.Config
		if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err == nil {
			appName = projConf.App
		}
		if appName == "" {
			log.Fatalf("❌ Could not determine application name. Please use --app or run in project directory (shipyard.toml required)")
		}
	}

	log.Printf("--- Build Artifacts for app '%s' ---", appName)

	artifacts, err := apiClient.ListBuildArtifacts(appName)
	if err != nil {
		log.Fatalf("❌ Failed to list build artifacts: %v", err)
	}

	if len(artifacts) == 0 {
		fmt.Println("No build artifacts found.")
		return
	}

	// Print header
	fmt.Printf("\n%-16s %-12s %-40s %-20s\n", "VERSION", "MD5 (short)", "GIT COMMIT SHA", "CREATED AT")
	fmt.Println("--------------------------------------------------------------------------------")

	// Print artifacts
	for _, artifact := range artifacts {
		md5Short := artifact.MD5Hash
		if len(md5Short) > 10 {
			md5Short = md5Short[:10]
		}
		gitSHA := artifact.GitCommitSHA
		if len(gitSHA) > 40 {
			gitSHA = gitSHA[:40]
		}
		createdAt := ""
		if artifact.CreatedAt != nil {
			createdAt = artifact.CreatedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-16s %-12s %-40s %-20s\n", artifact.Version, md5Short, gitSHA, createdAt)
	}

	fmt.Printf("\nTotal: %d build(s)\n", len(artifacts))
}
