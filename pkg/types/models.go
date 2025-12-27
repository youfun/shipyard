// Package types contains shared data types and DTOs used by both CLI and Server.
package types

import (
	"time"
)

// SSHHostDTO represents SSH host information for API transfer
type SSHHostDTO struct {
	ID            string     `json:"id,omitempty"`
	UID           string     `json:"uid,omitempty"`
	Name          string     `json:"name"`
	Addr          string     `json:"addr"`
	Port          int        `json:"port"`
	User          string     `json:"user"`
	Password      *string    `json:"password,omitempty"`
	PrivateKey    *string    `json:"private_key,omitempty"`
	Status        string     `json:"status,omitempty"`
	Arch          string     `json:"arch,omitempty"`
	InitializedAt *time.Time `json:"initialized_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

// ApplicationDTO represents application information for API transfer
type ApplicationDTO struct {
	ID        string     `json:"id,omitempty"`
	UID       string     `json:"uid,omitempty"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// ApplicationInstanceDTO represents an application instance for API transfer
type ApplicationInstanceDTO struct {
	ID                 string `json:"id,omitempty"`
	UID                string `json:"uid,omitempty"`
	ApplicationID      string `json:"application_id,omitempty"`
	HostID             string `json:"host_id,omitempty"`
	Status             string `json:"status,omitempty"`
	ActivePort         int64  `json:"active_port,omitempty"`
	PreviousActivePort int64  `json:"previous_active_port,omitempty"`
}

// DeploymentHistoryDTO represents deployment history for API transfer
type DeploymentHistoryDTO struct {
	ID          string     `json:"id,omitempty"`
	UID         string     `json:"uid,omitempty"`
	InstanceID  string     `json:"instance_id,omitempty"`
	Version     string     `json:"version"`
	ReleasePath string     `json:"release_path,omitempty"`
	Status      string     `json:"status"`
	LogOutput   string     `json:"log_output,omitempty"`
	HostName    string     `json:"host_name,omitempty"`
	DeployedAt  *time.Time `json:"deployed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// SecretDTO represents a secret for API transfer (value is not transferred for security)
type SecretDTO struct {
	Key string `json:"key"`
}

// BuildArtifactDTO represents build artifact metadata for API transfer
type BuildArtifactDTO struct {
	ID            string     `json:"id,omitempty"`
	ApplicationID string     `json:"application_id,omitempty"`
	GitCommitSHA  string     `json:"git_commit_sha,omitempty"`
	Version       string     `json:"version"`
	MD5Hash       string     `json:"md5_hash"`
	LocalPath     string     `json:"local_path,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

// DeployConfigResponse is the aggregated config returned by the server for CLI deployment
type DeployConfigResponse struct {
	DeploymentID string                 `json:"deployment_id"`
	App          ApplicationDTO         `json:"app"`
	Host         SSHHostDTO             `json:"host"`
	Instance     ApplicationInstanceDTO `json:"instance"`
	Secrets      map[string]string      `json:"secrets,omitempty"`
	Domains      []string               `json:"domains,omitempty"`
}

// CreateDeploymentRequest is the request to create a new deployment
type CreateDeploymentRequest struct {
	AppName  string `json:"app_name"`
	HostName string `json:"host_name"`
	Version  string `json:"version,omitempty"`
}

// UpdateDeploymentStatusRequest is the request to update deployment status
type UpdateDeploymentStatusRequest struct {
	Status       string `json:"status"`
	Port         int    `json:"port,omitempty"`
	ReleasePath  string `json:"release_path,omitempty"`
	GitCommitSHA string `json:"git_commit_sha,omitempty"`
}

// UploadDeploymentLogsRequest is the request to upload deployment logs
type UploadDeploymentLogsRequest struct {
	Logs string `json:"logs"`
}

// LinkAppRequest is the request to link an application to a host
type LinkAppRequest struct {
	AppName  string `json:"app_name"`
	HostName string `json:"host_name"`
}

// CreateAppRequest is the request to create a new application
type CreateAppRequest struct {
	Name string `json:"name"`
}

// SyncDomainsRequest is the request to sync domains for an application instance
type SyncDomainsRequest struct {
	InstanceID    string   `json:"instance_id"`
	Domains       []string `json:"domains"`
	PrimaryDomain string   `json:"primary_domain,omitempty"`
}

// APIResponse is a generic API response wrapper
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}
