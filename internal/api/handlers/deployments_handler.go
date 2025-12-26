package handlers

import (
	"fmt"
	"log"
	"os"
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"youfun/shipyard/internal/deploy"

	"github.com/gin-gonic/gin"
)

// Legacy function wrappers for backward compatibility
var defaultDeploymentsRepo = &DefaultRepository{}

// DeploymentResponse represents a deployment in API responses
type DeploymentResponse struct {
	UID       string `json:"uid"`
	AppUID    string `json:"app_uid"`
	AppName   string `json:"app_name"`
	HostUID   string `json:"host_uid"`
	HostName  string `json:"host_name"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	Port      int    `json:"port"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CreateDeploymentRequest represents a deployment creation request from CLI
type CreateDeploymentRequest struct {
	AppName  string `json:"app_name" binding:"required"`
	HostName string `json:"host_name" binding:"required"`
	Version  string `json:"version"`
}

// ListDeployments returns deployments for an application
func ListDeployments(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.ListDeployments(c)
}

// ListDeploymentsHandler returns deployments for an application (method on Handlers)
func (h *Handlers) ListDeployments(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// Get app info
	app, err := h.Repo.GetApplicationByID(appID)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	// Get deployment history
	history, err := h.Repo.GetDeploymentHistoryForApp(app.Name)
	if err != nil {
		response.InternalServerError(c, "Failed to get deployments")
		return
	}

	responses := make([]gin.H, len(history))
	for i, h := range history {
		responses[i] = gin.H{
			"uid":        utils.EncodeFriendlyID(utils.PrefixDeployment, h.ID),
			"version":    h.Version,
			"status":     h.Status,
			"host_name":  h.HostName,
			"port":       h.Port,
			"created_at": h.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Data(c, responses)
}

// GetDeployment returns a specific deployment
func GetDeployment(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.GetDeployment(c)
}

// GetDeploymentHandler returns a specific deployment (method on Handlers)
func (h *Handlers) GetDeployment(c *gin.Context) {
	uid := c.Param("uid")
	deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	history, err := h.Repo.GetDeploymentHistoryByID(deployID)
	if err != nil {
		response.NotFound(c, "Deployment not found")
		return
	}

	response.Data(c, gin.H{
		"uid":        utils.EncodeFriendlyID(utils.PrefixDeployment, history.ID),
		"version":    history.Version,
		"status":     history.Status,
		"host_name":  history.HostName,
		"port":       history.Port,
		"created_at": history.CreatedAt.Format("2006-01-02 15:04:05"),
		"output":     history.Output,
	})
}

// GetDeploymentLogs returns logs for a specific deployment
func GetDeploymentLogs(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.GetDeploymentLogs(c)
}

// GetDeploymentLogsHandler returns logs for a specific deployment (method on Handlers)
func (h *Handlers) GetDeploymentLogs(c *gin.Context) {
	uid := c.Param("uid")
	deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	history, err := h.Repo.GetDeploymentHistoryByID(deployID)
	if err != nil {
		response.NotFound(c, "Deployment not found")
		return
	}

	response.Data(c, gin.H{
		"logs": history.Output,
	})
}

// CreateDeployment creates a new deployment (from CLI)
func CreateDeployment(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.CreateDeployment(c)
}

// CreateDeploymentHandler creates a new deployment (method on Handlers)
func (h *Handlers) CreateDeployment(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// Get application
	app, err := h.Repo.GetApplicationByName(req.AppName)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	// Get host
	host, err := h.Repo.GetSSHHostByName(req.HostName)
	if err != nil {
		response.NotFound(c, "Host not found")
		return
	}

	// Get instance
	instance, _, _, err := h.Repo.GetInstance(req.AppName, req.HostName)
	if err != nil {
		response.NotFound(c, "Application instance not found. Please link the app to the host first.")
		return
	}

	// Create deployment history record
	history, err := h.Repo.CreateDeploymentHistoryWithStatus(instance.ID, req.Version, "pending", "")
	if err != nil {
		response.InternalServerError(c, "Failed to create deployment record")
		return
	}

	// Note: Host credentials are already decrypted by database.GetSSHHostByName

	// Get secrets for the app
	secrets, err := h.Repo.GetAllSecrets(app.ID)
	if err != nil {
		secrets = map[string]string{}
	}

	// Return deployment config for CLI to execute
	response.Created(c, gin.H{
		"deployment_id": utils.EncodeFriendlyID(utils.PrefixDeployment, history.ID),
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
		"instance": gin.H{
			"uid":         utils.EncodeFriendlyID(utils.PrefixAppInstance, instance.ID),
			"active_port": instance.ActivePort.Int64,
		},
		"secrets": secrets,
	})
}

// UploadDeploymentLogsRequest represents the request to upload deployment logs
type UploadDeploymentLogsRequest struct {
	Logs string `json:"logs" binding:"required"`
}

// UploadDeploymentLogs uploads logs for a deployment (from CLI)
func UploadDeploymentLogs(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.UploadDeploymentLogs(c)
}

// UploadDeploymentLogsHandler uploads logs for a deployment (method on Handlers)
func (h *Handlers) UploadDeploymentLogs(c *gin.Context) {
	uid := c.Param("uid")
	deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	var req UploadDeploymentLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	if err := h.Repo.AppendDeploymentHistoryOutput(deployID, req.Logs); err != nil {
		response.InternalServerError(c, "Failed to upload logs")
		return
	}

	response.Message(c, "Logs uploaded successfully")
}

// UpdateDeploymentStatusRequest represents the request to update deployment status
type UpdateDeploymentStatusRequest struct {
	Status       string `json:"status" binding:"required"`
	Port         int    `json:"port"`
	ReleasePath  string `json:"release_path"`
	GitCommitSHA string `json:"git_commit_sha"`
}

// UpdateDeploymentStatus updates deployment status (from CLI)
func UpdateDeploymentStatus(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.UpdateDeploymentStatus(c)
}

// UpdateDeploymentStatusHandler updates deployment status (method on Handlers)
func (h *Handlers) UpdateDeploymentStatus(c *gin.Context) {
	uid := c.Param("uid")
	deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	var req UpdateDeploymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// If status is success and we have port details, perform atomic update
	if req.Status == "success" && req.Port > 0 && req.ReleasePath != "" {
		if err := h.Repo.RecordSuccessfulDeployment(deployID, req.Port, req.ReleasePath, req.GitCommitSHA); err != nil {
			response.InternalServerError(c, "Failed to record successful deployment: "+err.Error())
			return
		}
	} else {
		// Just update the status (e.g. failed)
		if err := h.Repo.UpdateDeploymentHistoryStatusOnly(deployID, req.Status); err != nil {
			response.InternalServerError(c, "Failed to update status")
			return
		}
	}

	response.Message(c, "Status updated successfully")
}

// UploadDeploymentArtifact uploads a build artifact for server-side deployment
func UploadDeploymentArtifact(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.UploadDeploymentArtifact(c)
}

// UploadDeploymentArtifactHandler handles build artifact upload (method on Handlers)
func (h *Handlers) UploadDeploymentArtifact(c *gin.Context) {
	uid := c.Param("uid")
	deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	// Get deployment to verify it exists
	history, err := h.Repo.GetDeploymentHistoryByID(deployID)
	if err != nil {
		response.NotFound(c, "Deployment not found")
		return
	}

	// Receive uploaded file
	file, err := c.FormFile("artifact")
	if err != nil {
		response.BadRequest(c, "Failed to receive artifact file: "+err.Error())
		return
	}

	// Create artifacts directory if it doesn't exist
	artifactsDir := "/var/lib/shipyard/artifacts"
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		response.InternalServerError(c, "Failed to create artifacts directory: "+err.Error())
		return
	}

	// Save file with deployment ID as filename
	artifactPath := fmt.Sprintf("%s/%s.tar.gz", artifactsDir, history.ID)
	if err := c.SaveUploadedFile(file, artifactPath); err != nil {
		response.InternalServerError(c, "Failed to save artifact: "+err.Error())
		return
	}

	log.Printf("✅ Artifact uploaded for deployment %s: %s (size: %d bytes)", uid, artifactPath, file.Size)
	response.Message(c, "Artifact uploaded successfully")
}

// ExecuteDeploymentRequest represents the request to execute server-side deployment
type ExecuteDeploymentRequest struct {
	Version      string `json:"version"`
	GitCommitSHA string `json:"git_commit_sha,omitempty"`
	MD5Hash      string `json:"md5_hash,omitempty"`
}

// ExecuteServerDeployment triggers server-side deployment execution
func ExecuteServerDeployment(c *gin.Context) {
	h := &Handlers{Repo: defaultDeploymentsRepo}
	h.ExecuteServerDeployment(c)
}

// ExecuteServerDeploymentHandler triggers server-side deployment (method on Handlers)
func (h *Handlers) ExecuteServerDeployment(c *gin.Context) {
	uid := c.Param("uid")
	deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	var req ExecuteDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// Get deployment history
	history, err := h.Repo.GetDeploymentHistoryByID(deployID)
	if err != nil {
		response.NotFound(c, "Deployment not found")
		return
	}

	// Get instance details
	instance, err := h.Repo.GetApplicationInstanceByID(history.InstanceID)
	if err != nil {
		response.NotFound(c, "Application instance not found")
		return
	}

	// Get application
	app, err := h.Repo.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	// Verify artifact exists
	artifactPath := fmt.Sprintf("/var/lib/shipyard/artifacts/%s.tar.gz", history.ID)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		response.BadRequest(c, "Artifact not found. Please upload artifact first.")
		return
	}

	// Execute deployment asynchronously
	go func() {
		log.Printf("🚀 Starting server-side deployment for %s (deployment: %s)", app.Name, uid)

		// Call the deploy package function
		if err := deploy.ExecuteServerSideDeployment(deployID.String(), app.Name, req.Version); err != nil {
			log.Printf("❌ Server-side deployment failed: %v", err)
			_ = h.Repo.UpdateDeploymentHistoryStatusOnly(deployID, "failed")
			_ = h.Repo.AppendDeploymentHistoryOutput(deployID, fmt.Sprintf("Deployment failed: %v", err))
		} else {
			log.Printf("✅ Server-side deployment completed successfully")
			// Status is already updated by ExecuteServerSideDeployment
		}
	}()

	response.Message(c, "Server-side deployment started")
}
