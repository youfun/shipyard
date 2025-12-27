package handlers

import (
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository interfaces for dependency injection and testing

// SSHHostRepository defines methods for SSH host data operations
type SSHHostRepository interface {
	GetAllSSHHosts() ([]models.SSHHost, error)
	GetSSHHostByID(id uuid.UUID) (*database.SSHHostRow, error)
	GetSSHHostByName(name string) (*models.SSHHost, error)
	CreateSSHHost(name, addr string, port int, user string, password, privateKey, hostKey *string) (*database.SSHHostRow, error)
	UpdateSSHHost(id uuid.UUID, name, addr string, port int, user string, password, privateKey *string) error
	DeleteSSHHost(id uuid.UUID) error
}

// ApplicationRepository defines methods for application data operations
type ApplicationRepository interface {
	GetAllApplications() ([]models.Application, error)
	GetApplicationByID(id uuid.UUID) (*models.Application, error)
	GetApplicationByName(name string) (*models.Application, error)
	AddApplication(app *models.Application) error
	GetApplicationInstances(appID uuid.UUID) ([]database.ApplicationInstanceRow, error)
	GetLastSuccessfulHostForApp(appID uuid.UUID) (string, time.Time, error)
}

// ApplicationInstanceRepository defines methods for application instance data operations
type ApplicationInstanceRepository interface {
	GetApplicationInstance(appID uuid.UUID, hostID uuid.UUID) (*models.ApplicationInstance, error)
	GetApplicationInstanceByID(instanceID uuid.UUID) (*models.ApplicationInstance, error)
	GetInstance(appName, hostName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error)
	EnsureLocalhostInstance(appName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error)
	LinkApplicationToHost(instance *models.ApplicationInstance) error
	UpdateApplicationInstanceStatus(id uuid.UUID, status string) error
}

// SecretRepository defines methods for secrets data operations
type SecretRepository interface {
	ListSecretKeys(appID uuid.UUID) ([]string, error)
	GetSecretsForApp(appID uuid.UUID) (map[string]string, error)
	GetAllSecrets(appID uuid.UUID) (map[string]string, error)
	SetSecret(appID uuid.UUID, key, value string) error
	UnsetSecret(appID uuid.UUID, key string) error
}

// DeploymentRepository defines methods for deployment data operations
type DeploymentRepository interface {
	GetDeploymentHistoryForApp(appName string) ([]database.DeploymentHistoryRow, error)
	GetDeploymentHistoryByID(id uuid.UUID) (*database.DeploymentHistoryRow, error)
	GetLatestDeploymentHistoryForApp(appName string) (*models.DeploymentHistory, error)
	CreateDeploymentHistoryWithStatus(instanceID uuid.UUID, version, status, output string) (*models.DeploymentHistory, error)
	AppendDeploymentHistoryOutput(id uuid.UUID, output string) error
	UpdateDeploymentHistoryStatusOnly(id uuid.UUID, status string) error
	RecordSuccessfulDeployment(deploymentID uuid.UUID, port int, releasePath string, gitCommitSHA string) error
	GetDeploymentsCount() (int, error)
	GetRecentDeploymentsGlobal(limit int) ([]database.RecentDeploymentRow, error)
}

// DomainRepository defines methods for domain data operations
type DomainRepository interface {
	GetDomainsForInstance(instanceID uuid.UUID) ([]models.Domain, error)
	GetDomainByID(id uuid.UUID) (*models.Domain, error)
	AddDomain(domain *models.Domain) error
	UpdateDomain(id uuid.UUID, hostname string, isPrimary bool) error
	DeleteDomainByID(id uuid.UUID) error
	SetPrimaryDomain(instanceID uuid.UUID, hostname string) error
}

// BuildArtifactRepository defines methods for build artifact data operations
type BuildArtifactRepository interface {
	GetBuildArtifactByGitSHA(appID uuid.UUID, gitSHA string) (*models.BuildArtifact, error)
	GetBuildArtifactByMD5Prefix(appID uuid.UUID, md5Prefix string) (*models.BuildArtifact, error)
	GetAllBuildArtifactsForApp(appID uuid.UUID) ([]models.BuildArtifact, error)
	AddBuildArtifact(artifact *models.BuildArtifact) error
}

// UserRepository defines methods for user data operations
type UserRepository interface {
	GetUserCount() (int64, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	CreateUser(tx *sqlx.Tx, username, password string) (*models.User, error)
	UpdateUserPassword(id uuid.UUID, password string) error
}

// AuthDeviceRepository defines methods for auth device data operations
type AuthDeviceRepository interface {
	CreateAuthDevice(tx *sqlx.Tx, userID uuid.UUID, ip, userAgent string) (*models.AuthDevice, error)
	ListAuthDevicesForUser(userID uuid.UUID) ([]*models.AuthDevice, error)
	RevokeAuthDevice(userID, deviceID uuid.UUID) error
}

// TwoFactorRepository defines methods for 2FA data operations
type TwoFactorRepository interface {
	Enable2FAForUser(tx *sqlx.Tx, userID uuid.UUID, secret string, recoveryCodes []string) error
	Disable2FAForUser(tx *sqlx.Tx, userID uuid.UUID) error
}

// ApplicationTokenRepository defines methods for application token data operations
type ApplicationTokenRepository interface {
	GetApplicationTokensByAppID(appID uuid.UUID) ([]database.ApplicationToken, error)
	CreateApplicationToken(appID uuid.UUID, name string, expiresAt *time.Time) (*database.ApplicationToken, string, error)
	DeleteApplicationToken(tokenID, appID uuid.UUID) error
}

// SystemSettingsRepository defines methods for system-wide configuration
type SystemSettingsRepository interface {
	GetSystemSetting(key string) (string, error)
	SetSystemSetting(key, value string) error
}

// DatabaseRepository combines all repository interfaces for convenience
type DatabaseRepository interface {
	SSHHostRepository
	ApplicationRepository
	ApplicationInstanceRepository
	SecretRepository
	DeploymentRepository
	DomainRepository
	BuildArtifactRepository
	UserRepository
	AuthDeviceRepository
	TwoFactorRepository
	ApplicationTokenRepository
	SystemSettingsRepository
	// DB returns the underlying database connection for transactions
	GetDB() *sqlx.DB
}

// DefaultRepository is the concrete implementation that uses the database package
type DefaultRepository struct{}

// Ensure DefaultRepository implements DatabaseRepository
var _ DatabaseRepository = (*DefaultRepository)(nil)

// GetDB returns the underlying database connection
func (r *DefaultRepository) GetDB() *sqlx.DB {
	return database.DB
}

// SSHHostRepository implementations
func (r *DefaultRepository) GetAllSSHHosts() ([]models.SSHHost, error) {
	return database.GetAllSSHHosts()
}

func (r *DefaultRepository) GetSSHHostByID(id uuid.UUID) (*database.SSHHostRow, error) {
	return database.GetSSHHostByID(id)
}

func (r *DefaultRepository) GetSSHHostByName(name string) (*models.SSHHost, error) {
	return database.GetSSHHostByName(name)
}

func (r *DefaultRepository) CreateSSHHost(name, addr string, port int, user string, password, privateKey, hostKey *string) (*database.SSHHostRow, error) {
	return database.CreateSSHHost(name, addr, port, user, password, privateKey, hostKey)
}

func (r *DefaultRepository) UpdateSSHHost(id uuid.UUID, name, addr string, port int, user string, password, privateKey *string) error {
	return database.UpdateSSHHost(id, name, addr, port, user, password, privateKey)
}

func (r *DefaultRepository) DeleteSSHHost(id uuid.UUID) error {
	return database.DeleteSSHHost(id)
}

// ApplicationRepository implementations
func (r *DefaultRepository) GetAllApplications() ([]models.Application, error) {
	return database.GetAllApplications()
}

func (r *DefaultRepository) GetApplicationByID(id uuid.UUID) (*models.Application, error) {
	return database.GetApplicationByID(id)
}

func (r *DefaultRepository) GetApplicationByName(name string) (*models.Application, error) {
	return database.GetApplicationByName(name)
}

func (r *DefaultRepository) AddApplication(app *models.Application) error {
	return database.AddApplication(app)
}

func (r *DefaultRepository) GetApplicationInstances(appID uuid.UUID) ([]database.ApplicationInstanceRow, error) {
	return database.GetApplicationInstances(appID)
}

func (r *DefaultRepository) GetLastSuccessfulHostForApp(appID uuid.UUID) (string, time.Time, error) {
	return database.GetLastSuccessfulHostForApp(appID)
}

// ApplicationInstanceRepository implementations
func (r *DefaultRepository) GetApplicationInstance(appID uuid.UUID, hostID uuid.UUID) (*models.ApplicationInstance, error) {
	return database.GetApplicationInstance(appID, hostID)
}

func (r *DefaultRepository) GetApplicationInstanceByID(instanceID uuid.UUID) (*models.ApplicationInstance, error) {
	return database.GetApplicationInstanceByID(instanceID)
}

func (r *DefaultRepository) GetInstance(appName, hostName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
	return database.GetInstance(appName, hostName)
}

func (r *DefaultRepository) EnsureLocalhostInstance(appName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
	return database.EnsureLocalhostInstance(appName)
}

func (r *DefaultRepository) LinkApplicationToHost(instance *models.ApplicationInstance) error {
	return database.LinkApplicationToHost(instance)
}

func (r *DefaultRepository) UpdateApplicationInstanceStatus(id uuid.UUID, status string) error {
	return database.UpdateApplicationInstanceStatus(id, status)
}

// SecretRepository implementations
func (r *DefaultRepository) ListSecretKeys(appID uuid.UUID) ([]string, error) {
	return database.ListSecretKeys(appID)
}

func (r *DefaultRepository) GetSecretsForApp(appID uuid.UUID) (map[string]string, error) {
	return database.GetSecretsForApp(appID)
}

func (r *DefaultRepository) GetAllSecrets(appID uuid.UUID) (map[string]string, error) {
	return database.GetAllSecrets(appID)
}

func (r *DefaultRepository) SetSecret(appID uuid.UUID, key, value string) error {
	return database.SetSecret(appID, key, value)
}

func (r *DefaultRepository) UnsetSecret(appID uuid.UUID, key string) error {
	return database.UnsetSecret(appID, key)
}

// DeploymentRepository implementations
func (r *DefaultRepository) GetDeploymentHistoryForApp(appName string) ([]database.DeploymentHistoryRow, error) {
	return database.GetDeploymentHistoryForApp(appName)
}

func (r *DefaultRepository) GetDeploymentHistoryByID(id uuid.UUID) (*database.DeploymentHistoryRow, error) {
	return database.GetDeploymentHistoryByID(id)
}

func (r *DefaultRepository) GetLatestDeploymentHistoryForApp(appName string) (*models.DeploymentHistory, error) {
	return database.GetLatestDeploymentHistoryForApp(appName)
}

func (r *DefaultRepository) CreateDeploymentHistoryWithStatus(instanceID uuid.UUID, version, status, output string) (*models.DeploymentHistory, error) {
	return database.CreateDeploymentHistoryWithStatus(instanceID, version, status, output)
}

func (r *DefaultRepository) AppendDeploymentHistoryOutput(id uuid.UUID, output string) error {
	return database.AppendDeploymentHistoryOutput(id, output)
}

func (r *DefaultRepository) UpdateDeploymentHistoryStatusOnly(id uuid.UUID, status string) error {
	return database.UpdateDeploymentHistoryStatusOnly(id, status)
}

func (r *DefaultRepository) RecordSuccessfulDeployment(deploymentID uuid.UUID, port int, releasePath string, gitCommitSHA string) error {
	return database.RecordSuccessfulDeployment(deploymentID, port, releasePath, gitCommitSHA)
}

func (r *DefaultRepository) GetDeploymentsCount() (int, error) {
	return database.GetDeploymentsCount()
}

func (r *DefaultRepository) GetRecentDeploymentsGlobal(limit int) ([]database.RecentDeploymentRow, error) {
	return database.GetRecentDeploymentsGlobal(limit)
}

// DomainRepository implementations
func (r *DefaultRepository) GetDomainsForInstance(instanceID uuid.UUID) ([]models.Domain, error) {
	return database.GetDomainsForInstance(instanceID)
}

func (r *DefaultRepository) GetDomainByID(id uuid.UUID) (*models.Domain, error) {
	return database.GetDomainByID(id)
}

func (r *DefaultRepository) AddDomain(domain *models.Domain) error {
	return database.AddDomain(domain)
}

func (r *DefaultRepository) UpdateDomain(id uuid.UUID, hostname string, isPrimary bool) error {
	return database.UpdateDomain(id, hostname, isPrimary)
}

func (r *DefaultRepository) DeleteDomainByID(id uuid.UUID) error {
	return database.DeleteDomainByID(id)
}

func (r *DefaultRepository) SetPrimaryDomain(instanceID uuid.UUID, hostname string) error {
	return database.SetPrimaryDomain(instanceID, hostname)
}

// BuildArtifactRepository implementations
func (r *DefaultRepository) GetBuildArtifactByGitSHA(appID uuid.UUID, gitSHA string) (*models.BuildArtifact, error) {
	return database.GetBuildArtifactByGitSHA(appID, gitSHA)
}

func (r *DefaultRepository) GetBuildArtifactByMD5Prefix(appID uuid.UUID, md5Prefix string) (*models.BuildArtifact, error) {
	return database.GetBuildArtifactByMD5Prefix(appID, md5Prefix)
}

func (r *DefaultRepository) GetAllBuildArtifactsForApp(appID uuid.UUID) ([]models.BuildArtifact, error) {
	return database.GetAllBuildArtifactsForApp(appID)
}

func (r *DefaultRepository) AddBuildArtifact(artifact *models.BuildArtifact) error {
	return database.AddBuildArtifact(artifact)
}

// UserRepository implementations
func (r *DefaultRepository) GetUserCount() (int64, error) {
	return database.GetUserCount()
}

func (r *DefaultRepository) GetUserByUsername(username string) (*models.User, error) {
	return database.GetUserByUsername(username)
}

func (r *DefaultRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	return database.GetUserByID(id)
}

func (r *DefaultRepository) CreateUser(tx *sqlx.Tx, username, password string) (*models.User, error) {
	return database.CreateUser(tx, username, password)
}

func (r *DefaultRepository) UpdateUserPassword(id uuid.UUID, password string) error {
	return database.UpdateUserPassword(id, password)
}

// AuthDeviceRepository implementations
func (r *DefaultRepository) CreateAuthDevice(tx *sqlx.Tx, userID uuid.UUID, ip, userAgent string) (*models.AuthDevice, error) {
	return database.CreateAuthDevice(tx, userID, ip, userAgent)
}

func (r *DefaultRepository) ListAuthDevicesForUser(userID uuid.UUID) ([]*models.AuthDevice, error) {
	return database.ListAuthDevicesForUser(userID)
}

func (r *DefaultRepository) RevokeAuthDevice(userID, deviceID uuid.UUID) error {
	return database.RevokeAuthDevice(userID, deviceID)
}

// TwoFactorRepository implementations
func (r *DefaultRepository) Enable2FAForUser(tx *sqlx.Tx, userID uuid.UUID, secret string, recoveryCodes []string) error {
	return database.Enable2FAForUser(tx, userID, secret, recoveryCodes)
}

func (r *DefaultRepository) Disable2FAForUser(tx *sqlx.Tx, userID uuid.UUID) error {
	return database.Disable2FAForUser(tx, userID)
}

// ApplicationTokenRepository implementations
func (r *DefaultRepository) GetApplicationTokensByAppID(appID uuid.UUID) ([]database.ApplicationToken, error) {
	return database.GetApplicationTokensByAppID(appID)
}

func (r *DefaultRepository) CreateApplicationToken(appID uuid.UUID, name string, expiresAt *time.Time) (*database.ApplicationToken, string, error) {
	return database.CreateApplicationToken(appID, name, expiresAt)
}

func (r *DefaultRepository) DeleteApplicationToken(tokenID, appID uuid.UUID) error {
	return database.DeleteApplicationToken(tokenID, appID)
}

// SystemSettingsRepository implementations
func (r *DefaultRepository) GetSystemSetting(key string) (string, error) {
	// Since DefaultRepository wraps the package-level functions (which are methods on database.Store but exposed via global DB in some places, but here we added methods to Store),
	// we need to access the methods we added to internal/database/system_settings.go.
	// Looking at system_settings.go, I added methods to *Store.
	// But database package usually exposes a global wrapper or we need to access the store.
	// The pattern in this file seems to be calling database.SomeFunction().
	// I need to check if I need to export wrapper functions in database package or how it's handled.
	// Looking at `internal/database/db-init.go`, there is no global `Store` variable, just `DB`.
	// Wait, the methods I added in `internal/database/system_settings.go` are on `*Store`.
	// But where is `Store` defined?
	// Let me check `internal/database/database.go` or similar.
	// If `Store` isn't used elsewhere, I might have made a mistake in Step 1.2 by attaching methods to `*Store`.
	// Most existing functions seem to be package-level functions using `DB` global.

	// Let's assume for now I should have made them package level functions or created a Store instance.
	// Since I cannot check `database.go` right now in the middle of a replace, I will assume I need to fix `internal/database/system_settings.go` to be package level functions first.
	// But I will write the implementation here assuming they are package level functions.
	return database.GetSystemSetting(key)
}

func (r *DefaultRepository) SetSystemSetting(key, value string) error {
	return database.SetSystemSetting(key, value)
}
