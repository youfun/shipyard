package commands

import (
	"flag"
	"os"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/cliutils"
	"youfun/shipyard/internal/deploy"

	"golang.org/x/crypto/ssh"
)

func DeployApp(apiClient *client.Client) {
	cmd := flag.NewFlagSet("deploy", flag.ExitOnError)
	appNameFlag := cmd.String("app", "", "Application name (optional, defaults to shipyard.toml)")
	hostNameFlag := cmd.String("host", "", "Host name (optional, defaults to interactive selection)")
	useBuild := cmd.String("use-build", "", "Reuse build artifact by MD5 (short), git commit SHA, or version (see: build list)")
	cmd.Parse(os.Args[2:])

	// Resolve app name: flag > shipyard.toml
	appName := *appNameFlag
	if appName == "" {
		appName = cliutils.ResolveAppNameFromConfig()
	}

	// Resolve host name: flag > interactive selection
	hostName := *hostNameFlag
	if hostName == "" {
		// Fetch last deployment info to suggest default host
		lastDeployment, _ := apiClient.GetLastDeployment(appName)
		var lastHostName string
		if lastDeployment != nil {
			lastHostName = lastDeployment.HostName
		}

		hostDTO := selectHost(apiClient, "", lastHostName)
		hostName = hostDTO.Name
	}

	// TODO: Implement proper host key verification for CLI client using API
	deploy.RunWithAPIClient(apiClient, appName, hostName, *useBuild, ssh.InsecureIgnoreHostKey())
}
