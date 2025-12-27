package sshutil

import (
	"bufio"
	"bytes"
	"youfun/shipyard/internal/depsinstall"
	"youfun/shipyard/internal/models"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// HostKeyVerifier handles host key verification with optional user confirmation.
type HostKeyVerifier struct {
	TrustedKey string // Base64 encoded trusted key
	// Confirm is called when the host key is unknown (TrustedKey is empty).
	// It should return true if the user trusts the key.
	Confirm func(hostname string, remote net.Addr, key ssh.PublicKey) bool
	// CapturedKey stores the base64 encoded key presented by the server.
	CapturedKey string
}

// Callback implements ssh.HostKeyCallback.
func (v *HostKeyVerifier) Callback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	marshaledKey := key.Marshal()
	encodedKey := base64.StdEncoding.EncodeToString(marshaledKey)
	v.CapturedKey = encodedKey

	if v.TrustedKey == "" {
		// New host or no trusted key recorded yet
		if v.Confirm != nil {
			if v.Confirm(hostname, remote, key) {
				return nil
			}
			return errors.New("host key rejected by user")
		}
		// If no confirm callback and no trusted key, fail securely
		return errors.New("unknown host key and no confirmation mechanism provided")
	}

	// Verify against trusted key
	trustedBytes, err := base64.StdEncoding.DecodeString(v.TrustedKey)
	if err != nil {
		return fmt.Errorf("invalid trusted key format: %w", err)
	}

	if !bytes.Equal(trustedBytes, marshaledKey) {
		return fmt.Errorf("REMOTE HOST IDENTIFICATION HAS CHANGED! Host key mismatch. Expected fingerprint: %s, Got: %s",
			ssh.FingerprintSHA256(tryParseKey(trustedBytes)),
			ssh.FingerprintSHA256(key))
	}
	return nil
}

func tryParseKey(b []byte) ssh.PublicKey {
	k, _ := ssh.ParsePublicKey(b)
	return k
}

// ExecuteRemoteCommand executes a command on the remote host and returns its combined output.
func ExecuteRemoteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("failed to run command '%s': %w", command, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// NewClientConfig creates an ssh.ClientConfig from a host model.
// It accepts an optional HostKeyCallback. If nil, it uses a default verifier that checks against host.HostKey.
func NewClientConfig(host *models.SSHHost, hostKeyCallback ssh.HostKeyCallback) (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	if host.PrivateKey != nil && *host.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(*host.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key for host %s: %w", host.Name, err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if host.Password != nil && *host.Password != "" {
		authMethods = append(authMethods, ssh.Password(*host.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method available for host %s", host.Name)
	}

	// Default to strict checking against the stored key if no callback provided
	if hostKeyCallback == nil {
		verifier := &HostKeyVerifier{
			TrustedKey: "",
		}
		if host.HostKey != nil {
			verifier.TrustedKey = *host.HostKey
		}
		// Without a confirm callback, this will fail for new hosts (which is secure for automated runs)
		// unless the key is already in the DB.
		hostKeyCallback = verifier.Callback
	}

	config := &ssh.ClientConfig{
		User:            host.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	return config, nil
}

// runCommand runs a command on the remote host and streams its output.
func runCommand(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// We use a pipe to get the combined output, which is simpler for streaming.
	outPipe, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	errPipe, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	log.Printf("Running command: %s\n", cmd)
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Create a combined reader
	go io.Copy(os.Stdout, outPipe)
	go io.Copy(os.Stderr, errPipe)

	if err := session.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// CheckAndInstallCaddy checks for Caddy and installs it if missing.
func CheckAndInstallCaddy(client *ssh.Client) error {
	// Prompt user for confirmation
	fmt.Print("Do you want to automatically install the latest version of Caddy? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	autoInstall := (input == "y" || input == "yes")

	if !autoInstall {
		log.Println("You chose not to install Caddy. Please note that deploying applications requires Caddy as a reverse proxy.")
		return nil
	}

	return depsinstall.CheckAndInstallCaddyRemote(client, true)
}
