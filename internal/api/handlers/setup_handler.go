package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRequest represents the initial setup request
type SetupRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// Legacy function wrappers for backward compatibility
var defaultSetupRepo = &DefaultRepository{}

// SetupStatus checks if initial setup is required
func SetupStatus(c *gin.Context) {
	h := &Handlers{Repo: defaultSetupRepo}
	h.SetupStatus(c)
}

// SetupStatusHandler checks if initial setup is required (method on Handlers)
func (h *Handlers) SetupStatus(c *gin.Context) {
	count, err := h.Repo.GetUserCount()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	response.Data(c, gin.H{
		"setup_required": count == 0,
	})
}

// Setup handles initial admin account setup
func Setup(c *gin.Context) {
	h := &Handlers{Repo: defaultSetupRepo}
	h.Setup(c)
}

// SetupHandler handles initial admin account setup (method on Handlers)
func (h *Handlers) Setup(c *gin.Context) {
	// Check if setup is already complete
	count, err := h.Repo.GetUserCount()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	if count > 0 {
		response.Error(c, http.StatusForbidden, "Setup already completed")
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: username and password (min 6 chars) required")
		return
	}

	// Create the admin user
	tx, err := database.DB.Beginx()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	_, err = h.Repo.CreateUser(tx, req.Username, req.Password)
	if err != nil {
		tx.Rollback()
		response.InternalServerError(c, "Failed to create user")
		return
	}

	if err := tx.Commit(); err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	response.Message(c, "Setup completed successfully")
}
