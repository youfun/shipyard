package database

import (
	"fmt"
	"log"
	"youfun/shipyard/internal/crypto"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// --- secrets Table Operations ---

// SetSecret adds or updates a secret.
func SetSecret(appID uuid.UUID, key, value string) error {
	encryptedValue, err := crypto.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	now := time.Now()

	// ON CONFLICT with EXCLUDED works across SQLite, TURSO, and PostgreSQL
	// SQLite is case-insensitive, PostgreSQL requires EXCLUDED in the alias
	query := Rebind(`
	INSERT INTO secrets (id, application_id, key, value, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(application_id, key) DO UPDATE SET
	value = EXCLUDED.value,
	updated_at = EXCLUDED.updated_at;
	`)
	_, err = DB.Exec(query, uuid.New(), appID, key, encryptedValue, now, now)
	return err
}

// UnsetSecret deletes a secret.
func UnsetSecret(appID uuid.UUID, key string) error {
	query := Rebind("DELETE FROM secrets WHERE application_id = ? AND key = ?")
	_, err := DB.Exec(query, appID, key)
	return err
}

// ListSecretKeys lists all secret names for an application.
func ListSecretKeys(appID uuid.UUID) ([]string, error) {
	var keys []string
	query := Rebind("SELECT key FROM secrets WHERE application_id = ? ORDER BY key ASC")
	err := DB.Select(&keys, query, appID)
	return keys, err
}

// GetSecretsForApp retrieves all decrypted secrets for an application.
func GetSecretsForApp(appID uuid.UUID) (map[string]string, error) {
	var secrets []models.Secret
	query := Rebind("SELECT * FROM secrets WHERE application_id = ?")
	err := DB.Select(&secrets, query, appID)
	if err != nil {
		return nil, err
	}

	decryptedSecrets := make(map[string]string)
	for _, s := range secrets {
		decryptedValue, err := crypto.Decrypt(s.Value)
		if err != nil {
			// If a single secret decryption fails, we can choose to skip or return an error
			log.Printf("Warning: failed to decrypt secret '%s', skipping. Error: %v", s.Key, err)
			continue
		}
		decryptedSecrets[s.Key] = decryptedValue
	}

	return decryptedSecrets, nil
}
