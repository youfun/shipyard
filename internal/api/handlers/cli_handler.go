package handlers

import (
	"net/http"
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// CLI-specific handlers for deployer-cli client

// Legacy function wrappers for backward compatibility
var defaultCLIRepo = &DefaultRepository{}

// CreateAppRequest represents the request to create an application from CLI
type CreateAppRequest struct {
	Name string `json:"name" binding:"required"`
}

// CLICreateApplication creates a new application (CLI endpoint)
func CLICreateApplication(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLICreateApplication(c)
}

// CLICreateApplicationHandler creates a new application (CLI endpoint) (method on Handlers)
func (h *Handlers) CLICreateApplication(c *gin.Context) {
	var req CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: name is required")
		return
	}

	// Check if app already exists
	existingApp, err := h.Repo.GetApplicationByName(req.Name)
	if err == nil && existingApp != nil {
		response.Data(c, gin.H{
			"uid":  utils.EncodeFriendlyID(utils.PrefixApplication, existingApp.ID),
			"name": existingApp.Name,
		})
		return
	}

	// Create new application
	newApp := &models.Application{Name: req.Name}
	if err := h.Repo.AddApplication(newApp); err != nil {
		response.InternalServerError(c, "Failed to create application: "+err.Error())
		return
	}

	response.Created(c, gin.H{
		"uid":  utils.EncodeFriendlyID(utils.PrefixApplication, newApp.ID),
		"name": newApp.Name,
	})
}

// CLIGetApplicationByName gets an application by name (CLI endpoint)
func CLIGetApplicationByName(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIGetApplicationByName(c)
}

// CLIGetApplicationByNameHandler gets an application by name (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIGetApplicationByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.BadRequest(c, "Application name is required")
		return
	}

	app, err := h.Repo.GetApplicationByName(name)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	resp := gin.H{
		"uid":  utils.EncodeFriendlyID(utils.PrefixApplication, app.ID),
		"name": app.Name,
	}

	if app.CreatedAt.Time != nil {
		resp["created_at"] = app.CreatedAt.Time.Format(time.RFC3339)
	}

	// Get linked host info
	hostName, deployedAt, err := h.Repo.GetLastSuccessfulHostForApp(app.ID)
	if err == nil && hostName != "" {
		resp["linked_host"] = hostName
		if !deployedAt.IsZero() {
			resp["last_deployed_at"] = deployedAt.Format(time.RFC3339)
		}
	}

	response.Data(c, resp)
}

// CLIGetSSHHostByName gets an SSH host by name (CLI endpoint)
func CLIGetSSHHostByName(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIGetSSHHostByName(c)
}

// CLIGetSSHHostByNameHandler gets an SSH host by name (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIGetSSHHostByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.BadRequest(c, "Host name is required")
		return
	}

	host, err := h.Repo.GetSSHHostByName(name)
	if err != nil {
		response.NotFound(c, "Host not found")
		return
	}

	resp := gin.H{
		"uid":    utils.EncodeFriendlyID(utils.PrefixSSHHost, host.ID),
		"name":   host.Name,
		"addr":   host.Addr,
		"port":   host.Port,
		"user":   host.User,
		"status": host.Status,
		"arch":   host.Arch,
	}

	if host.InitializedAt.Time != nil {
		resp["initialized_at"] = host.InitializedAt.Time.Format(time.RFC3339)
	}

	response.Data(c, resp)
}

// LinkAppRequest represents the request to link an app to a host
type LinkAppToHostRequest struct {
	AppName  string `json:"app_name" binding:"required"`
	HostName string `json:"host_name" binding:"required"`
}

// CLILinkAppToHost links an application to a host (CLI endpoint)
func CLILinkAppToHost(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLILinkAppToHost(c)
}

// CLILinkAppToHostHandler links an application to a host (CLI endpoint) (method on Handlers)
func (h *Handlers) CLILinkAppToHost(c *gin.Context) {
	var req LinkAppToHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: app_name and host_name are required")
		return
	}

	// Get application
	app, err := h.Repo.GetApplicationByName(req.AppName)
	if err != nil {
		response.NotFound(c, "Application not found: "+req.AppName)
		return
	}

	// Get host
	host, err := h.Repo.GetSSHHostByName(req.HostName)
	if err != nil {
		response.NotFound(c, "Host not found: "+req.HostName)
		return
	}

	// Check if instance already exists
	existingInstance, err := h.Repo.GetApplicationInstance(app.ID, host.ID)
	if err == nil && existingInstance != nil {
		response.Data(c, gin.H{
			"instance_uid": utils.EncodeFriendlyID(utils.PrefixAppInstance, existingInstance.ID),
		})
		return
	}

	// Create new instance link
	newInstance := &models.ApplicationInstance{
		ApplicationID: app.ID,
		HostID:        host.ID,
		Status:        "linked",
	}
	if err := h.Repo.LinkApplicationToHost(newInstance); err != nil {
		response.InternalServerError(c, "Failed to link application: "+err.Error())
		return
	}

	response.Created(c, gin.H{
		"instance_uid": utils.EncodeFriendlyID(utils.PrefixAppInstance, newInstance.ID),
	})
}

// CLIListHosts lists all available SSH hosts for CLI selection
func CLIListHosts(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIListHosts(c)
}

// CLIListHostsHandler lists all available SSH hosts for CLI selection (method on Handlers)
func (h *Handlers) CLIListHosts(c *gin.Context) {
	hosts, err := h.Repo.GetAllSSHHosts()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch hosts: "+err.Error())
		return
	}

	var resp []gin.H
	for _, host := range hosts {
		hostMap := gin.H{
			"uid":    utils.EncodeFriendlyID(utils.PrefixSSHHost, host.ID),
			"name":   host.Name,
			"addr":   host.Addr,
			"port":   host.Port,
			"user":   host.User,
			"status": host.Status,
			"arch":   host.Arch,
		}
		if host.InitializedAt.Time != nil {
			hostMap["initialized_at"] = host.InitializedAt.Time.Format(time.RFC3339)
		}

		// Host credentials are already decrypted by database.GetAllSSHHosts
		if host.Password != nil {
			hostMap["password"] = *host.Password
		}
		if host.PrivateKey != nil {
			hostMap["private_key"] = *host.PrivateKey
		}

		resp = append(resp, hostMap)
	}

	response.Data(c, resp)
}

// CLIGetInstance gets instance info for app+host combination
func CLIGetInstance(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIGetInstance(c)
}

// CLIGetInstanceHandler gets instance info for app+host combination (method on Handlers)
func (h *Handlers) CLIGetInstance(c *gin.Context) {
	appName := c.Query("app")
	hostName := c.Query("host")

	if appName == "" || hostName == "" {
		response.BadRequest(c, "app and host query parameters are required")
		return
	}

	instance, app, host, err := h.Repo.GetInstance(appName, hostName)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Data(c, gin.H{
		"instance": gin.H{
			"uid":                  utils.EncodeFriendlyID(utils.PrefixAppInstance, instance.ID),
			"status":               instance.Status,
			"active_port":          instance.ActivePort.Int64,
			"previous_active_port": instance.PreviousActivePort.Int64,
		},
		"app": gin.H{
			"uid":  utils.EncodeFriendlyID(utils.PrefixApplication, app.ID),
			"name": app.Name,
		},
		"host": gin.H{
			"uid":         utils.EncodeFriendlyID(utils.PrefixSSHHost, host.ID),
			"name":        host.Name,
			"addr":        host.Addr,
			"port":        host.Port,
			"user":        host.User,
			"password":    host.Password,
			"private_key": host.PrivateKey,
		},
	})
}

// CLIGetDeployConfig aggregates all necessary config for a deployment (CLI endpoint)
func CLIGetDeployConfig(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIGetDeployConfig(c)
}

// CLIGetDeployConfigHandler aggregates all necessary config for a deployment (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIGetDeployConfig(c *gin.Context) {
	appName := c.Query("app")
	hostName := c.Query("host")

	if appName == "" || hostName == "" {
		response.BadRequest(c, "app and host query parameters are required")
		return
	}

	// Check if this is localhost deployment (server-side deployment mode)
	isLocalhost := hostName == "localhost" || hostName == "127.0.0.1" || hostName == "local"

	var instance *models.ApplicationInstance
	var app *models.Application
	var host *models.SSHHost
	var err error

	if isLocalhost {
		// For localhost deployment, ensure host and instance exist (auto-create if needed)
		instance, app, host, err = h.Repo.EnsureLocalhostInstance(appName)
	} else {
		// For regular SSH deployment, require existing instance
		instance, app, host, err = h.Repo.GetInstance(appName, hostName)
	}

	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	// 2. Get Secrets
	secrets, err := h.Repo.GetSecretsForApp(app.ID)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch secrets: "+err.Error())
		return
	}

	// 3. Get Domains
	domains, err := h.Repo.GetDomainsForInstance(instance.ID)
	if err != nil {
		// Domains are optional for deployment config, log error but continue with empty list
		domains = nil
	}
	var domainList []string
	for _, d := range domains {
		domainList = append(domainList, d.Hostname)
	}

	// 4. Construct Response using gin.H for consistency with other handlers

	resp := gin.H{
		"app": gin.H{
			"id":   app.ID.String(), // Raw UUID for client parsing
			"uid":  utils.EncodeFriendlyID(utils.PrefixApplication, app.ID),
			"name": app.Name,
		},
		"host": gin.H{
			"id":          host.ID.String(), // Raw UUID for client parsing
			"uid":         utils.EncodeFriendlyID(utils.PrefixSSHHost, host.ID),
			"name":        host.Name,
			"addr":        host.Addr,
			"port":        host.Port,
			"user":        host.User,
			"password":    host.Password,
			"private_key": host.PrivateKey,
		},
		"instance": gin.H{
			"id":                   instance.ID.String(), // Raw UUID for client parsing
			"uid":                  utils.EncodeFriendlyID(utils.PrefixAppInstance, instance.ID),
			"application_id":       instance.ApplicationID.String(), // Raw UUID for client parsing
			"host_id":              instance.HostID.String(),        // Raw UUID for client parsing
			"status":               instance.Status,
			"active_port":          instance.ActivePort.Int64,
			"previous_active_port": instance.PreviousActivePort.Int64,
		},
		"secrets": secrets,
		"domains": domainList,
	}

	response.Data(c, resp)
}

// CLICheckArtifact checks if a build artifact exists (CLI endpoint)
func CLICheckArtifact(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLICheckArtifact(c)
}

// CLICheckArtifactHandler checks if a build artifact exists (CLI endpoint) (method on Handlers)
// Supports query by: md5_prefix (short MD5 hash), git_sha (full git commit SHA), or query (auto-detect)
func (h *Handlers) CLICheckArtifact(c *gin.Context) {
	appID := c.Query("app_id")
	md5Prefix := c.Query("md5_prefix")
	gitSHA := c.Query("git_sha")
	query := c.Query("query") // Generic query that tries both

	if appID == "" {
		response.BadRequest(c, "app_id is required")
		return
	}

	if md5Prefix == "" && gitSHA == "" && query == "" {
		response.BadRequest(c, "One of md5_prefix, git_sha, or query is required")
		return
	}

	// Convert friendly ID back to UUID (support both friendly ID and raw UUID)
	appUUID, err := utils.DecodeFriendlyID(utils.PrefixApplication, appID)
	if err != nil {
		// Try parsing as raw UUID
		appUUID, err = utils.ParseUUID(appID)
		if err != nil {
			response.BadRequest(c, "Invalid app_id")
			return
		}
	}

	var artifact *models.BuildArtifact

	// Priority: md5_prefix > git_sha > query (try md5 first, then git_sha)
	if md5Prefix != "" {
		artifact, err = h.Repo.GetBuildArtifactByMD5Prefix(appUUID, md5Prefix)
	} else if gitSHA != "" {
		artifact, err = h.Repo.GetBuildArtifactByGitSHA(appUUID, gitSHA)
	} else if query != "" {
		// Auto-detect: try MD5 prefix first, then git SHA
		artifact, err = h.Repo.GetBuildArtifactByMD5Prefix(appUUID, query)
		if err != nil {
			// Try git SHA
			artifact, err = h.Repo.GetBuildArtifactByGitSHA(appUUID, query)
		}
	}

	if err != nil || artifact == nil {
		response.NotFound(c, "Artifact not found")
		return
	}

	response.Data(c, gin.H{
		"id":             utils.EncodeFriendlyID(utils.PrefixBuildArtifact, artifact.ID),
		"application_id": utils.EncodeFriendlyID(utils.PrefixApplication, artifact.ApplicationID),
		"git_commit_sha": artifact.GitCommitSHA,
		"version":        artifact.Version,
		"md5_hash":       artifact.MD5Hash,
		"local_path":     artifact.LocalPath,
		"created_at":     artifact.CreatedAt.Time.Format(time.RFC3339),
	})
}

// RegisterArtifactRequest represents request to register artifact
type RegisterArtifactRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	GitCommitSHA  string `json:"git_commit_sha"`
	Version       string `json:"version" binding:"required"`
	MD5Hash       string `json:"md5_hash" binding:"required"`
	LocalPath     string `json:"local_path" binding:"required"`
}

// CLIRegisterArtifact registers a new build artifact (CLI endpoint)
func CLIRegisterArtifact(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIRegisterArtifact(c)
}

// CLIRegisterArtifactHandler registers a new build artifact (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIRegisterArtifact(c *gin.Context) {
	var req RegisterArtifactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Support both friendly ID (app_xxx) and raw UUID
	appUUID, err := utils.DecodeFriendlyID(utils.PrefixApplication, req.ApplicationID)
	if err != nil {
		// Try parsing as raw UUID
		appUUID, err = utils.ParseUUID(req.ApplicationID)
		if err != nil {
			response.BadRequest(c, "Invalid application_id")
			return
		}
	}

	artifact := &models.BuildArtifact{
		ApplicationID: appUUID,
		GitCommitSHA:  req.GitCommitSHA,
		Version:       req.Version,
		MD5Hash:       req.MD5Hash,
		LocalPath:     req.LocalPath,
	}

	if err := h.Repo.AddBuildArtifact(artifact); err != nil {
		response.InternalServerError(c, "Failed to register artifact: "+err.Error())
		return
	}

	response.Created(c, gin.H{
		"id": utils.EncodeFriendlyID(utils.PrefixBuildArtifact, artifact.ID),
	})
}

// --- Environment Variables (Secrets) Management ---

// SetSecretRequest represents the request to set a secret
type SetSecretRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// CLIListSecrets lists all secret keys for an application (CLI endpoint)
func CLIListSecrets(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIListSecrets(c)
}

// CLIListSecretsHandler lists all secret keys for an application (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIListSecrets(c *gin.Context) {
	appName := c.Query("app")
	if appName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app query parameter is required"})
		return
	}

	app, err := h.Repo.GetApplicationByName(appName)
	if err != nil {
		response.NotFound(c, "Application not found: "+appName)
		return
	}

	keys, err := h.Repo.ListSecretKeys(app.ID)
	if err != nil {
		response.InternalServerError(c, "Failed to list secrets: "+err.Error())
		return
	}

	response.Data(c, gin.H{
		"app":  appName,
		"keys": keys,
	})
}

// CLISetSecret sets a secret for an application (CLI endpoint)
func CLISetSecret(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLISetSecret(c)
}

// CLISetSecretHandler sets a secret for an application (CLI endpoint) (method on Handlers)
func (h *Handlers) CLISetSecret(c *gin.Context) {
	appName := c.Query("app")
	if appName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app query parameter is required"})
		return
	}

	var req SetSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: key and value are required")
		return
	}

	app, err := h.Repo.GetApplicationByName(appName)
	if err != nil {
		response.NotFound(c, "Application not found: "+appName)
		return
	}

	if err := h.Repo.SetSecret(app.ID, req.Key, req.Value); err != nil {
		response.InternalServerError(c, "Failed to set secret: "+err.Error())
		return
	}

	response.Data(c, gin.H{
		"message": "Secret set successfully",
		"app":     appName,
		"key":     req.Key,
	})
}

// CLIUnsetSecret removes a secret for an application (CLI endpoint)
func CLIUnsetSecret(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIUnsetSecret(c)
}

// CLIUnsetSecretHandler removes a secret for an application (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIUnsetSecret(c *gin.Context) {
	appName := c.Query("app")
	key := c.Query("key")

	if appName == "" || key == "" {
		response.BadRequest(c, "app and key query parameters are required")
		return
	}

	app, err := h.Repo.GetApplicationByName(appName)
	if err != nil {
		response.NotFound(c, "Application not found: "+appName)
		return
	}

	if err := h.Repo.UnsetSecret(app.ID, key); err != nil {
		response.InternalServerError(c, "Failed to unset secret: "+err.Error())
		return
	}

	response.Data(c, gin.H{
		"message": "Secret removed successfully",
		"app":     appName,
		"key":     key,
	})
}

// --- Domains Sync Management ---

// CLISyncDomains syncs domains from config to database (CLI endpoint)
func CLISyncDomains(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLISyncDomains(c)
}

// CLISyncDomainsHandler syncs domains from config to database (CLI endpoint) (method on Handlers)
func (h *Handlers) CLISyncDomains(c *gin.Context) {
	var req struct {
		InstanceID    string   `json:"instance_id" binding:"required"`
		Domains       []string `json:"domains" binding:"required"`
		PrimaryDomain string   `json:"primary_domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: instance_id and domains are required")
		return
	}

	// Parse instance ID
	instanceID, err := utils.DecodeFriendlyID(utils.PrefixAppInstance, req.InstanceID)
	if err != nil {
		// Try parsing as raw UUID
		instanceID, err = utils.ParseUUID(req.InstanceID)
		if err != nil {
			response.BadRequest(c, "Invalid instance_id")
			return
		}
	}

	// Get existing domains
	existingDomains, err := h.Repo.GetDomainsForInstance(instanceID)
	if err != nil {
		response.InternalServerError(c, "Failed to get existing domains: "+err.Error())
		return
	}

	// Create a map for existing domains
	existingMap := make(map[string]bool)
	for _, d := range existingDomains {
		existingMap[d.Hostname] = true
	}

	// Add new domains
	addedCount := 0
	for _, hostname := range req.Domains {
		if !existingMap[hostname] {
			isPrimary := hostname == req.PrimaryDomain
			domain := &models.Domain{
				ApplicationInstanceID: instanceID,
				Hostname:              hostname,
				IsPrimary:             isPrimary,
			}
			if err := h.Repo.AddDomain(domain); err != nil {
				response.InternalServerError(c, "Failed to add domain: "+err.Error())
				return
			}
			addedCount++
		}
	}

	// Update primary domain if specified
	if req.PrimaryDomain != "" {
		// Check if primary domain is in the list
		for _, hostname := range req.Domains {
			if hostname == req.PrimaryDomain {
				if err := h.Repo.SetPrimaryDomain(instanceID, req.PrimaryDomain); err != nil {
					response.InternalServerError(c, "Failed to set primary domain: "+err.Error())
					return
				}
				break
			}
		}
	}

	response.Data(c, gin.H{
		"message":     "Domains synced successfully",
		"added_count": addedCount,
	})
}

// CLIGetLastDeployment retrieves the last successful deployment for an application (CLI endpoint)
func CLIGetLastDeployment(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIGetLastDeployment(c)
}

// CLIGetLastDeploymentHandler retrieves the last successful deployment for an application (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIGetLastDeployment(c *gin.Context) {
	appName := c.Query("app_name")
	if appName == "" {
		response.BadRequest(c, "app_name is required")
		return
	}

	// Use GetLatestDeploymentHistoryForApp logic but we need full deployment info

	history, err := h.Repo.GetLatestDeploymentHistoryForApp(appName)
	if err != nil {
		response.NotFound(c, "No deployment history found")
		return
	}
	if history == nil {
		response.NotFound(c, "No deployment history found")
		return
	}

	// We need to fetch the host name for this deployment to show it in CLI
	// Let's rely on GetDeploymentHistoryByID which returns a DeploymentHistoryRow with HostName.
	row, err := h.Repo.GetDeploymentHistoryByID(history.ID)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch deployment details")
		return
	}

	response.Data(c, gin.H{
		"id":          utils.EncodeFriendlyID(utils.PrefixDeployment, row.ID),
		"host_name":   row.HostName,
		"deployed_at": row.CreatedAt.Format(time.RFC3339),
		"status":      row.Status,
	})
}

// --- Build Artifacts Management ---

// CLIListBuildArtifacts lists all build artifacts for an application (CLI endpoint)
func CLIListBuildArtifacts(c *gin.Context) {
	h := &Handlers{Repo: defaultCLIRepo}
	h.CLIListBuildArtifacts(c)
}

// CLIListBuildArtifactsHandler lists all build artifacts for an application (CLI endpoint) (method on Handlers)
func (h *Handlers) CLIListBuildArtifacts(c *gin.Context) {
	appName := c.Query("app")
	if appName == "" {
		response.BadRequest(c, "app query parameter is required")
		return
	}

	app, err := h.Repo.GetApplicationByName(appName)
	if err != nil {
		response.NotFound(c, "Application not found: "+appName)
		return
	}

	artifacts, err := h.Repo.GetAllBuildArtifactsForApp(app.ID)
	if err != nil {
		response.InternalServerError(c, "Failed to list build artifacts: "+err.Error())
		return
	}

	var resp []gin.H
	for _, artifact := range artifacts {
		item := gin.H{
			"id":             utils.EncodeFriendlyID(utils.PrefixBuildArtifact, artifact.ID),
			"version":        artifact.Version,
			"git_commit_sha": artifact.GitCommitSHA,
			"md5_hash":       artifact.MD5Hash,
		}
		if artifact.CreatedAt.Time != nil {
			item["created_at"] = artifact.CreatedAt.Time.Format(time.RFC3339)
		}
		resp = append(resp, item)
	}

	response.Data(c, gin.H{
		"app":       appName,
		"artifacts": resp,
	})
}
