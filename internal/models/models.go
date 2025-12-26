package models

import (
	"database/sql"
	"database/sql/driver"
	"youfun/shipyard/internal/crypto"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NullableTime is a wrapper around *time.Time that handles database scanning
// from string, time.Time, or nil values.
type NullableTime struct {
	*time.Time
}

// Scan implements the sql.Scanner interface.
// It can scan from a string (in several formats), a time.Time object, or nil.
func (nt *NullableTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time = nil
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		nt.Time = &v
		return nil
	case []byte:
		return nt.parseString(string(v))
	case string:
		return nt.parseString(v)
	}

	return fmt.Errorf("unsupported type for NullableTime: %T", value)
}

func (nt *NullableTime) parseString(s string) error {
	// Check for and truncate the monotonic clock part, if present.
	if i := strings.LastIndex(s, " m="); i != -1 {
		s = s[:i]
	}

	// Turso often returns timestamps as strings.
	// We'll try to parse a few common formats.
	layouts := []string{
		"2006-01-02 15:04:05.999999999 -0700 MST", // For formats like '... +0800 CST'
		"2006-01-02 15:04:05.999999999-07:00",     // Turso Pro
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05", // Common format
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			nt.Time = &t
			return nil
		}
	}
	return fmt.Errorf("could not parse time string '%s' with any supported layout", s)
}

// Value implements the driver.Valuer interface.
func (nt NullableTime) Value() (driver.Value, error) {
	if nt.Time == nil {
		return nil, nil
	}
	return *nt.Time, nil
}

// SSHHost represents a remote server connection details.
type SSHHost struct {
	ID            uuid.UUID    `db:"id"`
	Name          string       `db:"name"`
	Addr          string       `db:"addr"`
	Port          int          `db:"port"`
	User          string       `db:"user"`
	Password      *string      `db:"password"`    // Encrypted, now nullable
	PrivateKey    *string      `db:"private_key"` // Encrypted, nullable
	HostKey       *string      `db:"host_key"`    // Known host key (authorized_keys format or base64 wire format)
	Status        string       `db:"status"`
	Arch          string       `db:"arch"`
	InitializedAt NullableTime `db:"initialized_at"`
	CreatedAt     NullableTime `db:"created_at"`
	UpdatedAt     NullableTime `db:"updated_at"`
}

// DecryptCredentials decrypts the password and private key of the SSHHost in-place.
func (h *SSHHost) DecryptCredentials() error {
	if h.Password != nil && *h.Password != "" {
		decryptedPassword, err := crypto.Decrypt(*h.Password)
		if err != nil {
			return fmt.Errorf("failed to decrypt password for host %s: %w", h.Name, err)
		}
		h.Password = &decryptedPassword
	}

	if h.PrivateKey != nil && *h.PrivateKey != "" {
		decryptedKey, err := crypto.Decrypt(*h.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt private key for host %s: %w", h.Name, err)
		}
		h.PrivateKey = &decryptedKey
	}
	return nil
}

// Application represents a deployable application.
type Application struct {
	ID        uuid.UUID    `db:"id"`
	Name      string       `db:"name"`
	CreatedAt NullableTime `db:"created_at"`
	UpdatedAt NullableTime `db:"updated_at"`
}

// ApplicationInstance links an Application to an SSHHost.
type ApplicationInstance struct {
	ID                 uuid.UUID     `db:"id"`
	ApplicationID      uuid.UUID     `db:"application_id"`
	HostID             uuid.UUID     `db:"host_id"`
	Status             string        `db:"status"`
	ActivePort         sql.NullInt64 `db:"active_port"`
	PreviousActivePort sql.NullInt64 `db:"previous_active_port"`
	CreatedAt          NullableTime  `db:"created_at"`
	UpdatedAt          NullableTime  `db:"updated_at"`
}

// DeploymentStatus defines the types of deployment statuses
type DeploymentStatus string

const (
	DeploymentStatusPending DeploymentStatus = "pending"
	DeploymentStatusSuccess DeploymentStatus = "success"
	DeploymentStatusFailed  DeploymentStatus = "failed"
)

// DeploymentHistory stores history record of a deployment
type DeploymentHistory struct {
	ID          uuid.UUID        `db:"id"`
	InstanceID  uuid.UUID        `db:"instance_id"`
	Version     string           `db:"version"`
	ReleasePath string           `db:"release_path"`
	Status      DeploymentStatus `db:"status"`
	LogOutput   string           `db:"log_output"`
	Port        int              `db:"port"` // Added field
	DeployedAt  NullableTime     `db:"deployed_at"`
	CreatedAt   NullableTime     `db:"created_at"`
	UpdatedAt   NullableTime     `db:"updated_at"`
}

// Secret stores an encrypted sensitive variable
type Secret struct {
	ID            uuid.UUID    `db:"id"`
	ApplicationID uuid.UUID    `db:"application_id"`
	Key           string       `db:"key"`
	Value         string       `db:"value"` // Encrypted
	CreatedAt     NullableTime `db:"created_at"`
	UpdatedAt     NullableTime `db:"updated_at"`
}

// BuildArtifact stores metadata of a build artifact
type BuildArtifact struct {
	ID            uuid.UUID    `db:"id"`
	ApplicationID uuid.UUID    `db:"application_id"`
	GitCommitSHA  string       `db:"git_commit_sha"`
	Version       string       `db:"version"`
	MD5Hash       string       `db:"md5_hash"`
	LocalPath     string       `db:"local_path"`
	CreatedAt     NullableTime `db:"created_at"`
}

// DeploymentInstance tracks the running status of a deployment instance
type DeploymentInstance struct {
	ID                    uuid.UUID    `db:"id"`
	ApplicationInstanceID uuid.UUID    `db:"application_instance_id"`
	Version               string       `db:"version"`
	GitCommitSHA          string       `db:"git_commit_sha"` // Add this field
	ReleasePath           string       `db:"release_path"`
	Port                  int          `db:"port"`
	Status                string       `db:"status"` // e.g., running, stopped, active, standby, failed
	StartedAt             NullableTime `db:"started_at"`
	StoppedAt             NullableTime `db:"stopped_at"`
	CreatedAt             NullableTime `db:"created_at"`
}

// Domain stores a domain bound to an application instance
type Domain struct {
	ID                    uuid.UUID    `db:"id"`
	ApplicationInstanceID uuid.UUID    `db:"application_instance_id"`
	Hostname              string       `db:"hostname"`
	IsPrimary             bool         `db:"is_primary"`
	CreatedAt             NullableTime `db:"created_at"`
}
