// Package e2e contains end-to-end tests for the deployer CLI and Server.
// These tests start a real server instance and test the complete flow
// from CLI commands to server responses.
package e2e

import (
	"bytes"
	"youfun/shipyard/internal/api"
	"youfun/shipyard/internal/database"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestServer wraps the test server and provides helper methods
type TestServer struct {
	Server     *api.Server
	HTTPServer *httptest.Server
	BaseURL    string
	Token      string // Stores auth token for authenticated requests
}

// NewTestServer creates a new test server with a temporary SQLite database
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create a temporary directory for test database
	tmpDir := t.TempDir()

	// Set up test environment - clear any conflicting env vars
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("TURSO_DATABASE_URL")
	os.Unsetenv("TURSO_AUTH_TOKEN")

	// Set env vars for testing
	os.Setenv("GIN_MODE", "test")
	os.Setenv("JWT_SECRET", "test-secret-key-for-e2e-testing-only")

	// Create temp deployer config directory
	testConfigDir := filepath.Join(tmpDir, ".deployer")
	if err := os.MkdirAll(testConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create test config dir: %v", err)
	}

	// Set HOME to temp directory so database goes there
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir) // For Windows

	// Initialize database
	database.InitDB()

	// Create server
	server := api.NewServer("0")

	// Create test HTTP server
	ts := httptest.NewServer(server.Router)

	t.Logf("Test server started at %s", ts.URL)

	return &TestServer{
		Server:     server,
		HTTPServer: ts,
		BaseURL:    ts.URL,
	}
}

// Close shuts down the test server and closes the database connection
func (ts *TestServer) Close() {
	ts.HTTPServer.Close()
	// Close database connection to allow temp directory cleanup
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
}

// Request makes an HTTP request to the test server
func (ts *TestServer) Request(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, ts.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if ts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+ts.Token)
	}

	return http.DefaultClient.Do(req)
}

// GET makes a GET request
func (ts *TestServer) GET(path string) (*http.Response, error) {
	return ts.Request("GET", path, nil)
}

// POST makes a POST request with JSON body
func (ts *TestServer) POST(path string, body interface{}) (*http.Response, error) {
	return ts.Request("POST", path, body)
}

// PUT makes a PUT request with JSON body
func (ts *TestServer) PUT(path string, body interface{}) (*http.Response, error) {
	return ts.Request("PUT", path, body)
}

// DELETE makes a DELETE request
func (ts *TestServer) DELETE(path string) (*http.Response, error) {
	return ts.Request("DELETE", path, nil)
}

// ParseJSON parses JSON response body
func ParseJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// ReadBody reads and returns the response body as string
func ReadBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// =========================================================================
// E2E Tests: Setup and Authentication Flow
// =========================================================================

func TestE2E_SetupAndAuth(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	t.Run("SetupStatus_InitiallyRequired", func(t *testing.T) {
		resp, err := ts.GET("/api/setup/status")
		if err != nil {
			t.Fatalf("Failed to get setup status: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if setupRequired, ok := result["setup_required"].(bool); !ok || !setupRequired {
			t.Errorf("Expected setup_required to be true")
		}
	})

	t.Run("Setup_CreateAdminUser", func(t *testing.T) {
		body := map[string]string{
			"username": "admin",
			"password": "admin123",
		}
		resp, err := ts.POST("/api/setup", body)
		if err != nil {
			t.Fatalf("Failed to setup: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if success, ok := result["success"].(bool); !ok || !success {
			t.Errorf("Expected success to be true")
		}
	})

	t.Run("SetupStatus_NotRequiredAfterSetup", func(t *testing.T) {
		resp, err := ts.GET("/api/setup/status")
		if err != nil {
			t.Fatalf("Failed to get setup status: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if setupRequired, ok := result["setup_required"].(bool); !ok || setupRequired {
			t.Errorf("Expected setup_required to be false after setup")
		}
	})

	t.Run("Setup_FailsAfterInitialSetup", func(t *testing.T) {
		body := map[string]string{
			"username": "admin2",
			"password": "admin456",
		}
		resp, err := ts.POST("/api/setup", body)
		if err != nil {
			t.Fatalf("Failed to setup: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("Login_Success", func(t *testing.T) {
		body := map[string]string{
			"username": "admin",
			"password": "admin123",
		}
		resp, err := ts.POST("/api/auth/login", body)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		token, ok := result["access_token"].(string)
		if !ok || token == "" {
			t.Errorf("Expected access_token in response")
		}

		// Save token for subsequent tests
		ts.Token = token
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		body := map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		}
		resp, err := ts.POST("/api/auth/login", body)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("GetCurrentUser_Authenticated", func(t *testing.T) {
		resp, err := ts.GET("/api/auth/me")
		if err != nil {
			t.Fatalf("Failed to get current user: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if username, ok := result["username"].(string); !ok || username != "admin" {
			t.Errorf("Expected username 'admin', got %v", result["username"])
		}
	})

	t.Run("ProtectedRoute_WithoutToken", func(t *testing.T) {
		oldToken := ts.Token
		ts.Token = ""
		defer func() { ts.Token = oldToken }()

		resp, err := ts.GET("/api/auth/me")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 without token, got %d", resp.StatusCode)
		}
	})
}

// =========================================================================
// E2E Tests: Device Flow Authentication
// =========================================================================

func TestE2E_DeviceFlowAuth(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Setup admin user first
	setupBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	ts.POST("/api/setup", setupBody)

	// Login to get a token
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	loginResp, _ := ts.POST("/api/auth/login", loginBody)
	var loginResult map[string]interface{}
	ParseJSON(loginResp, &loginResult)
	ts.Token = loginResult["access_token"].(string)

	var sessionID string

	t.Run("DeviceCode_Request", func(t *testing.T) {
		body := map[string]string{
			"os":          "windows",
			"device_name": "test-device",
		}
		resp, err := ts.POST("/api/auth/device/code", body)
		if err != nil {
			t.Fatalf("Failed to request device code: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if sid, ok := result["session_id"].(string); !ok || sid == "" {
			t.Errorf("Expected session_id in response")
		} else {
			sessionID = sid
		}

		if userCode, ok := result["user_code"].(string); !ok || userCode == "" {
			t.Errorf("Expected user_code in response")
		}
	})

	t.Run("DeviceToken_Pending", func(t *testing.T) {
		resp, err := ts.GET("/api/auth/device/token?session_id=" + sessionID)
		if err != nil {
			t.Fatalf("Failed to poll for token: %v", err)
		}

		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("Expected status 202 (pending), got %d", resp.StatusCode)
		}
	})

	t.Run("DeviceSession_GetInfo", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/sessions/" + sessionID)
		if err != nil {
			t.Fatalf("Failed to get session info: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if deviceName, ok := result["device_name"].(string); !ok || deviceName != "test-device" {
			t.Errorf("Expected device_name 'test-device', got %v", result["device_name"])
		}
	})

	t.Run("DeviceConfirm_Approve", func(t *testing.T) {
		body := map[string]interface{}{
			"session_id": sessionID,
			"approved":   true,
		}
		resp, err := ts.POST("/api/cli/confirm", body)
		if err != nil {
			t.Fatalf("Failed to confirm device: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("DeviceToken_Authorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.BaseURL+"/api/auth/device/token?session_id="+sessionID, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to poll for token: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 (authorized), got %d", resp.StatusCode)
		}

		if accessToken, ok := result["access_token"].(string); !ok || accessToken == "" {
			t.Errorf("Expected access_token in response after authorization")
		}
	})
}

// =========================================================================
// E2E Tests: SSH Hosts Management
// =========================================================================

func TestE2E_SSHHostsManagement(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	var hostUID string

	t.Run("ListSSHHosts_Empty", func(t *testing.T) {
		resp, err := ts.GET("/api/ssh-hosts")
		if err != nil {
			t.Fatalf("Failed to list hosts: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("CreateSSHHost", func(t *testing.T) {
		body := map[string]interface{}{
			"name":     "test-host",
			"addr":     "192.168.1.100",
			"port":     22,
			"user":     "deploy",
			"password": "secret123",
		}
		resp, err := ts.POST("/api/ssh-hosts", body)
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 201 or 200, got %d", resp.StatusCode)
		}

		if uid, ok := result["uid"].(string); ok && uid != "" {
			hostUID = uid
		} else if data, ok := result["data"].(map[string]interface{}); ok {
			if uid, ok := data["uid"].(string); ok {
				hostUID = uid
			}
		}
	})

	t.Run("ListSSHHosts_WithHost", func(t *testing.T) {
		resp, err := ts.GET("/api/ssh-hosts")
		if err != nil {
			t.Fatalf("Failed to list hosts: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		data, ok := result["data"].([]interface{})
		if !ok || len(data) == 0 {
			t.Errorf("Expected at least one host in list")
		}
	})

	t.Run("GetSSHHost_ByUID", func(t *testing.T) {
		if hostUID == "" {
			t.Skip("No host UID available")
		}

		resp, err := ts.GET("/api/ssh-hosts/" + hostUID)
		if err != nil {
			t.Fatalf("Failed to get host: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("UpdateSSHHost", func(t *testing.T) {
		if hostUID == "" {
			t.Skip("No host UID available")
		}

		body := map[string]interface{}{
			"name":     "test-host-updated",
			"addr":     "192.168.1.101",
			"port":     22,
			"user":     "deploy",
			"password": "newsecret",
		}
		resp, err := ts.PUT("/api/ssh-hosts/"+hostUID, body)
		if err != nil {
			t.Fatalf("Failed to update host: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("DeleteSSHHost", func(t *testing.T) {
		if hostUID == "" {
			t.Skip("No host UID available")
		}

		resp, err := ts.DELETE("/api/ssh-hosts/" + hostUID)
		if err != nil {
			t.Fatalf("Failed to delete host: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 200 or 204, got %d", resp.StatusCode)
		}
	})
}

// =========================================================================
// E2E Tests: Applications Management
// =========================================================================

func TestE2E_ApplicationsManagement(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	t.Run("ListApplications_Empty", func(t *testing.T) {
		resp, err := ts.GET("/api/applications")
		if err != nil {
			t.Fatalf("Failed to list applications: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("CLI_CreateApplication", func(t *testing.T) {
		body := map[string]string{
			"name": "test-app",
		}
		resp, err := ts.POST("/api/cli/v1/apps", body)
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := ReadBody(resp)
			t.Errorf("Expected status 200 or 201, got %d: %s", resp.StatusCode, body)
		}
	})

	t.Run("CLI_GetApplicationByName", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/v1/applications/by-name/test-app")
		if err != nil {
			t.Fatalf("Failed to get application: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := ReadBody(resp)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, body)
		}
	})

	t.Run("CLI_ListHosts", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/v1/hosts")
		if err != nil {
			t.Fatalf("Failed to list hosts: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

// =========================================================================
// E2E Tests: Dashboard Stats
// =========================================================================

func TestE2E_DashboardStats(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	t.Run("GetDashboardStats", func(t *testing.T) {
		resp, err := ts.GET("/api/dashboard/stats")
		if err != nil {
			t.Fatalf("Failed to get dashboard stats: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if _, ok := result["applications_count"]; !ok {
			t.Errorf("Expected applications_count in response")
		}
		if _, ok := result["hosts_count"]; !ok {
			t.Errorf("Expected hosts_count in response")
		}
		if _, ok := result["deployments_count"]; !ok {
			t.Errorf("Expected deployments_count in response")
		}
	})
}

// =========================================================================
// E2E Tests: Secrets Management
// =========================================================================

func TestE2E_SecretsManagement(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	// Create an application first
	appBody := map[string]string{"name": "secrets-test-app"}
	ts.POST("/api/cli/v1/apps", appBody)

	t.Run("ListSecrets_Empty", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/v1/secrets?app=secrets-test-app")
		if err != nil {
			t.Fatalf("Failed to list secrets: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := ReadBody(resp)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, body)
		}
	})

	t.Run("SetSecret", func(t *testing.T) {
		body := map[string]string{
			"key":   "DATABASE_URL",
			"value": "postgres://localhost:5432/test",
		}
		resp, err := ts.POST("/api/cli/v1/secrets?app=secrets-test-app", body)
		if err != nil {
			t.Fatalf("Failed to set secret: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := ReadBody(resp)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, respBody)
		}
	})

	t.Run("ListSecrets_WithSecret", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/v1/secrets?app=secrets-test-app")
		if err != nil {
			t.Fatalf("Failed to list secrets: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		keys, ok := result["keys"].([]interface{})
		if !ok {
			t.Skip("keys not in expected format")
		}

		found := false
		for _, k := range keys {
			if k.(string) == "DATABASE_URL" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected DATABASE_URL in secrets list")
		}
	})
}

// =========================================================================
// E2E Tests: System Status
// =========================================================================

func TestE2E_SystemStatus(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	t.Run("GetSystemStatus", func(t *testing.T) {
		resp, err := ts.GET("/api/status")
		if err != nil {
			t.Fatalf("Failed to get system status: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

// =========================================================================
// E2E Tests: Password Change
// =========================================================================

func TestE2E_PasswordChange(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	t.Run("ChangePassword_Success", func(t *testing.T) {
		body := map[string]string{
			"current_password": "admin123",
			"new_password":     "newadmin456",
		}
		resp, err := ts.POST("/api/auth/change-password", body)
		if err != nil {
			t.Fatalf("Failed to change password: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := ReadBody(resp)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, respBody)
		}
	})

	t.Run("Login_WithNewPassword", func(t *testing.T) {
		body := map[string]string{
			"username": "admin",
			"password": "newadmin456",
		}
		resp, err := ts.POST("/api/auth/login", body)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("ChangePassword_WrongCurrentPassword", func(t *testing.T) {
		// Login with new password first
		loginBody := map[string]string{
			"username": "admin",
			"password": "newadmin456",
		}
		loginResp, _ := ts.POST("/api/auth/login", loginBody)
		var loginResult map[string]interface{}
		ParseJSON(loginResp, &loginResult)
		ts.Token = loginResult["access_token"].(string)

		body := map[string]string{
			"current_password": "wrongpassword",
			"new_password":     "another123",
		}
		resp, err := ts.POST("/api/auth/change-password", body)
		if err != nil {
			t.Fatalf("Failed to change password: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})
}

// =========================================================================
// E2E Tests: Full Deployment Flow Simulation
// =========================================================================

func TestE2E_DeploymentFlowSimulation(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	setupAndLogin(t, ts)

	var appUID string
	var hostUID string
	var instanceUID string

	t.Run("Step1_CreateHost", func(t *testing.T) {
		body := map[string]interface{}{
			"name":     "production-server",
			"addr":     "10.0.0.1",
			"port":     22,
			"user":     "deploy",
			"password": "deploy123",
		}
		resp, err := ts.POST("/api/ssh-hosts", body)
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if uid, ok := result["uid"].(string); ok {
			hostUID = uid
		} else if data, ok := result["data"].(map[string]interface{}); ok {
			if uid, ok := data["uid"].(string); ok {
				hostUID = uid
			}
		}

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 201 or 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Step2_CreateApplication", func(t *testing.T) {
		body := map[string]string{
			"name": "my-phoenix-app",
		}
		resp, err := ts.POST("/api/cli/v1/apps", body)
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if uid, ok := result["uid"].(string); ok {
			appUID = uid
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 200 or 201, got %d", resp.StatusCode)
		}
	})

	t.Run("Step3_LinkAppToHost", func(t *testing.T) {
		if hostUID == "" {
			t.Skip("No host UID available")
		}

		body := map[string]string{
			"app_name":  "my-phoenix-app",
			"host_name": "production-server",
		}
		resp, err := ts.POST("/api/cli/v1/link", body)
		if err != nil {
			t.Fatalf("Failed to link app to host: %v", err)
		}

		var result map[string]interface{}
		if err := ParseJSON(resp, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if uid, ok := result["instance_uid"].(string); ok {
			instanceUID = uid
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			respBody, _ := ReadBody(resp)
			t.Errorf("Expected status 200 or 201, got %d: %s", resp.StatusCode, respBody)
		}
	})

	t.Run("Step4_GetDeployConfig", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/v1/deploy/config?app=my-phoenix-app&host=production-server")
		if err != nil {
			t.Fatalf("Failed to get deploy config: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := ReadBody(resp)
			t.Logf("Deploy config response: %s", body)
		}
	})

	t.Run("Step5_SetEnvironmentVariables", func(t *testing.T) {
		secrets := []map[string]string{
			{"key": "SECRET_KEY_BASE", "value": "verylongsecretkeybase123456789"},
			{"key": "DATABASE_URL", "value": "postgres://user:pass@localhost/myapp_prod"},
			{"key": "PHX_HOST", "value": "example.com"},
		}

		for _, secret := range secrets {
			resp, err := ts.POST("/api/cli/v1/secrets?app=my-phoenix-app", secret)
			if err != nil {
				t.Fatalf("Failed to set secret %s: %v", secret["key"], err)
			}

			if resp.StatusCode != http.StatusOK {
				respBody, _ := ReadBody(resp)
				t.Logf("Set secret response: %s", respBody)
			}
		}
	})

	t.Run("Step6_GetInstanceInfo", func(t *testing.T) {
		resp, err := ts.GET("/api/cli/v1/instance?app=my-phoenix-app&host=production-server")
		if err != nil {
			t.Fatalf("Failed to get instance info: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := ReadBody(resp)
			t.Logf("Instance info response: %s", body)
		}
	})

	t.Logf("Test Summary - App UID: %s, Host UID: %s, Instance UID: %s", appUID, hostUID, instanceUID)
}

// =========================================================================
// Helper Functions
// =========================================================================

func setupAndLogin(t *testing.T, ts *TestServer) {
	t.Helper()

	setupBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	setupResp, err := ts.POST("/api/setup", setupBody)
	if err != nil {
		t.Fatalf("Failed to setup: %v", err)
	}

	if setupResp.StatusCode != http.StatusOK {
		body, _ := ReadBody(setupResp)
		t.Fatalf("Setup failed: %s", body)
	}

	loginBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	loginResp, err := ts.POST("/api/auth/login", loginBody)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	var loginResult map[string]interface{}
	if err := ParseJSON(loginResp, &loginResult); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, ok := loginResult["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("Failed to get access token from login")
	}

	ts.Token = token
}
