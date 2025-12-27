package deploy

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"youfun/shipyard/internal/depsinstall"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"youfun/shipyard/internal/static"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// InitializeHost prepares a remote host for deployments.
func InitializeHost(host *models.SSHHost, appName, runtime, startCmd, user string, hostKeyCallback ssh.HostKeyCallback) error {
	// Use embedded script
	scriptBytes := []byte(static.InitRuntimeScript)

	// Convert CRLF to LF for Unix-like systems
	scriptBytes = bytes.ReplaceAll(scriptBytes, []byte("\r\n"), []byte("\n"))
	b64 := base64.StdEncoding.EncodeToString(scriptBytes)

	// Connect to remote and execute
	sshConfig, err := sshutil.NewClientConfig(host, hostKeyCallback)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", host.Addr, host.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	// Check and install Caddy if needed
	if err := installCaddyIfNeeded(client); err != nil {
		return fmt.Errorf("failed to install Caddy: %w", err)
	}

	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer sess.Close()

	remoteCmd := fmt.Sprintf("bash -c \"mkdir -p /tmp; echo '%s' | base64 -d > /tmp/init_runtime.sh; chmod +x /tmp/init_runtime.sh; APP=%s USER=%s RUNTIME=%s START_CMD=\\\"%s\\\" bash /tmp/init_runtime.sh\"", b64, appName, user, runtime, strings.ReplaceAll(startCmd, "\"", "\\\""))
	log.Printf("🚀 Executing remote initialization: %s@%s runtime=%s user=%s app=%s", host.User, host.Addr, runtime, user, appName)
	out, err := sess.CombinedOutput(remoteCmd)
	fmt.Print(string(out))
	if err != nil {
		return fmt.Errorf("remote initialization failed: %w", err)
	}

	log.Println("✅ Remote initialization completed.")
	return nil
}

// installCaddyIfNeeded checks if Caddy is installed on the remote host and installs it if not.
func installCaddyIfNeeded(client *ssh.Client) error {
	return depsinstall.CheckAndInstallCaddyRemote(client, true)
}

/**
 * @brief Upload and untar a tarball via streaming (piping) with progress bar
 *
 * Streams a local tar.gz file to the remote host and pipes it into 'tar -xzf -'
 * so upload and extraction happen concurrently. Shows a progress bar during upload.
 *
 * @param localTarballPath  local .tar.gz file path (e.g., "./build/my_app.tar.gz")
 * @param remoteReleasePath remote directory to extract into (e.g., "/opt/app/releases/20251108")
 * @return error            non-nil on failure, nil on success
 */
func (d *Deployer) uploadTarFile(localTarballPath string, remoteReleasePath string) error {

	// 1. Open local tarball
	localFile, err := os.Open(localTarballPath)
	if err != nil {
		return fmt.Errorf("failed to open local tarball (%s): %w", localTarballPath, err)
	}
	defer localFile.Close()

	// 2. Get file size for progress bar
	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	fileSize := stat.Size()

	// 3. Create SSH session
	session, err := d.SSHClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// 4. [progress integration]
	log.Println("--- Starting streaming upload and untar ---")
	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true)
	// Create a proxy reader so we can track progress
	barReader := bar.NewProxyReader(localFile)

	// 5. Attach the progress reader to session stdin
	session.Stdin = barReader

	// 6. Prepare remote command
	remoteCmd := fmt.Sprintf("mkdir -p %s && tar -xzf - -C %s",
		remoteReleasePath,
		remoteReleasePath,
	)

	// 7. Run remote command (session.Run pumps session.Stdin)
	if err := session.Run(remoteCmd); err != nil {
		bar.Finish()
		return fmt.Errorf("remote streaming untar failed: %w", err)
	}

	// 8. Finish progress bar
	bar.Finish()

	log.Printf("✅ Streaming upload and untar succeeded: %s -> %s\n", localTarballPath, remoteReleasePath)
	return nil
}

// executeRemoteCommand executes a command on the remote host.
func (d *Deployer) executeRemoteCommand(command string, logOutput bool) error {
	if d.IsLocalhost {
		return d.executeLocalCommand(command, logOutput)
	}

	session, err := d.SSHClient.NewSession()
	if err != nil {
		log.Printf("failed to create SSH session: %v", err)
		return err
	}
	defer session.Close()

	// log.Printf("🚀 Executing remote command")
	// log.Printf("🚀 Executing remote command: %s", command)
	output, err := session.CombinedOutput(command)

	if logOutput && len(output) > 0 {
		log.Println(strings.TrimSpace(string(output)))
	}

	if err != nil {
		log.Printf("❌ command execution failed: %v\nOutput: %s", err, string(output))
		return err
	}
	return nil
}

// executeRemoteCommandWithOutput executes a command on the remote host and returns the output.
func (d *Deployer) executeRemoteCommandWithOutput(command string) (string, error) {
	if d.IsLocalhost {
		return d.executeLocalCommandWithOutput(command)
	}

	session, err := d.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// log.Printf("🚀 Executing remote command (capture output): %s", command)
	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// executeLocalCommand executes a command locally
func (d *Deployer) executeLocalCommand(command string, logOutput bool) error {
	// log.Printf("🚀 Executing local command: %s", command)

	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()

	if logOutput && len(output) > 0 {
		log.Println(strings.TrimSpace(string(output)))
	}

	if err != nil {
		log.Printf("❌ command execution failed: %v\nOutput: %s", err, string(output))
		return err
	}
	return nil
}

// executeLocalCommandWithOutput executes a command locally and returns output
func (d *Deployer) executeLocalCommandWithOutput(command string) (string, error) {
	log.Printf("🚀 Executing local command (capture output): %s", command)

	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// uploadFile uploads a file to the remote host using SFTP or copies locally.
func (d *Deployer) uploadFile(localPath, remoteDir string) error {
	if d.IsLocalhost {
		return d.copyFileLocally(localPath, remoteDir)
	}

	sftpClient, err := sftp.NewClient(d.SSHClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file size for progress bar
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	remotePath := filepath.Join(remoteDir, filepath.Base(localPath))
	log.Printf("DEBUG: SFTP remotePath: %s", remotePath)
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create file on remote server: %w", err)
	}
	defer remoteFile.Close()

	// Create and start a new progress bar
	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true) // Show progress in bytes
	// Create a proxy reader so we can track progress
	barReader := bar.NewProxyReader(localFile)

	// Copy using the proxy reader
	if _, err = io.Copy(remoteFile, barReader); err != nil {
		return fmt.Errorf("error occurred during file upload: %w", err)
	}

	// Finish the progress bar
	bar.Finish()

	log.Printf("✅ file upload succeeded: %s -> %s", localPath, remotePath)
	return nil
}

// copyFileLocally copies a file to a local directory
func (d *Deployer) copyFileLocally(localPath, targetDir string) error {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	targetPath := filepath.Join(targetDir, filepath.Base(localPath))
	log.Printf("📁 Copying file: %s -> %s", localPath, targetPath)

	sourceFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer destFile.Close()

	// Get file size for progress bar
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Copy with progress bar
	bar := pb.Full.Start64(fileInfo.Size())
	bar.Set(pb.Bytes, true)
	barReader := bar.NewProxyReader(sourceFile)

	if _, err = io.Copy(destFile, barReader); err != nil {
		return fmt.Errorf("file copy failed: %w", err)
	}

	bar.Finish()
	log.Printf("✅ file copy succeeded: %s", targetPath)
	return nil
}

// findFreePort finds a free port on the remote host.
func (d *Deployer) findFreePort() (int, error) {
	for range 100 { // try up to 100 times
		port := rand.Intn(10000) + 10000 // 10000-19999 range
		output, err := d.executeRemoteCommandWithOutput(fmt.Sprintf("ss -lntu | grep :%d", port))
		// We expect an error if grep finds nothing, so we check the output.
		if err != nil && !strings.Contains(output, ":"+fmt.Sprint(port)) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("unable to find a free port on remote host")
}

// healthCheck performs a health check on the service running on the given port.
func (d *Deployer) healthCheck(port int) error {
	log.Printf("Performing service health check (port: %d)", port)
	// Use systemctl is-active to check if the service is active and running.
	healthCheckCmd := fmt.Sprintf("systemctl is-active --quiet %s@%d", d.AppName, port)

	for i := 0; i < 10; i++ {
		if err := d.executeRemoteCommand(healthCheckCmd, false); err == nil {
			log.Println("✅ Service health check passed (systemd service is active).")
			return nil
		}

		log.Printf("Health check attempt %d/10 failed, retrying after 2s...", i+1)
		time.Sleep(2 * time.Second)
	}

	// After multiple attempts, dump unit status for debugging
	log.Println("Last health check attempt failed, dumping systemd unit status:")
	debugCmd := fmt.Sprintf("systemctl status %s@%d", d.AppName, port)
	_ = d.executeRemoteCommand(debugCmd, true) // run once to log output

	return fmt.Errorf("service did not reach 'active' state after multiple attempts")
}

// killProcessOnPort kills the process running on the given port.
func (d *Deployer) killProcessOnPort(port int) {
	log.Printf("Stopping old process running on port %d...", port)
	// Use lsof to find and kill the process
	pid, err := d.executeRemoteCommandWithOutput(fmt.Sprintf("lsof -t -i:%d", port))
	if err != nil || pid == "" {
		log.Printf("No running process found on port %d or command failed.", port)
		return
	}

	if err := d.executeRemoteCommand(fmt.Sprintf("kill %s", pid), true); err != nil {
		log.Printf("Failed to send kill signal to process ID %s: %v", pid, err)
	} else {
		log.Printf("Sent kill signal to process ID: %s", pid)
	}
}
