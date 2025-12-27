package client

import (
	"youfun/shipyard/pkg/types"
)

// APIClient defines the interface for communicating with the Deployer Server.
type APIClient interface {
	// Configuration
	GetDeployConfig(appName, hostName string) (*types.DeployConfigResponse, error)

	// Deployment History
	CreateDeployment(req *types.CreateDeploymentRequest) (*types.DeploymentHistoryDTO, error)
	UpdateDeploymentStatus(deploymentID, status string, port int, releasePath, gitCommitSHA string) error
	UploadDeploymentLogs(deploymentID string, logs string) error
	
	// Server-side Deployment
	UploadDeploymentArtifact(deploymentID string, artifactPath string) error
	ExecuteServerDeployment(deploymentID string, version, gitCommitSHA, md5Hash string) error

	// Artifacts
	// CheckArtifact checks if a build artifact exists by query (MD5 prefix, full MD5, or git SHA)
	CheckArtifact(appID, query string) (*types.BuildArtifactDTO, error)
	RegisterArtifact(artifact *types.BuildArtifactDTO) error

	// Host Init
	LinkApp(appName, hostName string) error
	CreateApp(appName string) error
	ListHosts() ([]types.SSHHostDTO, error)

	// Domains
	SyncDomains(instanceID string, domains []string, primaryDomain string) error

	// Secrets (Environment Variables)
	ListSecrets(appName string) ([]string, error)
	SetSecret(appName, key, value string) error
	UnsetSecret(appName, key string) error

	// Instance Management
	GetInstance(appName, hostName string) (*InstanceInfo, error)
	GetHostByName(hostName string) (*types.SSHHostDTO, error)
	GetLastDeployment(appName string) (*types.DeploymentHistoryDTO, error)
}
