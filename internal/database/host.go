package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"youfun/shipyard/internal/crypto"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// --- ssh_hosts Table Operations ---

// GetAllSSHHosts retrieves all SSH hosts from the database, decrypting credentials for each.
func GetAllSSHHosts() ([]models.SSHHost, error) {
	var hosts []models.SSHHost
	query := `
		SELECT 
			id, name, addr, port, "user", password, private_key, host_key,
			COALESCE(status, '') as status, 
			COALESCE(arch, '') as arch, 
			initialized_at, created_at, updated_at 
		FROM ssh_hosts 
		ORDER BY name ASC
	`
	err := DB.Select(&hosts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all SSH hosts: %w", err)
	}

	// Decrypt credentials for each host
	for i := range hosts {
		host := &hosts[i] // Use a pointer to modify the item in the slice
		if host.Password != nil && *host.Password != "" {
			decryptedPassword, err := crypto.Decrypt(*host.Password)
			if err != nil {
				// Log the error but don't fail the whole operation, maybe one key is corrupted
				log.Printf("Warning: failed to decrypt password for host '%s': %v", host.Name, err)
				// Optionally, set password to nil or an empty string
				host.Password = nil
			} else {
				host.Password = &decryptedPassword
			}
		}

		if host.PrivateKey != nil && *host.PrivateKey != "" {
			decryptedKey, err := crypto.Decrypt(*host.PrivateKey)
			if err != nil {
				log.Printf("Warning: failed to decrypt private key for host '%s': %v", host.Name, err)
				host.PrivateKey = nil
			} else {
				host.PrivateKey = &decryptedKey
			}
		}
	}

	return hosts, nil
}

// AddSSHHost adds a new SSH host to the database, encrypting credentials.
func AddSSHHost(host *models.SSHHost) error {
	host.ID = uuid.New() // Generate UUID for the new host
	var encryptedPassword, encryptedKey interface{}
	var err error

	if host.Password != nil && *host.Password != "" {
		encPass, err := crypto.Encrypt(*host.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		encryptedPassword = encPass
	}

	if host.PrivateKey != nil && *host.PrivateKey != "" {
		encKey, err := crypto.Encrypt(*host.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}
		encryptedKey = encKey
	}

	query := Rebind(`INSERT INTO ssh_hosts (id, name, addr, port, "user", password, private_key, host_key, status, arch) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err = DB.Exec(query, host.ID, host.Name, host.Addr, host.Port, host.User, encryptedPassword, encryptedKey, host.HostKey, "healthy", host.Arch)
	return err
}

// GetSSHHostByName retrieves a single SSH host by its name, decrypting credentials.
func GetSSHHostByName(name string) (*models.SSHHost, error) {
	var host models.SSHHost
	query := Rebind(`
		SELECT 
			id, name, addr, port, "user", password, private_key, host_key,
			COALESCE(status, '') as status, 
			COALESCE(arch, '') as arch, 
			initialized_at, created_at, updated_at 
		FROM ssh_hosts 
		WHERE name = ?
	`)
	err := DB.Get(&host, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHostNotFound
		}
		return nil, err
	}

	if host.Password != nil && *host.Password != "" {
		decryptedPassword, err := crypto.Decrypt(*host.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		host.Password = &decryptedPassword
	}

	if host.PrivateKey != nil && *host.PrivateKey != "" {
		decryptedKey, err := crypto.Decrypt(*host.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
		host.PrivateKey = &decryptedKey
	}

	return &host, nil
}

// GetHostByID retrieves a single SSH host by its ID, decrypting credentials.
func GetHostByID(id uuid.UUID) (*models.SSHHost, error) {
	var host models.SSHHost
	query := Rebind(`
		SELECT 
			id, name, addr, port, "user", password, private_key, host_key,
			COALESCE(status, '') as status, 
			COALESCE(arch, '') as arch, 
			initialized_at, created_at, updated_at 
		FROM ssh_hosts 
		WHERE id = ?
	`)
	err := DB.Get(&host, query, id)
	if err != nil {
		return nil, err
	}

	if host.Password != nil && *host.Password != "" {
		decryptedPassword, err := crypto.Decrypt(*host.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		host.Password = &decryptedPassword
	}

	if host.PrivateKey != nil && *host.PrivateKey != "" {
		decryptedKey, err := crypto.Decrypt(*host.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
		host.PrivateKey = &decryptedKey
	}

	return &host, nil
}

// SetHostInitialized marks a host as initialized by setting the initialized_at timestamp.
func SetHostInitialized(hostID uuid.UUID) error {
	query := Rebind(`UPDATE ssh_hosts SET initialized_at = ? WHERE id = ?`)
	_, err := DB.Exec(query, time.Now(), hostID)
	return err
}

// UpdateHostKey updates the known host key for an SSH host.
func UpdateHostKey(hostID uuid.UUID, hostKey string) error {
	query := Rebind(`UPDATE ssh_hosts SET host_key = ?, updated_at = ? WHERE id = ?`)
	_, err := DB.Exec(query, hostKey, time.Now(), hostID)
	return err
}
