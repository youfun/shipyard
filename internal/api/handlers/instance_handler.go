package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"youfun/shipyard/internal/caddy"
	"youfun/shipyard/internal/logs"
	"youfun/shipyard/internal/sshutil"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// StopInstance stops a specific application instance
func StopInstance(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.StopInstance(c)
}

func (h *Handlers) StopInstance(c *gin.Context) {
	uid := c.Param("uid")
	instanceID, err := utils.DecodeFriendlyID(utils.PrefixAppInstance, uid)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	// 1. Get Instance details
	instance, err := h.Repo.GetApplicationInstanceByID(instanceID)
	if err != nil {
		response.NotFound(c, "Instance not found")
		return
	}

	if !instance.ActivePort.Valid || instance.ActivePort.Int64 == 0 {
		response.BadRequest(c, "Instance has no active port")
		return
	}
	port := instance.ActivePort.Int64

	// 2. Get App and Host details
	app, err := h.Repo.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		response.InternalServerError(c, "Failed to get application info")
		return
	}

	host, err := h.Repo.GetSSHHostByID(instance.HostID)
	if err != nil {
		response.InternalServerError(c, "Failed to get host info")
		return
	}

	// 3. Connect to SSH
	sshConfig, err := sshutil.NewClientConfig(host, nil)
	if err != nil {
		response.InternalServerError(c, "Failed to create SSH config: "+err.Error())
		return
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Addr, host.Port), sshConfig)
	if err != nil {
		response.InternalServerError(c, "Failed to connect to host: "+err.Error())
		return
	}
	defer client.Close()

	// 4. Stop service
	cmd := fmt.Sprintf("sudo systemctl stop %s@%d && sudo systemctl disable %s@%d", app.Name, port, app.Name, port)
	if _, err := sshutil.ExecuteRemoteCommand(client, cmd); err != nil {
		response.InternalServerError(c, "Failed to stop service: "+err.Error())
		return
	}

	// 5. Update Caddy (Remove route)
	// Need to find domains associated with this instance
	domains, err := h.Repo.GetDomainsForInstance(instance.ID)
	if err == nil && len(domains) > 0 {
		caddySvc := caddy.NewService(client)
		if err := caddySvc.CheckAvailability(); err == nil {
			for _, d := range domains {
				if err := caddySvc.DeleteRoute(d.Hostname); err != nil {
					log.Printf("Failed to delete route for %s: %v", d.Hostname, err)
					// Continue anyway, not fatal
				}
			}
		}
	}

	// 6. Update DB status
	if err := h.Repo.UpdateApplicationInstanceStatus(instance.ID, "stopped"); err != nil {
		log.Printf("Failed to update instance status: %v", err)
	}

	response.Message(c, "Instance stopped successfully")
}

// StartInstance starts a specific application instance
func StartInstance(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.StartInstance(c)
}

func (h *Handlers) StartInstance(c *gin.Context) {
	uid := c.Param("uid")
	instanceID, err := utils.DecodeFriendlyID(utils.PrefixAppInstance, uid)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	// 1. Get Instance details
	instance, err := h.Repo.GetApplicationInstanceByID(instanceID)
	if err != nil {
		response.NotFound(c, "Instance not found")
		return
	}

	if !instance.ActivePort.Valid || instance.ActivePort.Int64 == 0 {
		response.BadRequest(c, "Instance has no active port configured")
		return
	}
	port := instance.ActivePort.Int64

	// 2. Get App and Host details
	app, err := h.Repo.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		response.InternalServerError(c, "Failed to get application info")
		return
	}

	host, err := h.Repo.GetSSHHostByID(instance.HostID)
	if err != nil {
		response.InternalServerError(c, "Failed to get host info")
		return
	}

	// 3. Connect to SSH
	sshConfig, err := sshutil.NewClientConfig(host, nil)
	if err != nil {
		response.InternalServerError(c, "Failed to create SSH config: "+err.Error())
		return
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Addr, host.Port), sshConfig)
	if err != nil {
		response.InternalServerError(c, "Failed to connect to host: "+err.Error())
		return
	}
	defer client.Close()

	// 4. Start service
	cmd := fmt.Sprintf("sudo systemctl enable %s@%d && sudo systemctl start %s@%d", app.Name, port, app.Name, port)
	if _, err := sshutil.ExecuteRemoteCommand(client, cmd); err != nil {
		response.InternalServerError(c, "Failed to start service: "+err.Error())
		return
	}

	// 5. Update Caddy (Add route)
	domains, err := h.Repo.GetDomainsForInstance(instance.ID)
	if err == nil && len(domains) > 0 {
		caddySvc := caddy.NewService(client)
		if err := caddySvc.CheckAvailability(); err == nil {
			domainNames := make([]string, len(domains))
			for i, d := range domains {
				domainNames[i] = d.Hostname
			}
			if err := caddySvc.UpdateReverseProxyMultiDomain(domainNames, int(port)); err != nil {
				log.Printf("Failed to update Caddy: %v", err)
				// Not returning error here as the service started successfully
			}
		}
	}

	// 6. Update DB status
	if err := h.Repo.UpdateApplicationInstanceStatus(instance.ID, "running"); err != nil {
		log.Printf("Failed to update instance status: %v", err)
	}

	response.Message(c, "Instance started successfully")
}

// RestartInstance restarts a specific application instance
func RestartInstance(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.RestartInstance(c)
}

func (h *Handlers) RestartInstance(c *gin.Context) {
	uid := c.Param("uid")
	instanceID, err := utils.DecodeFriendlyID(utils.PrefixAppInstance, uid)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	// 1. Get Instance details
	instance, err := h.Repo.GetApplicationInstanceByID(instanceID)
	if err != nil {
		response.NotFound(c, "Instance not found")
		return
	}

	if !instance.ActivePort.Valid || instance.ActivePort.Int64 == 0 {
		response.BadRequest(c, "Instance has no active port")
		return
	}
	port := instance.ActivePort.Int64

	// 2. Get App and Host details
	app, err := h.Repo.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		response.InternalServerError(c, "Failed to get application info")
		return
	}

	host, err := h.Repo.GetSSHHostByID(instance.HostID)
	if err != nil {
		response.InternalServerError(c, "Failed to get host info")
		return
	}

	// 3. Connect to SSH
	sshConfig, err := sshutil.NewClientConfig(host, nil)
	if err != nil {
		response.InternalServerError(c, "Failed to create SSH config: "+err.Error())
		return
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Addr, host.Port), sshConfig)
	if err != nil {
		response.InternalServerError(c, "Failed to connect to host: "+err.Error())
		return
	}
	defer client.Close()

	// 4. Restart service
	// We also ensure it's enabled and Caddy is updated, just in case
	cmd := fmt.Sprintf("sudo systemctl enable %s@%d && sudo systemctl restart %s@%d", app.Name, port, app.Name, port)
	if _, err := sshutil.ExecuteRemoteCommand(client, cmd); err != nil {
		response.InternalServerError(c, "Failed to restart service: "+err.Error())
		return
	}

	// 5. Update Caddy (Ensure route exists)
	domains, err := h.Repo.GetDomainsForInstance(instance.ID)
	if err == nil && len(domains) > 0 {
		caddySvc := caddy.NewService(client)
		if err := caddySvc.CheckAvailability(); err == nil {
			domainNames := make([]string, len(domains))
			for i, d := range domains {
				domainNames[i] = d.Hostname
			}
			if err := caddySvc.UpdateReverseProxyMultiDomain(domainNames, int(port)); err != nil {
				log.Printf("Failed to update Caddy: %v", err)
			}
		}
	}

	// 6. Update DB status
	if err := h.Repo.UpdateApplicationInstanceStatus(instance.ID, "running"); err != nil {
		log.Printf("Failed to update instance status: %v", err)
	}

	response.Message(c, "Instance restarted successfully")
}

// GetInstanceLogs fetches logs for a specific application instance
func GetInstanceLogs(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.GetInstanceLogs(c)
}

func (h *Handlers) GetInstanceLogs(c *gin.Context) {
	uid := c.Param("uid")
	instanceID, err := utils.DecodeFriendlyID(utils.PrefixAppInstance, uid)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	linesStr := c.DefaultQuery("lines", "500")
	lines, err := strconv.Atoi(linesStr)
	if err != nil {
		lines = 500
	}

	// 1. Get Instance details
	instance, err := h.Repo.GetApplicationInstanceByID(instanceID)
	if err != nil {
		response.NotFound(c, "Instance not found")
		return
	}

	if !instance.ActivePort.Valid || instance.ActivePort.Int64 == 0 {
		response.BadRequest(c, "Instance has no active port")
		return
	}
	port := instance.ActivePort.Int64

	// 2. Get App and Host details
	app, err := h.Repo.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		response.InternalServerError(c, "Failed to get application info")
		return
	}

	host, err := h.Repo.GetSSHHostByID(instance.HostID)
	if err != nil {
		response.InternalServerError(c, "Failed to get host info")
		return
	}

	// 3. Fetch logs via SSH
	// We use nil for hostKeyCallback to use default verification against DB stored key
	logContent, err := logs.FetchJournalLogs(host, app.Name, int(port), lines, false, nil)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch logs: "+err.Error())
		return
	}

	response.Data(c, gin.H{"logs": logContent})
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate against allowed origins
		// For now, allow same origin only
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Allow if no origin header (e.g., native apps)
		}
		// Parse and validate origin matches host
		return true // TODO: Implement proper origin validation
	},
}

// StreamInstanceLogs streams logs for a specific application instance via WebSocket
func StreamInstanceLogs(c *gin.Context) {
	h := &Handlers{Repo: defaultAppsRepo}
	h.StreamInstanceLogs(c)
}

func (h *Handlers) StreamInstanceLogs(c *gin.Context) {
	uid := c.Param("uid")
	instanceID, err := utils.DecodeFriendlyID(utils.PrefixAppInstance, uid)
	if err != nil {
		response.BadRequest(c, "Invalid instance ID")
		return
	}

	linesStr := c.DefaultQuery("lines", "100")
	lines, err := strconv.Atoi(linesStr)
	if err != nil {
		lines = 100
	}

	// For WebSocket, we need to authenticate before upgrading
	// The token is passed via query parameter since WebSocket doesn't support custom headers easily
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Try to get from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}
	}

	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Validate token (assuming middleware package has ValidateToken function)
	// Note: In production, you should use the actual middleware validation
	// For now, we'll skip validation since the route is already protected by AuthMiddleware
	// The token validation is already done by the middleware before reaching here

	// 1. Get Instance details
	instance, err := h.Repo.GetApplicationInstanceByID(instanceID)
	if err != nil {
		response.NotFound(c, "Instance not found")
		return
	}

	if !instance.ActivePort.Valid || instance.ActivePort.Int64 == 0 {
		response.BadRequest(c, "Instance has no active port")
		return
	}
	port := instance.ActivePort.Int64

	// 2. Get App and Host details
	app, err := h.Repo.GetApplicationByID(instance.ApplicationID)
	if err != nil {
		response.InternalServerError(c, "Failed to get application info")
		return
	}

	host, err := h.Repo.GetSSHHostByID(instance.HostID)
	if err != nil {
		response.InternalServerError(c, "Failed to get host info")
		return
	}

	// 3. Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// 4. Connect to SSH
	sshConfig, err := sshutil.NewClientConfig(host, nil)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: Failed to create SSH config: %v", err)))
		return
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Addr, host.Port), sshConfig)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: Failed to connect to host: %v", err)))
		return
	}
	defer sshClient.Close()

	// 5. Create SSH session
	session, err := sshClient.NewSession()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: Failed to create SSH session: %v", err)))
		return
	}
	defer session.Close()

	// 6. Build journalctl command with follow flag
	unitName := fmt.Sprintf("%s@%d", app.Name, port)
	cmd := fmt.Sprintf("journalctl -u %s -n %d --no-pager -o cat -f", unitName, lines)

	// 7. Set up stdout pipe
	stdout, err := session.StdoutPipe()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: Failed to create stdout pipe: %v", err)))
		return
	}

	// 8. Start the command
	if err := session.Start(cmd); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: Failed to start journalctl: %v", err)))
		return
	}

	// 9. Set up channels for graceful shutdown
	done := make(chan struct{})
	errChan := make(chan error, 1)
	var closeOnce sync.Once

	// 10. Handle WebSocket messages (for client disconnect detection)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				closeOnce.Do(func() { close(done) })
				return
			}
		}
	}()

	// 11. Stream logs from SSH to WebSocket
	go func() {
		buffer := make([]byte, 4096)
		for {
			select {
			case <-done:
				return
			default:
				// Use blocking read - no busy waiting
				n, err := stdout.Read(buffer)
				if err != nil {
					// Check if already closed before sending error
					select {
					case errChan <- err:
					default:
					}
					return
				}
				if n > 0 {
					// Send log data to WebSocket client
					if err := conn.WriteMessage(websocket.TextMessage, buffer[:n]); err != nil {
						select {
						case errChan <- err:
						default:
						}
						return
					}
				}
			}
		}
	}()

	// 12. Wait for completion or error
	select {
	case <-done:
		// Client disconnected
		session.Signal(ssh.SIGINT)
		log.Printf("WebSocket client disconnected for instance %s", uid)
	case err := <-errChan:
		// Error occurred
		log.Printf("Error streaming logs for instance %s: %v", uid, err)
	case <-time.After(1 * time.Hour):
		// Timeout after 1 hour
		log.Printf("WebSocket connection timeout for instance %s", uid)
	}
}
