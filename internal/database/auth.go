package database

import (
	"database/sql"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// --- User Operations ---

// CreateUser creates a new user with a hashed password.
// It manually sets the UUID and timestamps.
func CreateUser(tx *sqlx.Tx, username, password string) (*models.User, error) {
	user := &models.User{
		ID:       uuid.New(),
		Username: username,
	}

	if err := user.HashPassword(password); err != nil {
		return nil, err
	}

	now := time.Now()
	query := Rebind(`
		INSERT INTO users (id, username, password, two_factor_secret, two_factor_enabled, created_at, updated_at)
		VALUES (?, ?, ?, '', ?, ?, ?)
	`)
	// Using a transaction passed from the caller
	_, err := tx.Exec(query, user.ID, user.Username, user.Password, false, now, now)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by their username.
func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	query := Rebind("SELECT * FROM users WHERE username = ?")
	err := DB.Get(&user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID.
func GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := Rebind("SELECT * FROM users WHERE id = ?")
	err := DB.Get(&user, query, id.String())
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserCount returns the number of users in the database.
func GetUserCount() (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM users"
	err := DB.Get(&count, query)
	return count, err
}

// UpdateUserPassword updates a user's password.
func UpdateUserPassword(id uuid.UUID, newPassword string) error {
	user := &models.User{ID: id}
	if err := user.HashPassword(newPassword); err != nil {
		return err
	}

	now := time.Now()
	query := Rebind("UPDATE users SET password = ?, updated_at = ? WHERE id = ?")
	_, err := DB.Exec(query, user.Password, now, user.ID)
	return err
}

// Enable2FAForUser enables 2FA for a user and stores recovery codes in a transaction.
func Enable2FAForUser(tx *sqlx.Tx, userID uuid.UUID, secret string, hashedCodes []string) error {
	// 1. Update the user record
	userQuery := Rebind("UPDATE users SET two_factor_enabled = ?, two_factor_secret = ? WHERE id = ?")
	_, err := tx.Exec(userQuery, true, secret, userID)
	if err != nil {
		return err
	}

	// 2. Create recovery codes
	if len(hashedCodes) > 0 {
		now := time.Now()
		codeQuery := Rebind("INSERT INTO two_factor_recovery_codes (id, user_id, code, used, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)")
		for _, code := range hashedCodes {
			_, err := tx.Exec(codeQuery, uuid.New(), userID, code, false, now, now)
			if err != nil {
				return err // The transaction will be rolled back
			}
		}
	}

	return nil
}

// Disable2FAForUser disables 2FA for a user and deletes their recovery codes in a transaction.
func Disable2FAForUser(tx *sqlx.Tx, userID uuid.UUID) error {
	// 1. Update the user record
	userQuery := Rebind("UPDATE users SET two_factor_enabled = ?, two_factor_secret = '' WHERE id = ?")
	_, err := tx.Exec(userQuery, false, userID)
	if err != nil {
		return err
	}

	// 2. Delete all recovery codes for the user
	recoveryQuery := Rebind("DELETE FROM two_factor_recovery_codes WHERE user_id = ?")
	_, err = tx.Exec(recoveryQuery, userID)
	if err != nil {
		return err
	}

	return nil
}

// GetFirstUser is a convenience function, typically for checking if setup is needed.
func GetFirstUser() (*models.User, error) {
	var user models.User
	query := "SELECT * FROM users ORDER BY created_at ASC LIMIT 1"
	err := DB.Get(&user, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if no users exist, not an error
		}
		return nil, err
	}
	return &user, nil
}

// --- AuthDevice Operations ---

// CreateAuthDevice records a new device/session for a user upon successful login.
func CreateAuthDevice(tx *sqlx.Tx, userID uuid.UUID, ipAddress, userAgent string) (*models.AuthDevice, error) {
	device := &models.AuthDevice{
		ID:        uuid.New(),
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	now := time.Now()
	query := Rebind(`
		INSERT INTO auth_devices (id, user_id, ip_address, user_agent, last_used_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)

	// Use the provided transaction
	_, err := tx.Exec(query, device.ID, device.UserID, device.IPAddress, device.UserAgent, now, now)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// ListAuthDevicesForUser retrieves all active devices/sessions for a given user.
func ListAuthDevicesForUser(userID uuid.UUID) ([]*models.AuthDevice, error) {
	var devices []*models.AuthDevice
	query := Rebind("SELECT * FROM auth_devices WHERE user_id = ? ORDER BY last_used_at DESC")
	err := DB.Select(&devices, query, userID)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// GetAuthDeviceByID retrieves a single device by its ID.
func GetAuthDeviceByID(deviceID uuid.UUID) (*models.AuthDevice, error) {
	var device models.AuthDevice
	query := Rebind("SELECT * FROM auth_devices WHERE id = ?")
	err := DB.Get(&device, query, deviceID)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// RevokeAuthDevice removes a device record, effectively logging it out.
// The JWT related to this device will eventually expire, but this provides immediate revocation
// if the JWT contains a session ID (`jti`) that can be checked against this table.
func RevokeAuthDevice(userID, deviceID uuid.UUID) error {
	// The userID is included to ensure a user can only revoke their own devices.
	query := Rebind("DELETE FROM auth_devices WHERE id = ? AND user_id = ?")
	result, err := DB.Exec(query, deviceID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		// This can happen if the device does not exist or does not belong to the user.
		// Depending on the desired security policy, you might want to return an error here.
		// For now, we treat it as a non-error.
	}
	return err
}

// UpdateAuthDeviceLastUsed updates the timestamp for when a device was last used.
// This can be called periodically by the authentication middleware.
func UpdateAuthDeviceLastUsed(deviceID uuid.UUID) error {
	now := time.Now()
	query := Rebind("UPDATE auth_devices SET last_used_at = ? WHERE id = ?")
	_, err := DB.Exec(query, now, deviceID)
	return err
}
