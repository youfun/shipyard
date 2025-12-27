package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"youfun/shipyard/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Legacy function wrappers for backward compatibility
var defaultRoutingsRepo = &DefaultRepository{}

// RoutingResponse represents a routing/domain in API responses
type RoutingResponse struct {
	UID        string `json:"uid"`
	DomainName string `json:"domainName"`
	HostPort   int    `json:"hostPort"`
	IsActive   bool   `json:"isActive"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// RoutingRequest represents the request to create/update a routing
type RoutingRequest struct {
	DomainName string `json:"domainName" binding:"required"`
	HostPort   int    `json:"hostPort" binding:"required"`
	IsActive   bool   `json:"isActive"`
}

// ListRoutings returns all routings (domains) for an application
func ListRoutings(c *gin.Context) {
	h := &Handlers{Repo: defaultRoutingsRepo}
	h.ListRoutings(c)
}

// ListRoutingsHandler returns all routings (domains) for an application (method on Handlers)
func (h *Handlers) ListRoutings(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// Get all instances for this application
	instances, err := h.Repo.GetApplicationInstances(appID)
	if err != nil {
		response.InternalServerError(c, "Failed to get application instances")
		return
	}

	var responses []RoutingResponse
	for _, instance := range instances {
		// Get domains for each instance
		domains, err := h.Repo.GetDomainsForInstance(instance.ID)
		if err != nil {
			continue // Skip instances with errors
		}

		for _, domain := range domains {
			createdAt := ""
			if domain.CreatedAt.Time != nil {
				createdAt = domain.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
			}

			// Use instance's active port as host port
			hostPort := 8080 // Default port
			if instance.ActivePort.Valid && instance.ActivePort.Int64 > 0 {
				hostPort = int(instance.ActivePort.Int64)
			}

			responses = append(responses, RoutingResponse{
				UID:        utils.EncodeFriendlyID(utils.PrefixRouting, domain.ID),
				DomainName: domain.Hostname,
				HostPort:   hostPort,
				IsActive:   domain.IsPrimary, // Using IsPrimary as IsActive for now
				CreatedAt:  createdAt,
			})
		}
	}

	if responses == nil {
		responses = []RoutingResponse{}
	}

	response.Data(c, responses)
}

// CreateRouting creates a new routing (domain) for an application
func CreateRouting(c *gin.Context) {
	h := &Handlers{Repo: defaultRoutingsRepo}
	h.CreateRouting(c)
}

// CreateRoutingHandler creates a new routing (domain) for an application (method on Handlers)
func (h *Handlers) CreateRouting(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	var req RoutingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Get the first instance for this application
	instances, err := h.Repo.GetApplicationInstances(appID)
	if err != nil || len(instances) == 0 {
		response.NotFound(c, "No application instance found. Please link the app to a host first.")
		return
	}

	// Use the first instance
	instance := instances[0]

	// Create the domain
	domain := &models.Domain{
		ApplicationInstanceID: instance.ID,
		Hostname:              req.DomainName,
		IsPrimary:             req.IsActive,
	}

	if err := h.Repo.AddDomain(domain); err != nil {
		response.InternalServerError(c, "Failed to create routing: "+err.Error())
		return
	}

	createdAt := ""
	if domain.CreatedAt.Time != nil {
		createdAt = domain.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	response.Created(c, RoutingResponse{
		UID:        utils.EncodeFriendlyID(utils.PrefixRouting, domain.ID),
		DomainName: domain.Hostname,
		HostPort:   req.HostPort,
		IsActive:   domain.IsPrimary,
		CreatedAt:  createdAt,
	})
}

// UpdateRouting updates an existing routing
func UpdateRouting(c *gin.Context) {
	h := &Handlers{Repo: defaultRoutingsRepo}
	h.UpdateRouting(c)
}

// UpdateRoutingHandler updates an existing routing (method on Handlers)
func (h *Handlers) UpdateRouting(c *gin.Context) {
	routingID := c.Param("routingId")
	domainID, err := utils.DecodeFriendlyID(utils.PrefixRouting, routingID)
	if err != nil {
		response.BadRequest(c, "Invalid routing ID")
		return
	}

	var req RoutingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Get the domain to find its instance ID
	domain, err := h.Repo.GetDomainByID(domainID)
	if err != nil {
		response.NotFound(c, "Routing not found")
		return
	}

	// Update domain hostname and primary status
	if err := h.Repo.UpdateDomain(domainID, req.DomainName, req.IsActive); err != nil {
		response.InternalServerError(c, "Failed to update routing: "+err.Error())
		return
	}

	createdAt := ""
	if domain.CreatedAt.Time != nil {
		createdAt = domain.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	response.Data(c, RoutingResponse{
		UID:        utils.EncodeFriendlyID(utils.PrefixRouting, domainID),
		DomainName: req.DomainName,
		HostPort:   req.HostPort,
		IsActive:   req.IsActive,
		CreatedAt:  createdAt,
	})
}

// DeleteRouting deletes a routing
func DeleteRouting(c *gin.Context) {
	h := &Handlers{Repo: defaultRoutingsRepo}
	h.DeleteRouting(c)
}

// DeleteRoutingHandler deletes a routing (method on Handlers)
func (h *Handlers) DeleteRouting(c *gin.Context) {
	routingID := c.Param("routingId")
	domainID, err := utils.DecodeFriendlyID(utils.PrefixRouting, routingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid routing ID"})
		return
	}

	if err := h.Repo.DeleteDomainByID(domainID); err != nil {
		response.InternalServerError(c, "Failed to delete routing: "+err.Error())
		return
	}

	response.Message(c, "Routing deleted successfully")
}

// GetLatestRelease returns the latest release/deployment for an application
func GetLatestRelease(c *gin.Context) {
	h := &Handlers{Repo: defaultRoutingsRepo}
	h.GetLatestRelease(c)
}

// GetLatestReleaseHandler returns the latest release/deployment for an application (method on Handlers)
func (h *Handlers) GetLatestRelease(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// Get the app by ID
	app, err := h.Repo.GetApplicationByID(appID)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	// Get the latest deployment history
	history, err := h.Repo.GetLatestDeploymentHistoryForApp(app.Name)
	if err != nil || history == nil {
		// No releases yet - return empty response
		response.Data(c, nil)
		return
	}

	// Get the instance for port info
	instances, err := h.Repo.GetApplicationInstances(appID)
	systemPort := 0
	if err == nil && len(instances) > 0 {
		if instances[0].ActivePort.Valid {
			systemPort = int(instances[0].ActivePort.Int64)
		}
	}

	createdAt := ""
	if history.CreatedAt.Time != nil {
		createdAt = history.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	response.Data(c, gin.H{
		"uid":        utils.EncodeFriendlyID(utils.PrefixRelease, history.ID),
		"version":    history.Version,
		"status":     history.Status,
		"systemPort": systemPort,
		"createdAt":  createdAt,
	})
}
