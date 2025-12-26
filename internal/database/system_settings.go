package database

import (
	"database/sql"
	"time"
)

// SystemSetting represents a system-wide configuration
type SystemSetting struct {
	Key       string    `db:"key"`
	Value     string    `db:"value"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetSystemSetting retrieves a system setting by key.
// Returns an empty string and nil error if the key doesn't exist.
func GetSystemSetting(key string) (string, error) {
	var value string
	query := Rebind(`SELECT value FROM system_settings WHERE key = ?`)
	err := DB.Get(&value, query, key)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSystemSetting creates or updates a system setting.
func SetSystemSetting(key, value string) error {
	query := Rebind(`
		INSERT INTO system_settings (key, value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP) 
		ON CONFLICT(key) DO UPDATE SET 
		value = excluded.value, 
		updated_at = CURRENT_TIMESTAMP
	`)
	_, err := DB.Exec(query, key, value)
	return err
}
