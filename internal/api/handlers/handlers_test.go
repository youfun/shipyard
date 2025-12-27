package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// MockRepository is a mock implementation of DatabaseRepository for testing
type MockRepository struct {
	// SSH Hosts
	MockGetAllSSHHosts   func() ([]models.SSHHost, error)
	MockGetSSHHostByID   func(id uuid.UUID) (*database.SSHHostRow, error)
	MockGetSSHHostByName func(name string) (*models.SSHHost, error)
	MockCreateSSHHost    func(name, addr string, port int, user string, password, privateKey, hostKey *string) (*database.SSHHostRow, error)
	MockUpdateSSHHost    func(id uuid.UUID, name, addr string, port int, user string, password, privateKey *string) error
	MockDeleteSSHHost    func(id uuid.UUID) error

	// Applications
	MockGetAllApplications          func() ([]models.Application, error)
	MockGetApplicationByID          func(id uuid.UUID) (*models.Application, error)
	MockGetApplicationByName        func(name string) (*models.Application, error)
	MockAddApplication              func(app *models.Application) error
	MockGetApplicationInstances     func(appID uuid.UUID) ([]database.ApplicationInstanceRow, error)
	MockGetLastSuccessfulHostForApp func(appID uuid.UUID) (string, time.Time, error)

	// Application Instances
	MockGetApplicationInstance func(appID uuid.UUID, hostID uuid.UUID) (*models.ApplicationInstance, error)
	MockGetInstance            func(appName, hostName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error)
	MockLinkApplicationToHost  func(instance *models.ApplicationInstance) error

	// Secrets
	MockListSecretKeys   func(appID uuid.UUID) ([]string, error)
	MockGetSecretsForApp func(appID uuid.UUID) (map[string]string, error)
	MockGetAllSecrets    func(appID uuid.UUID) (map[string]string, error)
	MockSetSecret        func(appID uuid.UUID, key, value string) error
	MockUnsetSecret      func(appID uuid.UUID, key string) error

	// Deployments
	MockGetDeploymentHistoryForApp        func(appName string) ([]database.DeploymentHistoryRow, error)
	MockGetDeploymentHistoryByID          func(id uuid.UUID) (*database.DeploymentHistoryRow, error)
	MockGetLatestDeploymentHistoryForApp  func(appName string) (*models.DeploymentHistory, error)
	MockCreateDeploymentHistoryWithStatus func(instanceID uuid.UUID, version, status, output string) (*models.DeploymentHistory, error)
	MockAppendDeploymentHistoryOutput     func(id uuid.UUID, output string) error
	MockUpdateDeploymentHistoryStatusOnly func(id uuid.UUID, status string) error
	MockGetDeploymentsCount               func() (int, error)

	// Domains
	MockGetDomainsForInstance func(instanceID uuid.UUID) ([]models.Domain, error)
	MockGetDomainByID         func(id uuid.UUID) (*models.Domain, error)
	MockAddDomain             func(domain *models.Domain) error
	MockUpdateDomain          func(id uuid.UUID, hostname string, isPrimary bool) error
	MockDeleteDomainByID      func(id uuid.UUID) error
	MockSetPrimaryDomain      func(instanceID uuid.UUID, hostname string) error

	// Build Artifacts
	MockGetBuildArtifactByGitSHA   func(appID uuid.UUID, gitSHA string) (*models.BuildArtifact, error)
	MockGetAllBuildArtifactsForApp func(appID uuid.UUID) ([]models.BuildArtifact, error)
	MockAddBuildArtifact           func(artifact *models.BuildArtifact) error

	// Users
	MockGetUserCount       func() (int64, error)
	MockGetUserByUsername  func(username string) (*models.User, error)
	MockGetUserByID        func(id uuid.UUID) (*models.User, error)
	MockCreateUser         func(tx *sqlx.Tx, username, password string) (*models.User, error)
	MockUpdateUserPassword func(id uuid.UUID, password string) error

	// Auth Devices
	MockCreateAuthDevice       func(tx *sqlx.Tx, userID uuid.UUID, ip, userAgent string) (*models.AuthDevice, error)
	MockListAuthDevicesForUser func(userID uuid.UUID) ([]*models.AuthDevice, error)
	MockRevokeAuthDevice       func(userID, deviceID uuid.UUID) error

	// 2FA
	MockEnable2FAForUser  func(tx *sqlx.Tx, userID uuid.UUID, secret string, recoveryCodes []string) error
	MockDisable2FAForUser func(tx *sqlx.Tx, userID uuid.UUID) error

	// Application Tokens
	MockGetApplicationTokensByAppID func(appID uuid.UUID) ([]database.ApplicationToken, error)
	MockCreateApplicationToken      func(appID uuid.UUID, name string, expiresAt *time.Time) (*database.ApplicationToken, string, error)
	MockDeleteApplicationToken      func(tokenID, appID uuid.UUID) error
}

// Implement the DatabaseRepository interface methods
func (m *MockRepository) GetDB() *sqlx.DB { return nil }

func (m *MockRepository) GetAllSSHHosts() ([]models.SSHHost, error) {
	if m.MockGetAllSSHHosts != nil {
		return m.MockGetAllSSHHosts()
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetSSHHostByID(id uuid.UUID) (*database.SSHHostRow, error) {
	if m.MockGetSSHHostByID != nil {
		return m.MockGetSSHHostByID(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetSSHHostByName(name string) (*models.SSHHost, error) {
	if m.MockGetSSHHostByName != nil {
		return m.MockGetSSHHostByName(name)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) CreateSSHHost(name, addr string, port int, user string, password, privateKey, hostKey *string) (*database.SSHHostRow, error) {
	if m.MockCreateSSHHost != nil {
		return m.MockCreateSSHHost(name, addr, port, user, password, privateKey, hostKey)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) UpdateSSHHost(id uuid.UUID, name, addr string, port int, user string, password, privateKey *string) error {
	if m.MockUpdateSSHHost != nil {
		return m.MockUpdateSSHHost(id, name, addr, port, user, password, privateKey)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) DeleteSSHHost(id uuid.UUID) error {
	if m.MockDeleteSSHHost != nil {
		return m.MockDeleteSSHHost(id)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetAllApplications() ([]models.Application, error) {
	if m.MockGetAllApplications != nil {
		return m.MockGetAllApplications()
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetApplicationByID(id uuid.UUID) (*models.Application, error) {
	if m.MockGetApplicationByID != nil {
		return m.MockGetApplicationByID(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetApplicationByName(name string) (*models.Application, error) {
	if m.MockGetApplicationByName != nil {
		return m.MockGetApplicationByName(name)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) AddApplication(app *models.Application) error {
	if m.MockAddApplication != nil {
		return m.MockAddApplication(app)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetApplicationInstances(appID uuid.UUID) ([]database.ApplicationInstanceRow, error) {
	if m.MockGetApplicationInstances != nil {
		return m.MockGetApplicationInstances(appID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetLastSuccessfulHostForApp(appID uuid.UUID) (string, time.Time, error) {
	if m.MockGetLastSuccessfulHostForApp != nil {
		return m.MockGetLastSuccessfulHostForApp(appID)
	}
	return "", time.Time{}, errors.New("not implemented")
}

func (m *MockRepository) GetApplicationInstance(appID uuid.UUID, hostID uuid.UUID) (*models.ApplicationInstance, error) {
	if m.MockGetApplicationInstance != nil {
		return m.MockGetApplicationInstance(appID, hostID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetInstance(appName, hostName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
	if m.MockGetInstance != nil {
		return m.MockGetInstance(appName, hostName)
	}
	return nil, nil, nil, errors.New("not implemented")
}

func (m *MockRepository) EnsureLocalhostInstance(appName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
	if m.MockGetInstance != nil {
		// Reuse MockGetInstance for localhost
		return m.MockGetInstance(appName, "localhost")
	}
	return nil, nil, nil, errors.New("not implemented")
}

func (m *MockRepository) LinkApplicationToHost(instance *models.ApplicationInstance) error {
	if m.MockLinkApplicationToHost != nil {
		return m.MockLinkApplicationToHost(instance)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) ListSecretKeys(appID uuid.UUID) ([]string, error) {
	if m.MockListSecretKeys != nil {
		return m.MockListSecretKeys(appID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetSecretsForApp(appID uuid.UUID) (map[string]string, error) {
	if m.MockGetSecretsForApp != nil {
		return m.MockGetSecretsForApp(appID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetAllSecrets(appID uuid.UUID) (map[string]string, error) {
	if m.MockGetAllSecrets != nil {
		return m.MockGetAllSecrets(appID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) SetSecret(appID uuid.UUID, key, value string) error {
	if m.MockSetSecret != nil {
		return m.MockSetSecret(appID, key, value)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) UnsetSecret(appID uuid.UUID, key string) error {
	if m.MockUnsetSecret != nil {
		return m.MockUnsetSecret(appID, key)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetDeploymentHistoryForApp(appName string) ([]database.DeploymentHistoryRow, error) {
	if m.MockGetDeploymentHistoryForApp != nil {
		return m.MockGetDeploymentHistoryForApp(appName)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetDeploymentHistoryByID(id uuid.UUID) (*database.DeploymentHistoryRow, error) {
	if m.MockGetDeploymentHistoryByID != nil {
		return m.MockGetDeploymentHistoryByID(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetLatestDeploymentHistoryForApp(appName string) (*models.DeploymentHistory, error) {
	if m.MockGetLatestDeploymentHistoryForApp != nil {
		return m.MockGetLatestDeploymentHistoryForApp(appName)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) CreateDeploymentHistoryWithStatus(instanceID uuid.UUID, version, status, output string) (*models.DeploymentHistory, error) {
	if m.MockCreateDeploymentHistoryWithStatus != nil {
		return m.MockCreateDeploymentHistoryWithStatus(instanceID, version, status, output)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) AppendDeploymentHistoryOutput(id uuid.UUID, output string) error {
	if m.MockAppendDeploymentHistoryOutput != nil {
		return m.MockAppendDeploymentHistoryOutput(id, output)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) UpdateDeploymentHistoryStatusOnly(id uuid.UUID, status string) error {
	if m.MockUpdateDeploymentHistoryStatusOnly != nil {
		return m.MockUpdateDeploymentHistoryStatusOnly(id, status)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetDeploymentsCount() (int, error) {
	if m.MockGetDeploymentsCount != nil {
		return m.MockGetDeploymentsCount()
	}
	return 0, errors.New("not implemented")
}

func (m *MockRepository) GetDomainsForInstance(instanceID uuid.UUID) ([]models.Domain, error) {
	if m.MockGetDomainsForInstance != nil {
		return m.MockGetDomainsForInstance(instanceID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetDomainByID(id uuid.UUID) (*models.Domain, error) {
	if m.MockGetDomainByID != nil {
		return m.MockGetDomainByID(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) AddDomain(domain *models.Domain) error {
	if m.MockAddDomain != nil {
		return m.MockAddDomain(domain)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) UpdateDomain(id uuid.UUID, hostname string, isPrimary bool) error {
	if m.MockUpdateDomain != nil {
		return m.MockUpdateDomain(id, hostname, isPrimary)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) DeleteDomainByID(id uuid.UUID) error {
	if m.MockDeleteDomainByID != nil {
		return m.MockDeleteDomainByID(id)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) SetPrimaryDomain(instanceID uuid.UUID, hostname string) error {
	if m.MockSetPrimaryDomain != nil {
		return m.MockSetPrimaryDomain(instanceID, hostname)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetBuildArtifactByGitSHA(appID uuid.UUID, gitSHA string) (*models.BuildArtifact, error) {
	if m.MockGetBuildArtifactByGitSHA != nil {
		return m.MockGetBuildArtifactByGitSHA(appID, gitSHA)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetAllBuildArtifactsForApp(appID uuid.UUID) ([]models.BuildArtifact, error) {
	if m.MockGetAllBuildArtifactsForApp != nil {
		return m.MockGetAllBuildArtifactsForApp(appID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) AddBuildArtifact(artifact *models.BuildArtifact) error {
	if m.MockAddBuildArtifact != nil {
		return m.MockAddBuildArtifact(artifact)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetUserCount() (int64, error) {
	if m.MockGetUserCount != nil {
		return m.MockGetUserCount()
	}
	return 0, errors.New("not implemented")
}

func (m *MockRepository) GetUserByUsername(username string) (*models.User, error) {
	if m.MockGetUserByUsername != nil {
		return m.MockGetUserByUsername(username)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	if m.MockGetUserByID != nil {
		return m.MockGetUserByID(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) CreateUser(tx *sqlx.Tx, username, password string) (*models.User, error) {
	if m.MockCreateUser != nil {
		return m.MockCreateUser(tx, username, password)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) UpdateUserPassword(id uuid.UUID, password string) error {
	if m.MockUpdateUserPassword != nil {
		return m.MockUpdateUserPassword(id, password)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) CreateAuthDevice(tx *sqlx.Tx, userID uuid.UUID, ip, userAgent string) (*models.AuthDevice, error) {
	if m.MockCreateAuthDevice != nil {
		return m.MockCreateAuthDevice(tx, userID, ip, userAgent)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) ListAuthDevicesForUser(userID uuid.UUID) ([]*models.AuthDevice, error) {
	if m.MockListAuthDevicesForUser != nil {
		return m.MockListAuthDevicesForUser(userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) RevokeAuthDevice(userID, deviceID uuid.UUID) error {
	if m.MockRevokeAuthDevice != nil {
		return m.MockRevokeAuthDevice(userID, deviceID)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) Enable2FAForUser(tx *sqlx.Tx, userID uuid.UUID, secret string, recoveryCodes []string) error {
	if m.MockEnable2FAForUser != nil {
		return m.MockEnable2FAForUser(tx, userID, secret, recoveryCodes)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) Disable2FAForUser(tx *sqlx.Tx, userID uuid.UUID) error {
	if m.MockDisable2FAForUser != nil {
		return m.MockDisable2FAForUser(tx, userID)
	}
	return errors.New("not implemented")
}

func (m *MockRepository) GetApplicationTokensByAppID(appID uuid.UUID) ([]database.ApplicationToken, error) {
	if m.MockGetApplicationTokensByAppID != nil {
		return m.MockGetApplicationTokensByAppID(appID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) CreateApplicationToken(appID uuid.UUID, name string, expiresAt *time.Time) (*database.ApplicationToken, string, error) {
	if m.MockCreateApplicationToken != nil {
		return m.MockCreateApplicationToken(appID, name, expiresAt)
	}
	return nil, "", errors.New("not implemented")
}

func (m *MockRepository) DeleteApplicationToken(tokenID, appID uuid.UUID) error {
	if m.MockDeleteApplicationToken != nil {
		return m.MockDeleteApplicationToken(tokenID, appID)
	}
	return errors.New("not implemented")
}

// Ensure MockRepository implements DatabaseRepository
var _ DatabaseRepository = (*MockRepository)(nil)

// Test helpers
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestListSSHHosts tests the ListSSHHosts handler with mock repository
func TestListSSHHosts(t *testing.T) {
	// Create mock repository
	mockRepo := &MockRepository{
		MockGetAllSSHHosts: func() ([]models.SSHHost, error) {
			now := time.Now()
			return []models.SSHHost{
				{
					ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name:      "test-host",
					Addr:      "192.168.1.1",
					Port:      22,
					User:      "admin",
					Status:    "healthy",
					CreatedAt: models.NullableTime{Time: &now},
				},
			}, nil
		},
	}

	// Create handlers with mock
	h := NewHandlers(mockRepo)

	// Set up router
	router := setupTestRouter()
	router.GET("/ssh-hosts", h.ListSSHHosts)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ssh-hosts", nil)
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data to be an array, got %T", response["data"])
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 host, got %d", len(data))
	}

	host := data[0].(map[string]interface{})
	if host["name"] != "test-host" {
		t.Errorf("Expected name 'test-host', got %v", host["name"])
	}
}

// TestListSSHHostsError tests error handling
func TestListSSHHostsError(t *testing.T) {
	// Create mock repository that returns error
	mockRepo := &MockRepository{
		MockGetAllSSHHosts: func() ([]models.SSHHost, error) {
			return nil, errors.New("database error")
		},
	}

	// Create handlers with mock
	h := NewHandlers(mockRepo)

	// Set up router
	router := setupTestRouter()
	router.GET("/ssh-hosts", h.ListSSHHosts)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ssh-hosts", nil)
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestListApplications tests the ListApplications handler with mock repository
func TestListApplications(t *testing.T) {
	appID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := time.Now()

	// Create mock repository
	mockRepo := &MockRepository{
		MockGetAllApplications: func() ([]models.Application, error) {
			return []models.Application{
				{
					ID:        appID,
					Name:      "test-app",
					CreatedAt: models.NullableTime{Time: &now},
				},
			}, nil
		},
		MockGetLastSuccessfulHostForApp: func(id uuid.UUID) (string, time.Time, error) {
			return "test-host", now, nil
		},
	}

	// Create handlers with mock
	h := NewHandlers(mockRepo)

	// Set up router
	router := setupTestRouter()
	router.GET("/applications", h.ListApplications)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/applications", nil)
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data to be an array, got %T", response["data"])
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 application, got %d", len(data))
	}

	app := data[0].(map[string]interface{})
	if app["name"] != "test-app" {
		t.Errorf("Expected name 'test-app', got %v", app["name"])
	}
	if app["linked_host"] != "test-host" {
		t.Errorf("Expected linked_host 'test-host', got %v", app["linked_host"])
	}
}

// TestDashboardStats tests the DashboardStats handler with mock repository
func TestDashboardStats(t *testing.T) {
	// Create mock repository
	mockRepo := &MockRepository{
		MockGetAllApplications: func() ([]models.Application, error) {
			return []models.Application{{}, {}}, nil // 2 apps
		},
		MockGetAllSSHHosts: func() ([]models.SSHHost, error) {
			return []models.SSHHost{{}}, nil // 1 host
		},
		MockGetDeploymentsCount: func() (int, error) {
			return 5, nil // 5 deployments
		},
	}

	// Create handlers with mock
	h := NewHandlers(mockRepo)

	// Set up router
	router := setupTestRouter()
	router.GET("/dashboard/stats", h.DashboardStats)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/stats", nil)
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response DashboardStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.ApplicationsCount != 2 {
		t.Errorf("Expected 2 applications, got %d", response.ApplicationsCount)
	}
	if response.HostsCount != 1 {
		t.Errorf("Expected 1 host, got %d", response.HostsCount)
	}
	if response.DeploymentsCount != 5 {
		t.Errorf("Expected 5 deployments, got %d", response.DeploymentsCount)
	}
}

// TestCLIGetInstance tests the CLIGetInstance handler returns SSH credentials
func TestCLIGetInstance(t *testing.T) {
	instanceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	appID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	hostID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	// Test credentials - these are placeholder values for testing only
	testPassword := "TEST_PASSWORD_PLACEHOLDER"
	testPrivateKey := "TEST_PRIVATE_KEY_PLACEHOLDER"

	// Create mock repository
	mockRepo := &MockRepository{
		MockGetInstance: func(appName, hostName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
			return &models.ApplicationInstance{
					ID:            instanceID,
					ApplicationID: appID,
					HostID:        hostID,
					Status:        "active",
				},
				&models.Application{
					ID:   appID,
					Name: appName,
				},
				&models.SSHHost{
					ID:         hostID,
					Name:       hostName,
					Addr:       "192.168.1.1",
					Port:       22,
					User:       "admin",
					Password:   &testPassword,
					PrivateKey: &testPrivateKey,
				}, nil
		},
	}

	// Create handlers with mock
	h := NewHandlers(mockRepo)

	// Set up router
	router := setupTestRouter()
	router.GET("/cli/v1/instance", h.CLIGetInstance)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cli/v1/instance?app=test-app&host=test-host", nil)
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check host contains credentials
	host, ok := response["host"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected host to be an object, got %T", response["host"])
	}

	// Verify password is returned
	if host["password"] != testPassword {
		t.Errorf("Expected password '%s', got '%v'", testPassword, host["password"])
	}

	// Verify private_key is returned
	if host["private_key"] != testPrivateKey {
		t.Errorf("Expected private_key '%s', got '%v'", testPrivateKey, host["private_key"])
	}

	// Verify other host fields
	if host["name"] != "test-host" {
		t.Errorf("Expected name 'test-host', got %v", host["name"])
	}
	if host["addr"] != "192.168.1.1" {
		t.Errorf("Expected addr '192.168.1.1', got %v", host["addr"])
	}
	if host["port"] != float64(22) {
		t.Errorf("Expected port 22, got %v", host["port"])
	}
	if host["user"] != "admin" {
		t.Errorf("Expected user 'admin', got %v", host["user"])
	}
}

// TestCLIGetInstanceMissingParams tests CLIGetInstance with missing parameters
func TestCLIGetInstanceMissingParams(t *testing.T) {
	mockRepo := &MockRepository{}
	h := NewHandlers(mockRepo)

	router := setupTestRouter()
	router.GET("/cli/v1/instance", h.CLIGetInstance)

	// Test missing app parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cli/v1/instance?host=test-host", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Test missing host parameter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/cli/v1/instance?app=test-app", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
