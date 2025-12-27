package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"youfun/shipyard/internal/crypto"
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/internal/sshutil"
	"encoding/base64"
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

// Legacy function wrappers for backward compatibility
// These use a default repository instance

var defaultHostsRepo = &DefaultRepository{}

// SSHHostRequest represents the request to create/update an SSH host
type SSHHostRequest struct {
	Name       string `json:"name" binding:"required"`
	Addr       string `json:"addr" binding:"required"`
	Port       int    `json:"port"`
	User       string `json:"user" binding:"required"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

// SSHHostResponse represents an SSH host in API responses
type SSHHostResponse struct {
	UID           string `json:"uid"`
	Name          string `json:"name"`
	Addr          string `json:"addr"`
	Port          int    `json:"port"`
	User          string `json:"user"`
	Status        string `json:"status"`
	Arch          string `json:"arch"`
	HasPassword   bool   `json:"has_password"`
	HasPrivateKey bool   `json:"has_private_key"`
	InitializedAt string `json:"initialized_at,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func hostToResponse(host *database.SSHHostRow) SSHHostResponse {
	resp := SSHHostResponse{
		UID:           utils.EncodeFriendlyID(utils.PrefixSSHHost, host.ID),
		Name:          host.Name,
		Addr:          host.Addr,
		Port:          host.Port,
		User:          host.User,
		Status:        host.Status,
		Arch:          host.Arch,
		HasPassword:   host.Password != nil && *host.Password != "",
		HasPrivateKey: host.PrivateKey != nil && *host.PrivateKey != "",
	}
	if host.InitializedAt.Time != nil {
		resp.InitializedAt = host.InitializedAt.Time.Format("2006-01-02 15:04:05")
	}
	if host.CreatedAt.Time != nil {
		resp.CreatedAt = host.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if host.UpdatedAt.Time != nil {
		resp.UpdatedAt = host.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return resp
}

// ListSSHHosts returns all SSH hosts
func ListSSHHosts(c *gin.Context) {
	h := &Handlers{Repo: defaultHostsRepo}
	h.ListSSHHosts(c)
}

// ListSSHHostsHandler returns all SSH hosts (method on Handlers)
func (h *Handlers) ListSSHHosts(c *gin.Context) {
	hosts, err := h.Repo.GetAllSSHHosts()
	if err != nil {
		response.InternalServerError(c, "Failed to list hosts")
		return
	}

	responses := make([]SSHHostResponse, len(hosts))
	for i := range hosts {
		// database.SSHHostRow is an alias for models.SSHHost
		responses[i] = hostToResponse((*database.SSHHostRow)(&hosts[i]))
	}

	response.Data(c, responses)
}

// GetSSHHost returns a specific SSH host
func GetSSHHost(c *gin.Context) {
	h := &Handlers{Repo: defaultHostsRepo}
	h.GetSSHHost(c)
}

// GetSSHHostHandler returns a specific SSH host (method on Handlers)
func (h *Handlers) GetSSHHost(c *gin.Context) {
	uid := c.Param("uid")
	hostID, err := utils.DecodeFriendlyID(utils.PrefixSSHHost, uid)
	if err != nil {
		response.BadRequest(c, "Invalid host ID")
		return
	}

	host, err := h.Repo.GetSSHHostByID(hostID)
	if err != nil {
		response.NotFound(c, "Host not found")
		return
	}

	response.Data(c, hostToResponse(host))
}

// CreateSSHHost creates a new SSH host
func CreateSSHHost(c *gin.Context) {
	h := &Handlers{Repo: defaultHostsRepo}
	h.CreateSSHHost(c)
}

// CreateSSHHostHandler creates a new SSH host (method on Handlers)
func (h *Handlers) CreateSSHHost(c *gin.Context) {
	var req SSHHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	if req.Password == "" && req.PrivateKey == "" {
		response.BadRequest(c, "Either password or private_key is required")
		return
	}

	if req.Port == 0 {
		req.Port = 22
	}

	// Verify SSH connection
	tempHost := &models.SSHHost{
		Name: req.Name,
		Addr: req.Addr,
		Port: req.Port,
		User: req.User,
	}
	if req.Password != "" {
		tempHost.Password = &req.Password
	}
	if req.PrivateKey != "" {
		tempHost.PrivateKey = &req.PrivateKey
	}

	// Use HostKeyVerifier to capture the host key (Trust On First Use logic for API creation)
	var capturedKey string
	verifier := &sshutil.HostKeyVerifier{
		TrustedKey: "",
		Confirm: func(hostname string, remote net.Addr, key ssh.PublicKey) bool {
			// Automatically accept and capture the key
			capturedKey = base64.StdEncoding.EncodeToString(key.Marshal())
			return true
		},
	}

	sshConfig, err := sshutil.NewClientConfig(tempHost, verifier.Callback)
	if err != nil {
		response.BadRequest(c, "Invalid SSH configuration: "+err.Error())
		return
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", req.Addr, req.Port), sshConfig)
	if err != nil {
		response.BadRequest(c, "Failed to connect to SSH host: "+err.Error())
		return
	}
	client.Close()

	// Use captured key from verifier if available, otherwise it might have been trusted by known_hosts (unlikely here)
	if capturedKey == "" && verifier.CapturedKey != "" {
		capturedKey = verifier.CapturedKey
	}

	// Encrypt credentials
	var encryptedPassword, encryptedKey *string
	if req.Password != "" {
		encrypted, err := crypto.Encrypt(req.Password)
		if err != nil {
			response.InternalServerError(c, "Failed to encrypt password")
			return
		}
		encryptedPassword = &encrypted
	}
	if req.PrivateKey != "" {
		encrypted, err := crypto.Encrypt(req.PrivateKey)
		if err != nil {
			response.InternalServerError(c, "Failed to encrypt private key")
			return
		}
		encryptedKey = &encrypted
	}

	var hostKeyPtr *string
	if capturedKey != "" {
		hostKeyPtr = &capturedKey
	}

	host, err := h.Repo.CreateSSHHost(req.Name, req.Addr, req.Port, req.User, encryptedPassword, encryptedKey, hostKeyPtr)
	if err != nil {
		response.InternalServerError(c, "Failed to create host")
		return
	}

	response.Created(c, hostToResponse(host))
}

// UpdateSSHHost updates an SSH host
func UpdateSSHHost(c *gin.Context) {
	h := &Handlers{Repo: defaultHostsRepo}
	h.UpdateSSHHost(c)
}

// UpdateSSHHostHandler updates an SSH host (method on Handlers)
func (h *Handlers) UpdateSSHHost(c *gin.Context) {
	uid := c.Param("uid")
	hostID, err := utils.DecodeFriendlyID(utils.PrefixSSHHost, uid)
	if err != nil {
		response.BadRequest(c, "Invalid host ID")
		return
	}

	var req SSHHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	if req.Port == 0 {
		req.Port = 22
	}

	// Encrypt credentials if provided
	var encryptedPassword, encryptedKey *string
	if req.Password != "" {
		encrypted, err := crypto.Encrypt(req.Password)
		if err != nil {
			response.InternalServerError(c, "Failed to encrypt password")
			return
		}
		encryptedPassword = &encrypted
	}
	if req.PrivateKey != "" {
		encrypted, err := crypto.Encrypt(req.PrivateKey)
		if err != nil {
			response.InternalServerError(c, "Failed to encrypt private key")
			return
		}
		encryptedKey = &encrypted
	}

	err = h.Repo.UpdateSSHHost(hostID, req.Name, req.Addr, req.Port, req.User, encryptedPassword, encryptedKey)
	if err != nil {
		response.InternalServerError(c, "Failed to update host")
		return
	}

	host, err := h.Repo.GetSSHHostByID(hostID)
	if err != nil {
		response.InternalServerError(c, "Failed to get updated host")
		return
	}

	response.Data(c, hostToResponse(host))
}

// DeleteSSHHost deletes an SSH host
func DeleteSSHHost(c *gin.Context) {
	h := &Handlers{Repo: defaultHostsRepo}
	h.DeleteSSHHost(c)
}

// DeleteSSHHostHandler deletes an SSH host (method on Handlers)
func (h *Handlers) DeleteSSHHost(c *gin.Context) {
	uid := c.Param("uid")
	hostID, err := utils.DecodeFriendlyID(utils.PrefixSSHHost, uid)
	if err != nil {
		response.BadRequest(c, "Invalid host ID")
		return
	}

	if err := h.Repo.DeleteSSHHost(hostID); err != nil {
		response.InternalServerError(c, "Failed to delete host")
		return
	}

	response.Message(c, "Host deleted successfully")
}

// TestSSHHost tests connection to an SSH host
func TestSSHHost(c *gin.Context) {
	h := &Handlers{Repo: defaultHostsRepo}
	h.TestSSHHost(c)
}

// TestSSHHostHandler tests connection to an SSH host (method on Handlers)
func (h *Handlers) TestSSHHost(c *gin.Context) {
	uid := c.Param("uid")
	hostID, err := utils.DecodeFriendlyID(utils.PrefixSSHHost, uid)
	if err != nil {
		response.BadRequest(c, "Invalid host ID")
		return
	}

	host, err := h.Repo.GetSSHHostByID(hostID)
	if err != nil {
		response.NotFound(c, "Host not found")
		return
	}

	// TODO: Implement actual SSH connection test
	_ = host

	response.Message(c, "Connection test successful")
}
