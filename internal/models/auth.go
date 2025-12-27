package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthDevice represents a device or session that a user has used to log in.
// This is used for security auditing and allows users to revoke specific sessions/devices.
type AuthDevice struct {
	ID         uuid.UUID `db:"id"`
	UserID     uuid.UUID `db:"user_id"`
	IPAddress  string    `db:"ip_address"`
	UserAgent  string    `db:"user_agent"`
	LastUsedAt time.Time `db:"last_used_at"`
	CreatedAt  time.Time `db:"created_at"`
}

// TwoFactorRecoveryCode stores recovery codes for 2FA.
// Tags are for sqlx.
type TwoFactorRecoveryCode struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Code      string    `db:"code"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// User represents a user account in the system.
// Tags are for sqlx, not gorm.
type User struct {
	ID               uuid.UUID `db:"id"`
	Username         string    `db:"username"`
	Password         string    `db:"password"`
	TwoFactorSecret  string    `db:"two_factor_secret"`
	TwoFactorEnabled bool      `db:"two_factor_enabled"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// HashPassword hashes the password using bcrypt.
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the password against the hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
