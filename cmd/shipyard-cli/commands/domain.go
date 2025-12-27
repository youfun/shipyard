package commands

import (
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/cliutils"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

// domainCommand handles the 'domain' command
func DomainCommand(apiClient *client.Client) {
	if len(os.Args) < 3 {
		printDomainUsage()
		return
	}

	switch os.Args[2] {
	case "check":
		domainCheckCommand(apiClient)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[2])
		printDomainUsage()
	}
}

func domainCheckCommand(apiClient *client.Client) {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	hostName := checkCmd.String("host", "", "The hostname to check")
	appName := checkCmd.String("app", "", "The app name (optional, defaults to shipyard.toml)")
	checkCmd.Parse(os.Args[3:])

	// Resolve app name
	finalAppName := *appName
	if finalAppName == "" {
		var projConf config.Config
		if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err == nil {
			finalAppName = projConf.App
		}
		if finalAppName == "" {
			log.Fatalf("Error: Could not determine app name. Please use --app or ensure %s file exists.", config.ConfigPath)
		}
	}

	// Resolve host name
	finalHostName := *hostName
	if finalHostName == "" {
		// Interactive host selection via API
		hosts, err := apiClient.ListHosts()
		if err != nil {
			log.Fatalf("Failed to get host list: %v", err)
		}
		if len(hosts) == 0 {
			log.Fatal("Error: No available hosts on server.")
		}
		if len(hosts) == 1 {
			finalHostName = hosts[0].Name
			log.Printf("Automatically selected unique host: %s", finalHostName)
		} else {
			var items []string
			for _, h := range hosts {
				items = append(items, fmt.Sprintf("%s (%s:%d)", h.Name, h.Addr, h.Port))
			}
			selectedIndex, err := cliutils.PromptForSelection("\n--- Please select a host ---", items, -1)
			if err != nil {
				log.Fatalf("Error selecting host: %v", err)
			}
			finalHostName = hosts[selectedIndex].Name
		}
	}

	// Get instance info from API (includes host credentials)
	instanceInfo, err := apiClient.GetInstance(finalAppName, finalHostName)
	if err != nil {
		log.Fatalf("Failed to get instance info: %v", err)
	}

	// Build SSH host model
	host := &models.SSHHost{
		Name: instanceInfo.Host.Name,
		Addr: instanceInfo.Host.Addr,
		Port: instanceInfo.Host.Port,
		User: instanceInfo.Host.User,
	}
	if instanceInfo.Host.Password != nil {
		host.Password = instanceInfo.Host.Password
	}
	if instanceInfo.Host.PrivateKey != nil {
		host.PrivateKey = instanceInfo.Host.PrivateKey
	}

	log.Printf("Connecting to host %s (%s)...", host.Name, host.Addr)

	sshConfig, err := sshutil.NewClientConfig(host, ssh.InsecureIgnoreHostKey())
	if err != nil {
		log.Fatalf("Could not create SSH client config: %v", err)
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Addr, host.Port), sshConfig)
	if err != nil {
		log.Fatalf("Could not create SSH client: %v", err)
	}
	defer sshClient.Close()

	caddyService := caddy.NewService(sshClient)

	log.Println("Getting Caddy config...")
	caddyConfig, err := caddyService.GetConfig("/")
	if err != nil {
		log.Fatalf("Could not get Caddy config: %v", err)
	}

	prettyJSON, err := json.MarshalIndent(caddyConfig, "", "  ")
	if err != nil {
		log.Fatalf("Could not format JSON: %v", err)
	}

	fmt.Println(string(prettyJSON))
}

func printDomainUsage() {
	fmt.Println("Usage: shipyard-cli domain <subcommand> [options]")
	fmt.Println("Subcommands:")
	fmt.Println("  check    Check the Caddy configuration for the domain")
	fmt.Println("Options:")
	fmt.Println("  --app <appname>      (Optional) Specify the app name")
	fmt.Println("  --host <hostname>    (Optional) Specify the host to check")
}
