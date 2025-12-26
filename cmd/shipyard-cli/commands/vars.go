package commands

import (
	"bufio"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/config"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// varsCommand handles listing, setting, and unsetting application variables via API.
func VarsCommand(apiClient *client.Client) {
	if len(os.Args) < 3 {
		log.Println("Error: vars command requires a subcommand: list, set, or unset")
		fmt.Println("Example: shipyard-cli vars list")
		fmt.Println("Example: shipyard-cli vars set KEY=VALUE")
		os.Exit(1)
	}

	subCmd := os.Args[2]
	// Use a custom flag set to avoid parsing conflicts
	fs := flag.NewFlagSet("vars-"+subCmd, flag.ExitOnError)
	appName := fs.String("app", "", "Application name (optional, defaults to shipyard.toml)")
	fs.Parse(os.Args[3:])

	// If app name is not provided via flag, try to read from shipyard.toml (or custom config)
	if *appName == "" {
		var projConf config.Config
		if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err == nil {
			*appName = projConf.App
		}
	}

	// If still no app name, it's an error for all subcommands.
	if *appName == "" {
		log.Fatalf("Error: --app flag required, or define 'app' in %s", config.ConfigPath)
	}

	switch subCmd {
	case "list":
		varsListCommand(apiClient, *appName)
	case "set":
		varsSetCommand(apiClient, *appName, fs.Args())
	case "unset":
		varsUnsetCommand(apiClient, *appName, fs.Args())
	default:
		log.Fatal("Unknown vars subcommand. Please use list, set, or unset.")
	}
}

func varsListCommand(apiClient *client.Client, appName string) {
	fmt.Printf("--- Environment Variables for '%s' ---\n\n", appName)

	// 1. List non-sensitive variable keys from config file
	fmt.Printf("--- From %s ---\n", config.ConfigPath)
	var projConf config.Config
	if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err != nil {
		fmt.Printf("Could not read %s, skipping.\n", config.ConfigPath)
	} else {
		if len(projConf.Env) == 0 {
			fmt.Println("None found.")
		} else {
			for key := range projConf.Env {
				fmt.Println(key)
			}
		}
	}

	// 2. List secret keys from the server via API
	fmt.Println("\n--- Secrets (from Server) ---")
	keys, err := apiClient.ListSecrets(appName)
	if err != nil {
		fmt.Printf("Failed to get secrets from server: %v\n", err)
		fmt.Println("App might not be synced to server, or no secrets set.")
	} else if len(keys) == 0 {
		fmt.Println("None found.")
	} else {
		for _, key := range keys {
			fmt.Println(key)
		}
	}
}

func varsSetCommand(apiClient *client.Client, appName string, args []string) {
	if len(args) == 0 {
		log.Fatal("Error: Requires at least one KEY=VALUE argument, or just KEY for interactive input")
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		key := parts[0]
		var value string

		if len(parts) == 2 {
			value = parts[1]
		} else if len(parts) == 1 {
			// Interactive mode
			fmt.Printf("Enter value for '%s' (press Enter to confirm): ", key)
			var input string
			// We need to read the full line to support spaces and special chars
			// bufio.Scanner is better for this than fmt.Scanln
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				input = scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				log.Fatalf("Failed to read input: %v", err)
			}
			value = input
		} else {
			log.Printf("Ignoring invalid key format: %s", arg)
			continue
		}

		if err := apiClient.SetSecret(appName, key, value); err != nil {
			log.Fatalf("Failed to set secret '%s': %v", key, err)
		}
		fmt.Printf("✅ Secret '%s' set for application '%s'.\n", key, appName)
	}
}

func varsUnsetCommand(apiClient *client.Client, appName string, args []string) {
	if len(args) == 0 {
		log.Fatal("Error: Requires at least one secret key name")
	}

	for _, key := range args {
		if err := apiClient.UnsetSecret(appName, key); err != nil {
			log.Fatalf("Failed to delete secret '%s': %v", key, err)
		}
		fmt.Printf("✅ Secret '%s' deleted for application '%s'.\n", key, appName)
	}
}
