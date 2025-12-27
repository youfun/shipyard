package cliutils

import (
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"encoding/base64"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

// GetInteractiveHostKeyCallback returns a HostKeyCallback that prompts the user
// via stdin/stdout to verify unknown host keys. If confirmed, it saves the key
// to the database using the host ID.
func GetInteractiveHostKeyCallback(host *models.SSHHost) ssh.HostKeyCallback {
	verifier := &sshutil.HostKeyVerifier{
		TrustedKey: "",
		Confirm: func(hostname string, remote net.Addr, key ssh.PublicKey) bool {
			fmt.Printf("\n⚠️  UNKNOWN HOST KEY for %s (%s)\n", hostname, remote)
			fmt.Printf("Fingerprint: %s\n", ssh.FingerprintSHA256(key))
			fmt.Println("This is the first time you have connected to this host (or the key has changed).")
			fmt.Println("Please verify that the fingerprint matches the host's key.")

			if PromptForConfirmation("Do you want to trust this host and save the key?", false) {
				// Save to database immediately
				encodedKey := base64.StdEncoding.EncodeToString(key.Marshal())
				if err := database.UpdateHostKey(host.ID, encodedKey); err != nil {
					fmt.Printf("Warning: failed to save host key to database: %v\n", err)
				} else {
					fmt.Println("✅ Host key saved to database.")
				}
				return true
			}
			return false
		},
	}
	if host.HostKey != nil {
		verifier.TrustedKey = *host.HostKey
	}
	return verifier.Callback
}
