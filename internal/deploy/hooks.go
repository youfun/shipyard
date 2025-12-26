package deploy

import (
	"bufio"
	"youfun/shipyard/internal/config"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/pkg/sftp"
)

// substituteVariables replaces placeholders in a hook command string.
func (d *Deployer) substituteVariables(command string) string {
	r := strings.NewReplacer(
		"{{release_path}}", d.CurrentReleasePath,
		"{{app_name}}", d.AppName,
		"{{version}}", d.Version,
		"{{commit_sha}}", d.GitCommitSHA,
	)
	return r.Replace(command)
}

// checkAndWarnForPathsInEnvFile checks for path-like variables in the .env file
// and warns the user that they might need to configure hooks.
func (d *Deployer) checkAndWarnForPathsInEnvFile(releasePath string) {
	sftpClient, err := sftp.NewClient(d.SSHClient)
	if err != nil {
		log.Printf("⚠️ Failed to create SFTP client to check .env file: %v", err)
		return
	}
	defer sftpClient.Close()

	envPath := path.Join(releasePath, ".env")
	log.Printf("DEBUG: Checking for .env file at: %s", envPath)

	file, err := sftpClient.Open(envPath)
	if err != nil {
		log.Printf("DEBUG: No .env file found at %s, skipping check: %v", envPath, err)
		return // .env file might not exist, which is fine.
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pathSuffixes := []string{"_PATH", "_DIR", "_FILE", "_DB"}
	var foundPaths []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" { // Ignore comments and empty lines
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			for _, suffix := range pathSuffixes {
				if strings.HasSuffix(strings.ToUpper(key), suffix) {
					foundPaths = append(foundPaths, key)
					break
				}
			}
		}
	}

	if len(foundPaths) > 0 {
		log.Printf("⚠️ Found environment variables that might require permission settings: %s", strings.Join(foundPaths, ", "))
		log.Println("Suggestion: If these paths require special permissions or initialization, please add a 'pre_deploy' hook in 'shipyard.toml'.")
		log.Println(`
  Example:
  [[hooks.pre_deploy]]
    name = "setup_my_app_dirs"
    command = "mkdir -p /data/my_app/db && chown -R app_user:app_user /data/my_app"`)
	}
}

// runHooks executes the defined hooks for a given stage.
func (d *Deployer) runHooks(stageName string, hooks []config.Hook) error {
	if len(hooks) == 0 {
		log.Printf("Stage '%s' has no configured hooks, skipping.", stageName)
		return nil
	}

	log.Printf("--- Starting execution of %s stage hooks ---", stageName)
	for _, hook := range hooks {
		log.Printf("--> Executing: %s", hook.Name)

		finalCommand := d.substituteVariables(hook.Command)
		var commandToExecute string

		switch hook.Type {
		case "eval":
			releaseBinPath := path.Join(d.CurrentReleasePath, "bin", d.AppName)
			commandToExecute = fmt.Sprintf("%s eval '%s'", releaseBinPath, finalCommand)
		case "shell":
			envFile := fmt.Sprintf("/etc/%s/env", d.AppName)
			// Source env file and execute command in sub-shell to ensure variables are passed
			// Prepend space to command to avoid recording in bash history (relies on HISTCONTROL=ignorespace)
			commandToExecute = fmt.Sprintf(` cd %s && (set -a; . %s; set +a; exec %s)`, d.CurrentReleasePath, envFile, finalCommand)
		default:
			return fmt.Errorf("unknown hook type: '%s'", hook.Type)
		}

		// Use true to log the output of the hook command
		if err := d.executeRemoteCommand(commandToExecute, true); err != nil {
			return fmt.Errorf("hook '%s' execution failed: %w", hook.Name, err)
		}
	}
	log.Printf("--- %s stage hooks execution completed ---", stageName)
	return nil
}
