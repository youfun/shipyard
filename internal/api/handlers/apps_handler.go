package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"youfun/shipyard/internal/database"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Legacy function wrappers for backward compatibility
var defaultAppsRepo = &DefaultRepository{}

// ApplicationResponse represents an application in API responses
type ApplicationResponse struct {
	UID            string `json:"uid"`
	Name           string `json:"name"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	LastDeployedAt string `json:"last_deployed_at,omitempty"`
	LinkedHost     string `json:"linked_host,omitempty"`
}

// ListApplications returns all applications
func ListApplications(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.ListApplications(c)
}

// ListApplicationsHandler returns all applications (method on Handlers)
func (h *Handlers) ListApplications(c *gin.Context) {
	apps, err := h.Repo.GetAllApplications()
	if err != nil {
		response.InternalServerError(c, "Failed to list applications")
		return
	}

	responses := make([]ApplicationResponse, len(apps))
	for i, app := range apps {
		resp := ApplicationResponse{
			UID:  utils.EncodeFriendlyID(utils.PrefixApplication, app.ID),
			Name: app.Name,
		}
		if app.CreatedAt.Time != nil {
			resp.CreatedAt = app.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if app.UpdatedAt.Time != nil {
			resp.UpdatedAt = app.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}

		// Fetch last successful deployment info
		hostName, deployedAt, err := h.Repo.GetLastSuccessfulHostForApp(app.ID)
		if err == nil && hostName != "" {
			resp.LinkedHost = hostName
			if !deployedAt.IsZero() {
				resp.LastDeployedAt = deployedAt.Format("2006-01-02 15:04:05")
			}
		}

		responses[i] = resp
	}

	response.Data(c, responses)
}

// GetApplication returns a specific application with details
func GetApplication(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.GetApplication(c)
}

// GetApplicationHandler returns a specific application with details (method on Handlers)
func (h *Handlers) GetApplication(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	app, err := h.Repo.GetApplicationByID(appID)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	// Get linked instances
	instances, err := h.Repo.GetApplicationInstances(appID)
	if err != nil {
		instances = []database.ApplicationInstanceRow{}
	}

	// Build response with instances
	instanceResponses := make([]gin.H, len(instances))
	var primaryDomain string
	var targetPort int64
	var appStatus string = "unknown"

	for i, inst := range instances {
		// Get host info
		host, _ := h.Repo.GetSSHHostByID(inst.HostID)
		hostName := ""
		hostAddr := ""
		if host != nil {
			hostName = host.Name
			hostAddr = host.Addr
		}

		instanceResponses[i] = gin.H{
			"uid":         utils.EncodeFriendlyID(utils.PrefixAppInstance, inst.ID),
			"host_id":     utils.EncodeFriendlyID(utils.PrefixSSHHost, inst.HostID),
			"host_name":   hostName,
			"host_addr":   hostAddr,
			"status":      inst.Status,
			"active_port": inst.ActivePort.Int64,
		}

		// Get first instance's domain and port for overview display
		if i == 0 {
			if inst.ActivePort.Valid && inst.ActivePort.Int64 > 0 {
				targetPort = inst.ActivePort.Int64
			}
			appStatus = inst.Status

			// Get domains for this instance
			domains, err := h.Repo.GetDomainsForInstance(inst.ID)
			if err == nil && len(domains) > 0 {
				primaryDomain = domains[0].Hostname
			}
		}
	}

	resp := gin.H{
		"uid":            utils.EncodeFriendlyID(utils.PrefixApplication, app.ID),
		"name":           app.Name,
		"instances":      instanceResponses,
		"primary_domain": primaryDomain,
		"active_port":    targetPort,
		"status":         appStatus,
	}

	// Get last successful deployment info
	_, deployedAt, err := h.Repo.GetLastSuccessfulHostForApp(appID)
	if err == nil && !deployedAt.IsZero() {
		resp["last_deployed_at"] = deployedAt.Format("2006-01-02 15:04:05")
	}

	if app.CreatedAt.Time != nil {
		resp["created_at"] = app.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if app.UpdatedAt.Time != nil {
		resp["updated_at"] = app.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}

	response.Data(c, resp)
}

// StartApplication starts an application
func StartApplication(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.StartApplication(c)
}

// StartApplicationHandler starts an application (method on Handlers)
func (h *Handlers) StartApplication(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// TODO: Implement actual start logic via SSH
	_ = appID

	response.Message(c, "Application start requested")
}

// StopApplication stops an application
func StopApplication(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.StopApplication(c)
}

// StopApplicationHandler stops an application (method on Handlers)
func (h *Handlers) StopApplication(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// TODO: Implement actual stop logic via SSH
	_ = appID

	response.Message(c, "Application stop requested")
}

// RestartApplication restarts an application
func RestartApplication(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.RestartApplication(c)
}

// RestartApplicationHandler restarts an application (method on Handlers)
func (h *Handlers) RestartApplication(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// TODO: Implement actual restart logic via SSH
	_ = appID

	response.Message(c, "Application restart requested")
}

// GetApplicationLogs returns application logs
func GetApplicationLogs(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.GetApplicationLogs(c)
}

// GetApplicationLogsHandler returns application logs (method on Handlers)
func (h *Handlers) GetApplicationLogs(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// TODO: Implement actual log retrieval via SSH
	_ = appID

	response.Data(c, gin.H{
		"logs": []string{
			"Application logs will be available here",
		},
	})
}

// --- Environment Variables (Secrets) Management for WebUI ---

// EnvironmentVariableResponse represents an environment variable in API responses
type EnvironmentVariableResponse struct {
	UID         string `json:"uid"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsEncrypted bool   `json:"isEncrypted"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// CreateEnvironmentVariableRequest represents the request to create/update an environment variable
type CreateEnvironmentVariableRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	IsEncrypted bool   `json:"isEncrypted"`
}

// ListEnvironmentVariables returns all environment variables for an application
func ListEnvironmentVariables(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.ListEnvironmentVariables(c)
}

// ListEnvironmentVariablesHandler returns all environment variables for an application (method on Handlers)
func (h *Handlers) ListEnvironmentVariables(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// Get secret keys only (values are encrypted, we'll show masked values for encrypted ones)
	keys, err := h.Repo.ListSecretKeys(appID)
	if err != nil {
		response.InternalServerError(c, "Failed to list environment variables: "+err.Error())
		return
	}

	// Get all secrets with decrypted values for display
	secrets, err := h.Repo.GetSecretsForApp(appID)
	if err != nil {
		response.InternalServerError(c, "Failed to get environment variables: "+err.Error())
		return
	}

	// Build response - treat all secrets as encrypted for now (since they're stored encrypted)
	responses := make([]EnvironmentVariableResponse, 0, len(keys))
	for _, key := range keys {
		value := ""
		if v, ok := secrets[key]; ok {
			value = v
		}
		responses = append(responses, EnvironmentVariableResponse{
			UID:         encodeEnvVarUID(appID, key),
			Key:         key,
			Value:       value,
			IsEncrypted: true, // All secrets are encrypted in storage
		})
	}

	response.Data(c, responses)
}

// CreateEnvironmentVariable creates a new environment variable for an application
func CreateEnvironmentVariable(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.CreateEnvironmentVariable(c)
}

// CreateEnvironmentVariableHandler creates a new environment variable for an application (method on Handlers)
func (h *Handlers) CreateEnvironmentVariable(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	var req CreateEnvironmentVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.Repo.SetSecret(appID, req.Key, req.Value); err != nil {
		response.InternalServerError(c, "Failed to create environment variable: "+err.Error())
		return
	}

	response.Created(c, gin.H{
		"uid":         encodeEnvVarUID(appID, req.Key),
		"key":         req.Key,
		"isEncrypted": true,
	})
}

// UpdateEnvironmentVariable updates an existing environment variable
func UpdateEnvironmentVariable(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.UpdateEnvironmentVariable(c)
}

// UpdateEnvironmentVariableHandler updates an existing environment variable (method on Handlers)
func (h *Handlers) UpdateEnvironmentVariable(c *gin.Context) {
	envVarID := c.Param("envVarId")

	// Parse the composite UID
	appID, oldKey, err := decodeEnvVarUID(envVarID)
	if err != nil {
		response.BadRequest(c, "Invalid environment variable ID")
		return
	}

	var req CreateEnvironmentVariableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// If key changed, delete old and create new
	if oldKey != req.Key {
		if err := h.Repo.UnsetSecret(appID, oldKey); err != nil {
			response.InternalServerError(c, "Failed to update environment variable: "+err.Error())
			return
		}
	}

	if err := h.Repo.SetSecret(appID, req.Key, req.Value); err != nil {
		response.InternalServerError(c, "Failed to update environment variable: "+err.Error())
		return
	}

	response.Data(c, gin.H{
		"uid":         encodeEnvVarUID(appID, req.Key),
		"key":         req.Key,
		"isEncrypted": true,
	})
}

// DeleteEnvironmentVariable deletes an environment variable
func DeleteEnvironmentVariable(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.DeleteEnvironmentVariable(c)
}

// DeleteEnvironmentVariableHandler deletes an environment variable (method on Handlers)
func (h *Handlers) DeleteEnvironmentVariable(c *gin.Context) {
	envVarID := c.Param("envVarId")

	// Parse the composite UID
	appID, key, err := decodeEnvVarUID(envVarID)
	if err != nil {
		response.BadRequest(c, "Invalid environment variable ID")
		return
	}

	if err := h.Repo.UnsetSecret(appID, key); err != nil {
		response.InternalServerError(c, "Failed to delete environment variable: "+err.Error())
		return
	}

	response.Message(c, "Environment variable deleted successfully")
}

// envVarUIDDelimiter is used to separate the appID prefix from the key in composite UIDs
// Using "::" as it's unlikely to appear in environment variable keys
const envVarUIDDelimiter = "::"

// encodeEnvVarUID creates a composite UID for an environment variable
// Format: env_<base58(appID)>::<key>
func encodeEnvVarUID(appID uuid.UUID, key string) string {
	return utils.EncodeFriendlyID(utils.PrefixEnvVar, appID) + envVarUIDDelimiter + key
}

// decodeEnvVarUID parses a composite environment variable UID
// Returns the appID and key
func decodeEnvVarUID(uid string) (uuid.UUID, string, error) {
	// Find the delimiter
	delimIndex := strings.Index(uid, envVarUIDDelimiter)
	if delimIndex == -1 {
		return uuid.Nil, "", errors.New("invalid environment variable UID format")
	}

	prefix := uid[:delimIndex]
	key := uid[delimIndex+len(envVarUIDDelimiter):]

	if key == "" {
		return uuid.Nil, "", errors.New("empty key in environment variable UID")
	}

	appID, err := utils.DecodeFriendlyID(utils.PrefixEnvVar, prefix)
	if err != nil {
		return uuid.Nil, "", err
	}

	return appID, key, nil
}
