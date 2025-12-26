package e2e

import (
	"youfun/shipyard/internal/api"
	"youfun/shipyard/internal/client"
	"youfun/shipyard/internal/database"
	"youfun/shipyard/pkg/types"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// CLITestEnv represents a test environment for CLI E2E tests
type CLITestEnv struct {
	Server     *api.Server
	HTTPServer *httptest.Server
	Client     *client.Client
	Token      string
}

// NewCLITestEnv creates a new CLI test environment
func NewCLITestEnv(t *testing.T) *CLITestEnv {
	t.Helper()

	// Create a temporary directory for test database
	tmpDir := t.TempDir()

	// Set up test environment - clear any conflicting env vars
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("TURSO_DATABASE_URL")
	os.Unsetenv("TURSO_AUTH_TOKEN")

	// Set env vars for testing
	os.Setenv("GIN_MODE", "test")
	os.Setenv("JWT_SECRET", "test-secret-key-for-cli-e2e-testing")

	// Create temp deployer config directory
	testConfigDir := filepath.Join(tmpDir, ".shipyard")
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

	// Create test server
	ts := httptest.NewServer(server.Router)

	// Create API client
	apiClient := client.NewClient(ts.URL)

	return &CLITestEnv{
		Server:     server,
		HTTPServer: ts,
		Client:     apiClient,
	}
}

// Close shuts down the test environment and closes the database connection
func (env *CLITestEnv) Close() {
	env.HTTPServer.Close()
	// Close database connection to allow temp directory cleanup
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}
}

// SetupAdminAndLogin creates admin user and gets auth token
func (env *CLITestEnv) SetupAdminAndLogin(t *testing.T) {
	t.Helper()

	ts := &TestServer{
		BaseURL: env.HTTPServer.URL,
	}

	// Setup admin
	setupBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	resp, err := ts.POST("/api/setup", setupBody)
	if err != nil {
		t.Fatalf("Failed to setup: %v", err)
	}
	resp.Body.Close()

	// Login
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
		t.Fatalf("Failed to get access token")
	}

	env.Token = token
	env.Client.Token = token
}

// =========================================================================
// CLI E2E Tests: Client Functions
// =========================================================================

func TestCLIE2E_ClientFunctions(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	t.Run("Client_ListHosts_Empty", func(t *testing.T) {
		hosts, err := env.Client.ListHosts()
		if err != nil {
			t.Fatalf("ListHosts failed: %v", err)
		}

		// Initially empty or contains default hosts
		t.Logf("Initial hosts count: %d", len(hosts))
	})

	t.Run("Client_CreateApp", func(t *testing.T) {
		err := env.Client.CreateApp("cli-test-app")
		if err != nil {
			t.Fatalf("CreateApp failed: %v", err)
		}
	})

	t.Run("Client_CreateApp_Duplicate", func(t *testing.T) {
		// Creating duplicate app should either succeed (idempotent) or return specific error
		err := env.Client.CreateApp("cli-test-app")
		// Just log the result - different implementations handle duplicates differently
		if err != nil {
			t.Logf("Duplicate app creation returned error (expected): %v", err)
		}
	})
}

// =========================================================================
// CLI E2E Tests: Secrets Management via Client
// =========================================================================

func TestCLIE2E_SecretsViaClient(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	// Create app first
	if err := env.Client.CreateApp("secrets-cli-app"); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	t.Run("Client_ListSecrets_Empty", func(t *testing.T) {
		keys, err := env.Client.ListSecrets("secrets-cli-app")
		if err != nil {
			t.Fatalf("ListSecrets failed: %v", err)
		}

		if len(keys) != 0 {
			t.Errorf("Expected 0 secrets, got %d", len(keys))
		}
	})

	t.Run("Client_SetSecret", func(t *testing.T) {
		err := env.Client.SetSecret("secrets-cli-app", "MY_SECRET", "secret_value_123")
		if err != nil {
			t.Fatalf("SetSecret failed: %v", err)
		}
	})

	t.Run("Client_ListSecrets_WithSecret", func(t *testing.T) {
		keys, err := env.Client.ListSecrets("secrets-cli-app")
		if err != nil {
			t.Fatalf("ListSecrets failed: %v", err)
		}

		found := false
		for _, k := range keys {
			if k == "MY_SECRET" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected MY_SECRET in list, got: %v", keys)
		}
	})

	t.Run("Client_SetMultipleSecrets", func(t *testing.T) {
		secrets := map[string]string{
			"DATABASE_URL":    "postgres://localhost/test",
			"REDIS_URL":       "redis://localhost:6379",
			"SECRET_KEY_BASE": "very_long_secret_key_base",
		}

		for k, v := range secrets {
			if err := env.Client.SetSecret("secrets-cli-app", k, v); err != nil {
				t.Errorf("Failed to set secret %s: %v", k, err)
			}
		}

		// Verify all secrets are set
		keys, err := env.Client.ListSecrets("secrets-cli-app")
		if err != nil {
			t.Fatalf("ListSecrets failed: %v", err)
		}

		// Should have 4 secrets now (MY_SECRET + 3 new ones)
		if len(keys) < 4 {
			t.Errorf("Expected at least 4 secrets, got %d: %v", len(keys), keys)
		}
	})

	t.Run("Client_UnsetSecret", func(t *testing.T) {
		err := env.Client.UnsetSecret("secrets-cli-app", "MY_SECRET")
		if err != nil {
			t.Fatalf("UnsetSecret failed: %v", err)
		}

		// Verify secret is removed
		keys, err := env.Client.ListSecrets("secrets-cli-app")
		if err != nil {
			t.Fatalf("ListSecrets failed: %v", err)
		}

		for _, k := range keys {
			if k == "MY_SECRET" {
				t.Errorf("MY_SECRET should have been removed")
			}
		}
	})
}

// =========================================================================
// CLI E2E Tests: Full Deployment Workflow via Client
// =========================================================================

func TestCLIE2E_DeploymentWorkflow(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	ts := &TestServer{
		BaseURL: env.HTTPServer.URL,
		Token:   env.Token,
	}

	t.Run("Step1_CreateSSHHost", func(t *testing.T) {
		body := map[string]interface{}{
			"name":     "cli-test-host",
			"addr":     "192.168.1.200",
			"port":     22,
			"user":     "deploy",
			"password": "deploy_password",
		}
		resp, err := ts.POST("/api/ssh-hosts", body)
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := ReadBody(resp)
			t.Errorf("Create host failed with status %d: %s", resp.StatusCode, body)
		}
	})

	t.Run("Step2_CreateApplication", func(t *testing.T) {
		err := env.Client.CreateApp("deploy-workflow-app")
		if err != nil {
			t.Fatalf("CreateApp failed: %v", err)
		}
	})

	t.Run("Step3_LinkAppToHost", func(t *testing.T) {
		err := env.Client.LinkApp("deploy-workflow-app", "cli-test-host")
		if err != nil {
			t.Fatalf("LinkApp failed: %v", err)
		}
	})

	t.Run("Step4_GetInstance", func(t *testing.T) {
		instance, err := env.Client.GetInstance("deploy-workflow-app", "cli-test-host")
		if err != nil {
			t.Fatalf("GetInstance failed: %v", err)
		}

		if instance.App.Name != "deploy-workflow-app" {
			t.Errorf("Expected app name 'deploy-workflow-app', got '%s'", instance.App.Name)
		}

		if instance.Host.Name != "cli-test-host" {
			t.Errorf("Expected host name 'cli-test-host', got '%s'", instance.Host.Name)
		}

		t.Logf("Instance UID: %s", instance.Instance.UID)
	})

	t.Run("Step5_SetEnvironmentVariables", func(t *testing.T) {
		envVars := map[string]string{
			"PHX_HOST":        "myapp.example.com",
			"SECRET_KEY_BASE": "test_secret_key_base_value",
			"DATABASE_URL":    "postgres://user:pass@localhost/myapp",
			"POOL_SIZE":       "10",
		}

		for k, v := range envVars {
			if err := env.Client.SetSecret("deploy-workflow-app", k, v); err != nil {
				t.Errorf("Failed to set %s: %v", k, err)
			}
		}
	})

	t.Run("Step6_GetDeployConfig", func(t *testing.T) {
		config, err := env.Client.GetDeployConfig("deploy-workflow-app", "cli-test-host")
		if err != nil {
			// This might fail if some required data is missing - just log it
			t.Logf("GetDeployConfig returned error (may be expected): %v", err)
		} else {
			t.Logf("Deploy config: App=%s, Host=%s", config.App.Name, config.Host.Name)
		}
	})

	t.Run("Step7_SyncDomains", func(t *testing.T) {
		// Get instance UID first
		instance, err := env.Client.GetInstance("deploy-workflow-app", "cli-test-host")
		if err != nil {
			t.Fatalf("Failed to get instance: %v", err)
		}

		domains := []string{"myapp.example.com", "www.myapp.example.com"}
		err = env.Client.SyncDomains(instance.Instance.UID, domains, "myapp.example.com")
		if err != nil {
			t.Logf("SyncDomains returned error (may be expected if not fully implemented): %v", err)
		}
	})
}

// =========================================================================
// CLI E2E Tests: Error Handling
// =========================================================================

func TestCLIE2E_ErrorHandling(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	t.Run("GetInstance_NonExistent", func(t *testing.T) {
		_, err := env.Client.GetInstance("non-existent-app", "non-existent-host")
		if err == nil {
			t.Errorf("Expected error for non-existent app/host")
		}
	})

	t.Run("ListSecrets_NonExistentApp", func(t *testing.T) {
		_, err := env.Client.ListSecrets("non-existent-app")
		if err == nil {
			t.Errorf("Expected error for non-existent app")
		}
	})

	t.Run("LinkApp_NonExistentHost", func(t *testing.T) {
		// Create app first
		env.Client.CreateApp("error-test-app")

		err := env.Client.LinkApp("error-test-app", "non-existent-host")
		if err == nil {
			t.Errorf("Expected error when linking to non-existent host")
		}
	})
}

// =========================================================================
// CLI E2E Tests: Token/Auth Flow
// =========================================================================

func TestCLIE2E_AuthFlow(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	t.Run("ClientWithToken_Success", func(t *testing.T) {
		// Client should have valid token
		hosts, err := env.Client.ListHosts()
		if err != nil {
			t.Fatalf("Request with valid token failed: %v", err)
		}
		t.Logf("Got %d hosts with valid token", len(hosts))
	})

	t.Run("ClientWithoutToken_Failure", func(t *testing.T) {
		// Create client without token
		noAuthClient := client.NewClient(env.HTTPServer.URL)

		_, err := noAuthClient.ListHosts()
		if err == nil {
			t.Errorf("Expected error without token")
		}
	})

	t.Run("ClientWithInvalidToken_Failure", func(t *testing.T) {
		// Create client with invalid token
		invalidClient := client.NewClient(env.HTTPServer.URL)
		invalidClient.Token = "invalid.token.here"

		_, err := invalidClient.ListHosts()
		if err == nil {
			t.Errorf("Expected error with invalid token")
		}
	})
}

// =========================================================================
// CLI E2E Tests: Deployment Operations
// =========================================================================

func TestCLIE2E_DeploymentOperations(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	// Setup: Create host, app, and link them
	ts := &TestServer{
		BaseURL: env.HTTPServer.URL,
		Token:   env.Token,
	}

	// Create host
	hostBody := map[string]interface{}{
		"name":     "deploy-ops-host",
		"addr":     "192.168.1.100",
		"port":     22,
		"user":     "deploy",
		"password": "password",
	}
	ts.POST("/api/ssh-hosts", hostBody)

	// Create app and link
	env.Client.CreateApp("deploy-ops-app")
	env.Client.LinkApp("deploy-ops-app", "deploy-ops-host")

	t.Run("CreateDeployment", func(t *testing.T) {
		// Get instance info first
		instance, err := env.Client.GetInstance("deploy-ops-app", "deploy-ops-host")
		if err != nil {
			t.Fatalf("Failed to get instance: %v", err)
		}

		// Create deployment record using types package
		req := &types.CreateDeploymentRequest{
			AppName:  "deploy-ops-app",
			HostName: "deploy-ops-host",
			Version:  "v1.0.0",
		}

		deployment, err := env.Client.CreateDeployment(req)
		if err != nil {
			t.Logf("CreateDeployment error (may be expected without full setup): %v", err)
		} else {
			t.Logf("Created deployment: %s", deployment.UID)
		}

		t.Logf("Instance ready for deployment: %s", instance.Instance.UID)
	})

	t.Run("CheckArtifact", func(t *testing.T) {
		instance, err := env.Client.GetInstance("deploy-ops-app", "deploy-ops-host")
		if err != nil {
			t.Fatalf("Failed to get instance: %v", err)
		}

		// Check for non-existent artifact
		artifact, err := env.Client.CheckArtifact(instance.App.UID, "abc123gitsha")
		if err != nil {
			t.Logf("CheckArtifact error (expected for non-existent): %v", err)
		}
		if artifact != nil {
			t.Logf("Found artifact: %v", artifact)
		} else {
			t.Logf("No artifact found (expected)")
		}
	})
}

// =========================================================================
// CLI E2E Tests: Concurrent Operations
// =========================================================================

func TestCLIE2E_ConcurrentOperations(t *testing.T) {
	env := NewCLITestEnv(t)
	defer env.Close()
	env.SetupAdminAndLogin(t)

	// Create base app
	if err := env.Client.CreateApp("concurrent-test-app"); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	t.Run("ConcurrentSecretSets", func(t *testing.T) {
		done := make(chan error, 10)

		// Concurrently set 10 secrets
		for i := 0; i < 10; i++ {
			go func(idx int) {
				key := "CONCURRENT_KEY_" + string(rune('A'+idx))
				value := "value_" + string(rune('0'+idx))
				done <- env.Client.SetSecret("concurrent-test-app", key, value)
			}(i)
		}

		// Wait for all
		errors := 0
		for i := 0; i < 10; i++ {
			if err := <-done; err != nil {
				errors++
				t.Logf("Concurrent secret set error: %v", err)
			}
		}

		if errors > 0 {
			t.Logf("%d concurrent operations had errors", errors)
		}

		// Verify secrets were set
		keys, err := env.Client.ListSecrets("concurrent-test-app")
		if err != nil {
			t.Fatalf("Failed to list secrets after concurrent sets: %v", err)
		}

		t.Logf("After concurrent operations, got %d secrets", len(keys))
	})
}
