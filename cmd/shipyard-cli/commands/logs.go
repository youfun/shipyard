package commands

import (
	"context"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/cliutils"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/logs"
	"youfun/shipyard/internal/models"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

// logsCommand handles viewing application logs via SSH
func LogsCommand(apiClient *client.Client) {
	// Define and parse flags
	logsCmd := flag.NewFlagSet("logs", flag.ExitOnError)
	hostFlag := logsCmd.String("host", "", "Target host name (optional, enters interactive mode if not provided)")
	portFlag := logsCmd.Int("port", 0, "Specific port number (optional)")
	linesFlag := logsCmd.Int("lines", 500, "Number of log lines to show")
	followFlag := logsCmd.Bool("follow", false, "Follow log output")
	fFlag := logsCmd.Bool("f", false, "Follow log output (shorthand for -f)")
	colorFlag := logsCmd.Bool("color", true, "Enable color output")
	noColorFlag := logsCmd.Bool("no-color", false, "Disable color output")

	logsCmd.Parse(os.Args[2:])

	// Handle -f shorthand
	if *fFlag {
		*followFlag = true
	}

	// Handle --no-color
	if *noColorFlag {
		*colorFlag = false
	}

	// First determine appName
	var appName string
	if len(logsCmd.Args()) > 0 {
		appName = logsCmd.Args()[0]
	} else {
		var projConf config.Config
		if _, err := toml.DecodeFile(config.ConfigPath, &projConf); err == nil {
			appName = projConf.App
		}
		if appName == "" {
			fmt.Printf("Error: Could not determine application name\n")
			fmt.Println("Usage: shipyard-cli logs <app-name> --host <host-name> [options]")
			os.Exit(1)
		}
	}

	// If host is not provided via flag, enter interactive selection
	if *hostFlag == "" {
		hosts, err := apiClient.ListHosts()
		if err != nil {
			log.Fatalf("Failed to get host list: %v", err)
		}

		if len(hosts) == 0 {
			log.Fatal("Error: No available hosts on server.")
		}

		var items []string
		for _, h := range hosts {
			items = append(items, fmt.Sprintf("%s (%s:%d)", h.Name, h.Addr, h.Port))
		}

		selectedIndex, err := cliutils.PromptForSelection("\n--- Please select a host to view logs ---", items, -1)
		if err != nil {
			log.Fatalf("Error selecting host: %v", err)
		}
		*hostFlag = hosts[selectedIndex].Name
		fmt.Println() // Add a newline for better formatting after selection
	}

	// Get instance info
	instanceInfo, err := apiClient.GetInstance(appName, *hostFlag)
	if err != nil {
		log.Fatalf("Failed to get instance info: %v", err)
	}

	// Determine target port
	var targetPort int
	var deployInfo string

	if *portFlag > 0 {
		// Use specified port
		targetPort = *portFlag
		deployInfo = fmt.Sprintf("Port %d", targetPort)
	} else {
		// Use active port
		if instanceInfo.Instance.ActivePort == 0 {
			log.Fatal("Error: No active deployment instance found")
		}
		targetPort = int(instanceInfo.Instance.ActivePort)
		deployInfo = fmt.Sprintf("Active Port %d", targetPort)
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

	// Show info message
	fmt.Printf("正在获取 %s@%s 的日志...\n", appName, *hostFlag)
	fmt.Printf("目标: %s\n", deployInfo)
	if *followFlag {
		fmt.Println("模式: 实时跟踪 (按 Ctrl+C 退出)")
	} else {
		fmt.Printf("模式: 静态查看 (最近 %d 行)\n", *linesFlag)
	}
	fmt.Println(strings.Repeat("-", 80))

	// Fetch logs
	if *followFlag {
		// Real-time mode using WebSocket API
		ctx := context.Background()
		err = apiClient.StreamInstanceLogs(ctx, instanceInfo.Instance.UID, *linesFlag)
		if err != nil {
			fmt.Printf("\n错误: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Static mode - direct SSH connection
		logContent, err := logs.FetchJournalLogs(host, appName, targetPort, *linesFlag, false, ssh.InsecureIgnoreHostKey())
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}

		// Parse and colorize (if enabled)
		if *colorFlag {
			lines := strings.Split(logContent, "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				colored := logs.ParseAndColorLine(line, true)
				fmt.Println(colored)
			}
		} else {
			fmt.Print(logContent)
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("已显示最近 %d 行日志\n", *linesFlag)
	}
}
