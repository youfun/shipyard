package api

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"youfun/shipyard/internal/api/handlers"
	"youfun/shipyard/internal/api/middleware"
	"syscall"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
)

//go:embed all:webui/*
var webUIFS embed.FS

// Server represents the HTTP server
type Server struct {
	Router *gin.Engine
	Port   string
}

// NewServer creates a new HTTP server with all routes configured
func NewServer(port string) *Server {
	// Initialize JWT Secret after database is ready
	middleware.InitJWTSecret()

	// Set global system port for handlers
	portInt, _ := strconv.Atoi(port)
	handlers.SetSystemPort(portInt)

	// Initialize global engine using the helper from engine.go
	debug := os.Getenv("GIN_MODE") != "release"
	router := Engine(debug)

	s := &Server{
		Router: router,
		Port:   strconv.Itoa(handlers.GlobalSystemPort),
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API routes group
	api := s.Router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/login/2fa", handlers.Login2FA)
			auth.POST("/device/code", handlers.DeviceCode)
			auth.GET("/device/token", handlers.DeviceToken)
		}

		// Setup route (public, only works if no users exist)
		api.POST("/setup", handlers.Setup)
		api.GET("/setup/status", handlers.SetupStatus)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User/Auth management
			protected.GET("/auth/me", handlers.GetCurrentUser)
			protected.POST("/auth/logout", handlers.Logout)
			protected.POST("/auth/change-password", handlers.ChangePassword)

			// 2FA management
			protected.POST("/auth/2fa/setup", handlers.Setup2FA)
			protected.POST("/auth/2fa/enable", handlers.Enable2FA)
			protected.POST("/auth/2fa/disable", handlers.Disable2FA)

			// Device/Session management
			protected.GET("/cli/sessions/:sessionId", handlers.GetDeviceSession)
			protected.POST("/cli/confirm", handlers.DeviceConfirm)
			protected.GET("/auth/devices", handlers.ListDevices)
			protected.DELETE("/auth/devices/:id", handlers.RevokeDevice)

			// SSH Hosts
			protected.GET("/ssh-hosts", handlers.ListSSHHosts)
			protected.POST("/ssh-hosts", handlers.CreateSSHHost)
			protected.GET("/ssh-hosts/:uid", handlers.GetSSHHost)
			protected.PUT("/ssh-hosts/:uid", handlers.UpdateSSHHost)
			protected.DELETE("/ssh-hosts/:uid", handlers.DeleteSSHHost)
			protected.POST("/ssh-hosts/:uid/test", handlers.TestSSHHost)

			// Applications
			protected.GET("/applications", handlers.ListApplications)
			protected.GET("/applications/:uid", handlers.GetApplication)
			protected.POST("/applications/:uid/start", handlers.StartApplication)
			protected.POST("/applications/:uid/stop", handlers.StopApplication)
			protected.POST("/applications/:uid/restart", handlers.RestartApplication)
			protected.GET("/applications/:uid/logs", handlers.GetApplicationLogs)
			protected.GET("/applications/:uid/releases/latest", handlers.GetLatestRelease)

			// Application Tokens
			protected.GET("/applications/:uid/tokens", handlers.ListApplicationTokens)
			protected.POST("/applications/:uid/tokens", handlers.CreateApplicationToken)
			protected.DELETE("/applications/:uid/tokens/:tokenId", handlers.DeleteApplicationToken)

			// Application Instances Management
			protected.POST("/instances/:uid/stop", handlers.StopInstance)
			protected.POST("/instances/:uid/start", handlers.StartInstance)
			protected.POST("/instances/:uid/restart", handlers.RestartInstance)
			protected.GET("/instances/:uid/logs", handlers.GetInstanceLogs)
			protected.GET("/instances/:uid/logs/stream", handlers.StreamInstanceLogs)

			// Routings (Domain management)
			protected.GET("/apps/:uid/routings", handlers.ListRoutings)
			protected.POST("/apps/:uid/routings", handlers.CreateRouting)
			protected.PUT("/routings/:routingId", handlers.UpdateRouting)
			protected.DELETE("/routings/:routingId", handlers.DeleteRouting)

			// Environment Variables
			protected.GET("/applications/:uid/environment-variables", handlers.ListEnvironmentVariables)
			protected.POST("/applications/:uid/environment-variables", handlers.CreateEnvironmentVariable)
			protected.PUT("/environment-variables/:envVarId", handlers.UpdateEnvironmentVariable)
			protected.DELETE("/environment-variables/:envVarId", handlers.DeleteEnvironmentVariable)

			// Deployments
			protected.GET("/applications/:uid/deployments", handlers.ListDeployments)
			protected.GET("/deployments/:uid", handlers.GetDeployment)
			protected.GET("/deployments/:uid/logs", handlers.GetDeploymentLogs)
			protected.POST("/deployments", handlers.CreateDeployment)
			protected.POST("/deployments/:uid/logs", handlers.UploadDeploymentLogs)
			protected.PATCH("/deployments/:uid/status", handlers.UpdateDeploymentStatus)

			// CLI-specific routes
			cli := protected.Group("/cli/v1")
			{
				cli.POST("/apps", handlers.CLICreateApplication)
				cli.POST("/applications", handlers.CLICreateApplication) // Alias
				cli.GET("/hosts", handlers.CLIListHosts)
				cli.GET("/applications/by-name/:name", handlers.CLIGetApplicationByName)
				cli.GET("/ssh-hosts/by-name/:name", handlers.CLIGetSSHHostByName)
				cli.POST("/apps/link", handlers.CLILinkAppToHost)
				cli.POST("/link", handlers.CLILinkAppToHost) // Alias
				cli.GET("/instance", handlers.CLIGetInstance)
				cli.GET("/deployments/latest", handlers.CLIGetLastDeployment) // Add this route

				// Deployments
				cli.POST("/deployments", handlers.CreateDeployment)
				cli.PUT("/deployments/:uid/status", handlers.UpdateDeploymentStatus)   // NOTE: Client uses PUT /status
				cli.PATCH("/deployments/:uid/status", handlers.UpdateDeploymentStatus) // Alias just in case
				cli.POST("/deployments/:uid/logs", handlers.UploadDeploymentLogs)
				
				// Server-side deployment (localhost deployment)
				cli.POST("/deployments/:uid/upload", handlers.UploadDeploymentArtifact)
				cli.POST("/deployments/:uid/execute", handlers.ExecuteServerDeployment)

				// Config
				cli.GET("/deploy/config", handlers.CLIGetDeployConfig)

				// Artifacts
				cli.GET("/artifacts/check", handlers.CLICheckArtifact)
				cli.POST("/artifacts", handlers.CLIRegisterArtifact)

				// Secrets (Environment Variables) management
				cli.GET("/secrets", handlers.CLIListSecrets)
				cli.POST("/secrets", handlers.CLISetSecret)
				cli.DELETE("/secrets", handlers.CLIUnsetSecret)

				// Domains management
				cli.POST("/domains/sync", handlers.CLISyncDomains)

				// Build artifacts management
				cli.GET("/builds", handlers.CLIListBuildArtifacts)
			}

			// System settings (Domain configuration)
			protected.GET("/system/settings", handlers.GetSystemSettings)
			protected.POST("/system/settings", handlers.UpdateSystemSettings)

			// System status
			protected.GET("/status", handlers.SystemStatus)

			// Dashboard statistics
			protected.GET("/dashboard/stats", handlers.DashboardStats)
			protected.GET("/dashboard/recent-deployments", handlers.RecentDeployments)
		}
	}

	// Serve static files for Web UI
	s.serveStaticFiles()
}

// serveStaticFiles serves the embedded frontend files
func (s *Server) serveStaticFiles() {
	// Get the subdirectory containing the build artifacts
	subFS, err := fs.Sub(webUIFS, "webui")
	if err != nil {
		log.Printf("Warning: Could not load embedded WebUI: %v", err)
		return
	}

	httpFS := http.FS(subFS)

	// Helper to serve index.html directly avoiding http.FileServer's redirect logic
	serveIndex := func(c *gin.Context) {
		f, err := subFS.Open("index.html")
		if err != nil {
			log.Printf("[Static] Error opening index.html: %v", err)
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			log.Printf("[Static] Error stat index.html: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// http.ServeContent handles Range requests, caching, etc.
		// embedded files implement io.ReadSeeker
		if rs, ok := f.(io.ReadSeeker); ok {
			http.ServeContent(c.Writer, c.Request, "index.html", stat.ModTime(), rs)
		} else {
			log.Println("[Static] index.html does not implement io.ReadSeeker")
			c.Status(http.StatusInternalServerError)
		}
	}

	// Middleware to serve static files if they exist
	s.Router.Use(func(c *gin.Context) {
		// Skip API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		log.Printf("[Static] Request: %s", path)

		// Explicitly serve index.html for root
		if path == "/" {
			log.Println("[Static] Serving index.html for root")
			serveIndex(c)
			c.Abort()
			return
		}

		// Check if file exists in FS
		// fs.Open requires paths without leading slash
		trimmedPath := path
		if len(path) > 0 && path[0] == '/' {
			trimmedPath = path[1:]
		}

		f, err := subFS.Open(trimmedPath)
		if err == nil {
			defer f.Close()
			stat, err := f.Stat()
			// Serve only if it's a file, not a directory
			if err == nil && !stat.IsDir() {
				log.Printf("[Static] Serving found file: %s", trimmedPath)
				c.FileFromFS(trimmedPath, httpFS)
				c.Abort()
				return
			}
		} else {
			log.Printf("[Static] File not found in FS: %s (err: %v)", trimmedPath, err)
		}

		log.Println("[Static] Delegating to next handler")
		c.Next()
	})

	// Fallback to index.html for SPA routing
	s.Router.NoRoute(func(c *gin.Context) {
		// Skip API routes
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}
		log.Printf("[NoRoute] Fallback serving index.html for: %s", path)
		serveIndex(c)
	})
}

// Run starts the HTTP server with graceful shutdown
func (s *Server) Run() error {
	// Sync system domain config with Caddy on startup
	// This ensures that if the port has changed, Caddy is updated
	handlers.SyncSystemDomainConfig(handlers.GlobalSystemPort)

	srv := &http.Server{
		Addr:    "0.0.0.0:" + s.Port,
		Handler: s.Router,
	}

	go func() {
		log.Printf("Server starting on port %s", s.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
	return nil
}
