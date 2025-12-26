package logs

import (
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// connectSSH establishes an SSH connection to the remote host
func connectSSH(host *models.SSHHost, hostKeyCallback ssh.HostKeyCallback) (*ssh.Client, error) {
	sshConfig, err := sshutil.NewClientConfig(host, hostKeyCallback)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH config: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", host.Addr, host.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH connection to %s failed: %w", addr, err)
	}

	return client, nil
}

// FetchJournalLogs connects to the remote host via SSH, executes the journalctl command, and returns the log.
// Parameters:
//   - host: SSH host info
//   - appName: Application name (used to build systemd unit name)
//   - port: Application running port
//   - lines: Show the last N lines of logs (default 500)
//   - follow: Whether to follow logs in real-time (-f mode)
//   - hostKeyCallback: SSH host key verification callback
//
// Returns: Log content string or error
func FetchJournalLogs(host *models.SSHHost, appName string, port int, lines int, follow bool, hostKeyCallback ssh.HostKeyCallback) (string, error) {
	// Establish SSH connection
	client, err := connectSSH(host, hostKeyCallback)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Build journalctl command
	// Use <app-name>@<port> format for systemd template unit
	unitName := fmt.Sprintf("%s@%d", appName, port)
	cmd := fmt.Sprintf("journalctl -u %s -n %d --no-pager -o cat", unitName, lines)

	if follow {
		cmd += " -f"
	}

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Execute command and get output
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("journalctl command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// StreamJournalLogs streams logs to stdout in real-time
// Used for -f/--follow mode
func StreamJournalLogs(host *models.SSHHost, appName string, port int, lines int, hostKeyCallback ssh.HostKeyCallback) error {
	// Establish SSH connection
	client, err := connectSSH(host, hostKeyCallback)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Build journalctl command with -f
	unitName := fmt.Sprintf("%s@%d", appName, port)
	cmd := fmt.Sprintf("journalctl -u %s -n %d --no-pager -o cat -f", unitName, lines)

	// Set stdout and stderr
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Set Ctrl+C signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start command in a separate goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Run(cmd)
	}()

	// Wait for command completion or interrupt signal
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("journalctl command failed: %w", err)
		}
		return nil
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal, exiting...")
		// Send signal to remote session
		session.Signal(ssh.SIGINT)
		return nil
	}
}

// BuildJournalctlCommand builds the journalctl command string (for testing)
func BuildJournalctlCommand(appName string, port int, lines int, follow bool) string {
	unitName := fmt.Sprintf("%s@%d", appName, port)
	cmd := fmt.Sprintf("journalctl -u %s -n %s --no-pager -o cat", unitName, strconv.Itoa(lines))
	if follow {
		cmd += " -f"
	}
	return cmd
}
