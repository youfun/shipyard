package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/depsinstall"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// GetSystemSettings returns all system settings (currently just the domain)
func (h *Handlers) GetSystemSettings(c *gin.Context) {
	domain, err := h.Repo.GetSystemSetting("system_domain")
	if err != nil {
		response.InternalServerError(c, "Failed to get system domain: "+err.Error())
		return
	}

	response.Data(c, gin.H{
		"domain": domain,
	})
}

// UpdateSystemSettingsRequest represents the request to update system settings
type UpdateSystemSettingsRequest struct {
	Domain string `json:"domain" binding:"required"`
}

// UpdateSystemSettings updates system settings and configures Caddy
func (h *Handlers) UpdateSystemSettings(c *gin.Context) {
	var req UpdateSystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	log.Printf("🌐 Updating system domain to: %s", req.Domain)

	// 1. Save to DB
	if err := h.Repo.SetSystemSetting("system_domain", req.Domain); err != nil {
		log.Printf("❌ Failed to save system domain to DB: %v", err)
		response.InternalServerError(c, "Failed to save system domain: "+err.Error())
		return
	}
	log.Printf("✅ System domain saved to database")

	// 2. Configure Caddy
	if err := h.ConfigureCaddyForSystemDomain(req.Domain); err != nil {
		log.Printf("⚠️ Caddy configuration failed for domain %s: %v", req.Domain, err)
		response.Data(c, gin.H{
			"message": "System domain updated, but Caddy configuration failed: " + err.Error(),
			"domain":  req.Domain,
		})
		return
	}

	log.Printf("✨ System domain configuration completed successfully")
	response.Data(c, gin.H{
		"message": "System settings updated successfully",
		"domain":  req.Domain,
	})
}

// ConfigureCaddyForSystemDomain updates the Caddy configuration for the system domain
func (h *Handlers) ConfigureCaddyForSystemDomain(domain string) error {
	port := h.SystemPort
	if port == 0 {
		port = 8080 // Default port if none specified
	}

	log.Printf("🛠️ Configuring Caddy: domain=%s, target_port=%d", domain, port)

	// Ensure Caddy is installed and running
	if err := depsinstall.EnsureCaddyRunning(true, true); err != nil {
		return fmt.Errorf("failed to ensure Caddy is running: %w", err)
	}

	// Initialize Caddy service (Local)
	caddySvc := caddy.NewLocalService()

	// Check Caddy availability before updating
	if err := caddySvc.CheckAvailability(); err != nil {
		log.Printf("❌ Caddy Admin API not accessible: %v", err)
		return err
	}

	log.Printf("🚀 Sending update request to Caddy Admin API...")
	err := caddySvc.UpdateSystemRoute(domain, port)
	if err != nil {
		log.Printf("❌ Caddy route update failed: %v", err)
		return err
	}

	log.Printf("✅ Caddy route updated for %s -> localhost:%d", domain, port)
	return nil
}

// SyncSystemDomainConfig ensures Caddy is configured with the correct port on startup
func SyncSystemDomainConfig(port int) {
	log.Printf("🔄 Syncing system domain configuration on startup (port: %d)...", port)
	h := &Handlers{Repo: defaultAppsRepo, SystemPort: port}

	domain, err := h.Repo.GetSystemSetting("system_domain")
	if err != nil || domain == "" {
		log.Printf("ℹ️ No system domain configured, skipping sync")
		return
	}

	log.Printf("🔍 Found configured system domain: %s", domain)
	if err := h.ConfigureCaddyForSystemDomain(domain); err != nil {
		log.Printf("❌ Startup sync failed: %v", err)
	}
}

// Wrapper functions for routing

func GetSystemSettings(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.GetSystemSettings(c)
}

func UpdateSystemSettings(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo, SystemPort: GlobalSystemPort}
	h.UpdateSystemSettings(c)
}
