package database

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"youfun/shipyard/internal/crypto"
	"youfun/shipyard/internal/models"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ApplicationToken represents an access token for a specific application
type ApplicationToken struct {
	ID            uuid.UUID    `db:"id"`
	ApplicationID uuid.UUID    `db:"application_id"`
	Name          string       `db:"name"`
	TokenHash     string       `db:"token_hash"`
	ExpiresAt     NullableTime `db:"expires_at"`
	LastUsedAt    NullableTime `db:"last_used_at"`
	IsActive      bool         `db:"is_active"`
	CreatedAt     NullableTime `db:"created_at"`
	UpdatedAt     NullableTime `db:"updated_at"`
}

// NullableTime is an alias for models.NullableTime to avoid circular import
type NullableTime = models.NullableTime

// GenerateRandomToken generates a random token string
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CreateApplicationToken creates a new application token
func CreateApplicationToken(applicationID uuid.UUID, name string, expiresAt *time.Time) (*ApplicationToken, string, error) {
	if name == "" {
		return nil, "", fmt.Errorf("name is required")
	}

	// Generate random token
	token, err := GenerateRandomToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Encrypt token
	encryptedToken, err := crypto.Encrypt(token)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	now := time.Now()
	appToken := &ApplicationToken{
		ID:            uuid.New(),
		ApplicationID: applicationID,
		Name:          name,
		TokenHash:     encryptedToken,
		IsActive:      true,
		CreatedAt:     NullableTime{Time: &now},
		UpdatedAt:     NullableTime{Time: &now},
	}

	if expiresAt != nil {
		appToken.ExpiresAt = NullableTime{Time: expiresAt}
	}

	query := Rebind(`INSERT INTO application_tokens (id, application_id, name, token_hash, expires_at, is_active, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err = DB.Exec(query, appToken.ID, appToken.ApplicationID, appToken.Name, appToken.TokenHash,
		appToken.ExpiresAt.Time, appToken.IsActive, now, now)
	if err != nil {
		return nil, "", err
	}

	return appToken, token, nil
}

// GetApplicationTokensByAppID retrieves all active application tokens for an app (without decrypted tokens)
func GetApplicationTokensByAppID(applicationID uuid.UUID) ([]ApplicationToken, error) {
	var tokens []ApplicationToken
	query := Rebind(`SELECT id, application_id, name, token_hash, expires_at, last_used_at, is_active, created_at, updated_at 
		FROM application_tokens WHERE application_id = ? AND is_active = ? ORDER BY created_at DESC`)
	err := DB.Select(&tokens, query, applicationID, true)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetApplicationTokenByID retrieves an application token by its ID
func GetApplicationTokenByID(tokenID uuid.UUID, applicationID uuid.UUID) (*ApplicationToken, error) {
	var token ApplicationToken
	query := Rebind(`SELECT id, application_id, name, token_hash, expires_at, last_used_at, is_active, created_at, updated_at 
		FROM application_tokens WHERE id = ? AND application_id = ? AND is_active = ?`)
	err := DB.Get(&token, query, tokenID, applicationID, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, err
	}
	return &token, nil
}

// ValidateApplicationToken validates a token and returns the associated application
func ValidateApplicationToken(tokenString string) (*models.Application, *ApplicationToken, error) {
	// Get all active tokens
	var tokens []ApplicationToken
	query := Rebind(`SELECT id, application_id, name, token_hash, expires_at, last_used_at, is_active, created_at, updated_at 
		FROM application_tokens WHERE is_active = ?`)
	if err := DB.Select(&tokens, query, true); err != nil {
		return nil, nil, fmt.Errorf("failed to query tokens: %w", err)
	}

	// Try to match token
	for _, token := range tokens {
		// Check if expired
		if token.ExpiresAt.Time != nil && token.ExpiresAt.Time.Before(time.Now()) {
			continue
		}

		// Decrypt and compare token
		decryptedToken, err := crypto.Decrypt(token.TokenHash)
		if err != nil {
			continue // Decryption failed, skip
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(decryptedToken), []byte(tokenString)) == 1 {
			// Found matching token, update last used time
			_ = UpdateAppTokenLastUsed(token.ID)

			// Get associated application
			app, err := GetApplicationByID(token.ApplicationID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get application: %w", err)
			}

			return app, &token, nil
		}
	}

	return nil, nil, fmt.Errorf("invalid token")
}

// UpdateApplicationToken updates an application token (excluding the token itself)
func UpdateApplicationToken(tokenID uuid.UUID, applicationID uuid.UUID, name string, expiresAt *time.Time) error {
	now := time.Now()
	query := Rebind(`UPDATE application_tokens SET name = ?, expires_at = ?, updated_at = ? 
		WHERE id = ? AND application_id = ?`)
	_, err := DB.Exec(query, name, expiresAt, now, tokenID, applicationID)
	return err
}

// DeleteApplicationToken soft deletes an application token
func DeleteApplicationToken(tokenID uuid.UUID, applicationID uuid.UUID) error {
	query := Rebind(`UPDATE application_tokens SET is_active = ?, updated_at = ? WHERE id = ? AND application_id = ?`)
	_, err := DB.Exec(query, false, time.Now(), tokenID, applicationID)
	return err
}

// UpdateAppTokenLastUsed updates the last used timestamp for a token
func UpdateAppTokenLastUsed(tokenID uuid.UUID) error {
	now := time.Now()
	query := Rebind(`UPDATE application_tokens SET last_used_at = ? WHERE id = ?`)
	_, err := DB.Exec(query, now, tokenID)
	return err
}
