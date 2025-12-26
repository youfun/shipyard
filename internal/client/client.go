package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"youfun/shipyard/pkg/types"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
)

// APIResponse represents the standard API response envelope from the server
type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Client struct remains unchanged
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient constructor remains unchanged
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// --- core request wrapper ---

// request is the core for all API calls.
// path: path relative to /api/cli/v1/, e.g. "deploy/config" or "deployments/123/status".
// body: optional request body struct; if nil there is no body.
// result: optional pointer to a response struct to decode JSON into; if nil response is ignored.
func (c *Client) request(method, path string, body interface{}, result interface{}) error {
	fullURL := fmt.Sprintf("%s/api/cli/v1/%s", c.BaseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check status code
	// Simplified to treat non-2xx as error. Ideally each method checks specific success codes.
	if resp.StatusCode >= 300 {
		return c.handleError(resp)
	}

	// Decode response body
	if result != nil && resp.StatusCode != http.StatusNoContent {
		// Read the response body into APIResponse envelope
		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Check if the response was successful
		if !apiResp.Success {
			return fmt.Errorf("API request failed: %s", apiResp.Message)
		}

		// Unwrap the data field and decode into the result
		if apiResp.Data != nil && len(apiResp.Data) > 0 {
			if err := json.Unmarshal(apiResp.Data, result); err != nil {
				return fmt.Errorf("failed to decode response data: %w", err)
			}
		}
	}

	return nil
}

// get wraps GET requests
func (c *Client) get(path string, queryParams url.Values, result interface{}) error {
	fullPath := path
	if queryParams != nil {
		fullPath = fmt.Sprintf("%s?%s", path, queryParams.Encode())
	}
	return c.request("GET", fullPath, nil, result)
}

// post wraps POST requests
func (c *Client) post(path string, body interface{}, result interface{}) error {
	return c.request("POST", path, body, result)
}

// put wraps PUT requests
func (c *Client) put(path string, body interface{}, result interface{}) error {
	return c.request("PUT", path, body, result)
}

// delete wraps DELETE requests
func (c *Client) delete(path string, queryParams url.Values) error {
	fullPath := path
	if queryParams != nil {
		fullPath = fmt.Sprintf("%s?%s", path, queryParams.Encode())
	}
	// DELETE typically has no response body, so result is nil
	return c.request("DELETE", fullPath, nil, nil)
}

// GetDeployConfig fetches the deployment configuration for a specific app and host.
func (c *Client) GetDeployConfig(appName, hostName string) (*types.DeployConfigResponse, error) {
	q := url.Values{}
	q.Add("app", appName)
	q.Add("host", hostName)

	var result types.DeployConfigResponse
	if err := c.get("deploy/config", q, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateDeployment creates a new deployment record.
func (c *Client) CreateDeployment(req *types.CreateDeploymentRequest) (*types.DeploymentHistoryDTO, error) {
	// API response is just a { "deployment_id": "..." } structure.
	var response struct {
		DeploymentID string `json:"deployment_id"`
	}

	// Expected status is 201 Created or 200 OK; request wrapper handles that.
	if err := c.post("deployments", req, &response); err != nil {
		return nil, err
	}

	// Return DTO
	return &types.DeploymentHistoryDTO{
		ID: response.DeploymentID,
	}, nil
}

// UpdateDeploymentStatus updates the status of a deployment.
func (c *Client) UpdateDeploymentStatus(deploymentID string, status string, port int, releasePath, gitCommitSHA string) error {
	reqBody := types.UpdateDeploymentStatusRequest{
		Status:       status,
		Port:         port,
		ReleasePath:  releasePath,
		GitCommitSHA: gitCommitSHA,
	}
	path := fmt.Sprintf("deployments/%s/status", deploymentID)
	// PUT request, no response body expected
	return c.put(path, reqBody, nil)
}

// UploadDeploymentLogs uploads logs for a deployment.
func (c *Client) UploadDeploymentLogs(deploymentID string, logs string) error {
	reqBody := types.UploadDeploymentLogsRequest{Logs: logs}
	path := fmt.Sprintf("deployments/%s/logs", deploymentID)
	// POST request, no response body expected
	return c.post(path, reqBody, nil)
}

// UploadDeploymentArtifact uploads a build artifact tarball for server-side deployment.
func (c *Client) UploadDeploymentArtifact(deploymentID string, artifactPath string) error {
	fullURL := fmt.Sprintf("%s/api/cli/v1/deployments/%s/upload", c.BaseURL, deploymentID)

	// Open the artifact file
	file, err := os.Open(artifactPath)
	if err != nil {
		return fmt.Errorf("failed to open artifact file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("artifact", filepath.Base(artifactPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fullURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return c.handleError(resp)
	}

	return nil
}

// ExecuteServerDeployment triggers server-side deployment execution.
func (c *Client) ExecuteServerDeployment(deploymentID string, version, gitCommitSHA, md5Hash string) error {
	reqBody := map[string]string{
		"version":        version,
		"git_commit_sha": gitCommitSHA,
		"md5_hash":       md5Hash,
	}
	path := fmt.Sprintf("deployments/%s/execute", deploymentID)
	// POST request, no response body expected
	return c.post(path, reqBody, nil)
}

// CheckArtifact checks if a build artifact exists.
// The query parameter can be an MD5 prefix (short hash), full MD5 hash, or git commit SHA.
func (c *Client) CheckArtifact(appID, query string) (*types.BuildArtifactDTO, error) {
	q := url.Values{}
	q.Add("app_id", appID)
	q.Add("query", query)

	var result types.BuildArtifactDTO
	err := c.get("artifacts/check", q, &result)

	// Special-case 404 Not Found: treat as "not found" rather than an error
	if httpErr, ok := err.(*APIError); ok && httpErr.StatusCode == http.StatusNotFound {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// RegisterArtifact registers a new build artifact.
func (c *Client) RegisterArtifact(artifact *types.BuildArtifactDTO) error {
	// POST request, no response body expected
	return c.post("artifacts", artifact, nil)
}

func (c *Client) LinkApp(appName, hostName string) error {
	reqBody := types.LinkAppRequest{
		AppName:  appName,
		HostName: hostName,
	}
	// POST request, no response body expected
	return c.post("apps/link", reqBody, nil)
}

func (c *Client) CreateApp(appName string) error {
	reqBody := types.CreateAppRequest{
		Name: appName,
	}
	// POST request, no response body expected
	return c.post("apps", reqBody, nil)
}

func (c *Client) ListHosts() ([]types.SSHHostDTO, error) {
	var hosts []types.SSHHostDTO
	// GET request, returns list
	if err := c.get("hosts", nil, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

// SyncDomains syncs domains from config to the database via API
func (c *Client) SyncDomains(instanceID string, domains []string, primaryDomain string) error {
	reqBody := types.SyncDomainsRequest{
		InstanceID:    instanceID,
		Domains:       domains,
		PrimaryDomain: primaryDomain,
	}
	// POST request, no response body expected
	return c.post("domains/sync", reqBody, nil)
}

// do retains original behavior (adds Auth Header and User-Agent)
func (c *Client) do(req *http.Request) (*http.Response, error) {
	// Add Auth Token if available
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	// Default User-Agent
	req.Header.Set("User-Agent", "Deployer-CLI/1.0")

	return c.HTTPClient.Do(req)
}

// APIError is a custom error type that includes HTTP status code
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API request failed with status %d: %s", e.StatusCode, e.Message)
}

// handleError handles error responses with the new standardized format
func (c *Client) handleError(resp *http.Response) error {
	// Try to read body and parse as APIResponse
	body, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
		// Successfully parsed as APIResponse with a message
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    apiResp.Message,
		}
	}

	// Fallback to raw body if parsing failed
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}

// --- Secrets (Environment Variables) Management ---

// ListSecrets lists all secret keys for an application
func (c *Client) ListSecrets(appName string) ([]string, error) {
	q := url.Values{}
	q.Add("app", appName)

	var result struct {
		App  string   `json:"app"`
		Keys []string `json:"keys"`
	}
	if err := c.get("secrets", q, &result); err != nil {
		return nil, err
	}

	return result.Keys, nil
}

// SetSecret sets a secret for an application
func (c *Client) SetSecret(appName, key, value string) error {
	q := url.Values{}
	q.Add("app", appName)

	reqBody := struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key:   key,
		Value: value,
	}
	// POST request; query params are part of the path
	path := fmt.Sprintf("secrets?%s", q.Encode())
	return c.post(path, reqBody, nil)
}

// UnsetSecret removes a secret for an application
func (c *Client) UnsetSecret(appName, key string) error {
	q := url.Values{}
	q.Add("app", appName)
	q.Add("key", key)

	// DELETE request, query parameters handled uniformly
	return c.delete("secrets", q)
}

// --- Instance Management ---

// InstanceInfo struct remains unchanged
type InstanceInfo struct {
	Instance struct {
		UID                string `json:"uid"`
		Status             string `json:"status"`
		ActivePort         int64  `json:"active_port"`
		PreviousActivePort int64  `json:"previous_active_port"`
	} `json:"instance"`
	App struct {
		UID  string `json:"uid"`
		Name string `json:"name"`
	} `json:"app"`
	Host struct {
		UID        string  `json:"uid"`
		Name       string  `json:"name"`
		Addr       string  `json:"addr"`
		Port       int     `json:"port"`
		User       string  `json:"user"`
		Password   *string `json:"password,omitempty"`
		PrivateKey *string `json:"private_key,omitempty"`
	} `json:"host"`
}

// GetInstance gets instance info for app+host combination
func (c *Client) GetInstance(appName, hostName string) (*InstanceInfo, error) {
	q := url.Values{}
	q.Add("app", appName)
	q.Add("host", hostName)

	var result InstanceInfo
	if err := c.get("instance", q, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetHostByName gets a host by its name
func (c *Client) GetHostByName(hostName string) (*types.SSHHostDTO, error) {
	var result types.SSHHostDTO
	path := fmt.Sprintf("ssh-hosts/by-name/%s", url.PathEscape(hostName))
	if err := c.get(path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetLastDeployment retrieves the last successful deployment for an application.
func (c *Client) GetLastDeployment(appName string) (*types.DeploymentHistoryDTO, error) {
	q := url.Values{}
	q.Add("app_name", appName)

	var result types.DeploymentHistoryDTO
	err := c.get("deployments/latest", q, &result)

	// Special-case 404 Not Found
	if httpErr, ok := err.(*APIError); ok && httpErr.StatusCode == http.StatusNotFound {
		return nil, nil // No deployment found
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListBuildArtifacts lists all build artifacts for an application
func (c *Client) ListBuildArtifacts(appName string) ([]types.BuildArtifactDTO, error) {
	q := url.Values{}
	q.Add("app", appName)

	var result struct {
		App       string                   `json:"app"`
		Artifacts []types.BuildArtifactDTO `json:"artifacts"`
	}
	if err := c.get("builds", q, &result); err != nil {
		return nil, err
	}

	return result.Artifacts, nil
}

// StreamInstanceLogs connects to the WebSocket endpoint and streams logs in real-time
// instanceUID: The unique identifier of the instance (e.g., inst_xxx)
// lines: Number of initial log lines to show
// Returns: error if connection or streaming fails
func (c *Client) StreamInstanceLogs(ctx context.Context, instanceUID string, lines int) error {
	// Build WebSocket URL
	// Convert http/https to ws/wss
	wsURL := strings.Replace(c.BaseURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = fmt.Sprintf("%s/api/instances/%s/logs/stream?lines=%d&token=%s",
		wsURL, instanceUID, lines, c.Token)

	// Create WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("unable to connect to log stream: %w", err)
	}
	defer conn.Close()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Channel to receive errors from the read goroutine
	errChan := make(chan error, 1)

	// Start reading messages in a goroutine
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					errChan <- fmt.Errorf("WebSocket connection closed unexpectedly: %w", err)
				} else {
					errChan <- nil // Normal closure
				}
				return
			}

			// Print log message directly to stdout
			fmt.Print(string(message))
		}
	}()

	// Wait for either context cancellation, signal, or error
	select {
	case <-ctx.Done():
		fmt.Println("\nLog stream stopped")
		return nil
	case <-sigChan:
		fmt.Println("\nInterrupt signal received, exiting...")
		return nil
	case err := <-errChan:
		if err != nil {
			return err
		}
		return nil
	}
}
