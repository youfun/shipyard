package handlers

import (
	"runtime"
	"youfun/shipyard/internal/api/response"

	"github.com/gin-gonic/gin"
)

// Legacy function wrappers for backward compatibility
var defaultStatusRepo = &DefaultRepository{}

// Global version variable
var serverVersion = "dev"

// SetVersion sets the server version
func SetVersion(version string) {
	serverVersion = version
}

// SystemStatusResponse represents the system status
type SystemStatusResponse struct {
	Status       string `json:"status"`
	Version      string `json:"version"`
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

// SystemStatus returns system status information
func SystemStatus(c *gin.Context) {
	h := &Handlers{Repo: defaultStatusRepo}
	h.SystemStatus(c)
}

// SystemStatusHandler returns system status information (method on Handlers)
func (h *Handlers) SystemStatus(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response.Data(c, gin.H{
		"status":        "healthy",
		"version":       serverVersion,
		"go_version":    runtime.Version(),
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"memory": gin.H{
			"alloc_mb":       memStats.Alloc / 1024 / 1024,
			"total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
			"sys_mb":         memStats.Sys / 1024 / 1024,
		},
	})
}

// DashboardStatsResponse represents dashboard statistics
type DashboardStatsResponse struct {
	ApplicationsCount int `json:"applications_count"`
	HostsCount        int `json:"hosts_count"`
	DeploymentsCount  int `json:"deployments_count"`
}

// DashboardStats returns statistics for the dashboard
func DashboardStats(c *gin.Context) {
	h := &Handlers{Repo: defaultStatusRepo}
	h.DashboardStats(c)
}

// DashboardStatsHandler returns statistics for the dashboard (method on Handlers)
func (h *Handlers) DashboardStats(c *gin.Context) {
	// Get applications count
	apps, err := h.Repo.GetAllApplications()
	appsCount := 0
	if err == nil {
		appsCount = len(apps)
	}

	// Get hosts count
	hosts, err := h.Repo.GetAllSSHHosts()
	hostsCount := 0
	if err == nil {
		hostsCount = len(hosts)
	}

	// Get deployments count
	deploymentsCount, err := h.Repo.GetDeploymentsCount()
	if err != nil {
		deploymentsCount = 0
	}

	response.Data(c, DashboardStatsResponse{
		ApplicationsCount: appsCount,
		HostsCount:        hostsCount,
		DeploymentsCount:  deploymentsCount,
	})
}

// RecentDeploymentResponse represents a recent deployment in API responses
type RecentDeploymentResponse struct {
	UID          string  `json:"uid"`
	AppName      string  `json:"app_name"`
	HostName     string  `json:"host_name"`
	HostAddr     string  `json:"host_addr"`
	Version      string  `json:"version"`
	GitCommitSHA string  `json:"git_commit"`
	Status       string  `json:"status"`
	Port         int     `json:"port"`
	DeployedAt   *string `json:"deployed_at,omitempty"`
}

// RecentDeployments returns the most recent deployments for the dashboard
func RecentDeployments(c *gin.Context) {
	h := &Handlers{Repo: defaultStatusRepo}
	h.RecentDeployments(c)
}

// RecentDeploymentsHandler returns the most recent deployments (method on Handlers)
func (h *Handlers) RecentDeployments(c *gin.Context) {
	deployments, err := h.Repo.GetRecentDeploymentsGlobal(20)
	if err != nil {
		response.Data(c, []RecentDeploymentResponse{})
		return
	}

	var result []RecentDeploymentResponse
	for _, d := range deployments {
		var deployedAt *string
		if d.DeployedAt != nil {
			t := d.DeployedAt.Format("2006-01-02T15:04:05Z07:00")
			deployedAt = &t
		}
		result = append(result, RecentDeploymentResponse{
			UID:          d.ID.String(),
			AppName:      d.AppName,
			HostName:     d.HostName,
			HostAddr:     d.HostAddr,
			Version:      d.Version,
			GitCommitSHA: d.GitCommitSHA,
			Status:       d.Status,
			Port:         d.Port,
			DeployedAt:   deployedAt,
		})
	}

	response.Data(c, result)
}
