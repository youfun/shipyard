package commands

import (
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/config"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

// statusCommand handles the 'status' or 'info' command - shows app status via API
func StatusCommand(apiClient *client.Client) {
	var projConf config.Config
	if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err != nil {
		log.Fatalf("Failed to read or parse %s: %v. Please ensure file exists and is correctly formatted.", config.ConfigPath, err)
	}

	if projConf.App == "" {
		log.Fatalf("'app' config item not found in %s.", config.ConfigPath)
	}

	appName := projConf.App

	// Get hosts list
	hosts, err := apiClient.ListHosts()
	if err != nil {
		log.Fatalf("Failed to get host list: %v", err)
	}

	fmt.Printf("--- App Info: %s ---\n", appName)
	fmt.Println("(Domain info is stored on server; use 'shipyard-cli domain check' to view domain config)")
	fmt.Println("\n--- Deployment Instances ---")

	if len(hosts) == 0 {
		fmt.Println("No available hosts on server.")
		return
	}

	foundInstances := false
	for _, host := range hosts {
		// Try to get instance info for each host
		instanceInfo, err := apiClient.GetInstance(appName, host.Name)
		if err != nil {
			// App might not be linked to this host, skip
			continue
		}

		foundInstances = true
		status := instanceInfo.Instance.Status
		if instanceInfo.Instance.ActivePort > 0 {
			status = fmt.Sprintf("active on port %d", instanceInfo.Instance.ActivePort)
		}
		fmt.Printf("- Host: %s, Status: %s\n", host.Name, status)
	}

	if !foundInstances {
		fmt.Println("No deployment instances found.")
		fmt.Println("Use 'shipyard-cli launch' to create first deployment.")
	}
}
