package caddy

import (
	"fmt"
	"log"

	"github.com/OrbitDeploy/fastcaddy"
	"golang.org/x/crypto/ssh"
)

// Service provides methods for interacting with the Caddy API.
// It uses the fastcaddy library to communicate over an existing SSH client.
type Service struct {
	client      *fastcaddy.FastCaddy
	isLocalhost bool // whether this is a local Caddy service
}

// NewService creates a new Caddy service wrapper for remote deployment.
// It requires a pre-configured *ssh.Client to establish the underlying tunnel for fastcaddy.
func NewService(sshClient *ssh.Client) *Service {
	// Initialize fastcaddy with the provided ssh.Client
	fc := fastcaddy.New(fastcaddy.WithSSHClient(sshClient))
	return &Service{
		client:      fc,
		isLocalhost: false,
	}
}

// NewLocalService creates a new Caddy service wrapper for localhost deployment.
// It connects directly to the local Caddy instance without SSH.
func NewLocalService() *Service {
	// Initialize fastcaddy without SSH (defaults to localhost:2019)
	fc := fastcaddy.New()
	return &Service{
		client:      fc,
		isLocalhost: true,
	}
}

// GetConfig retrieves the Caddy configuration from a specified path.
func (s *Service) GetConfig(path string) (interface{}, error) {
	log.Printf("Fetching Caddy config from path: %s", path)
	config, err := s.client.GetConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Caddy config from path '%s': %w", path, err)
	}
	return config, nil
}

// SetupCaddy initializes Caddy configuration with basic structure (apps/http/servers)
// Checks if already initialized to avoid overwriting existing configuration
func (s *Service) SetupCaddy() error {
	// Check if already initialized by checking if srv0 exists
	config, err := s.client.GetConfig("/apps/http/servers/srv0")
	if err == nil && config != nil {
		log.Println("ℹ️  Caddy configuration already initialized, skipping setup")
		return nil
	}

	log.Println("🔧 Setting up Caddy basic configuration...")

	// Setup with default parameters:
	// - cfToken: empty (not using Cloudflare yet)
	// - serverName: "srv0" (default server name)
	// - local: false (always use ACME-ready config, even for localhost)
	// - installTrust: nil (use default behavior)
	err = s.client.SetupCaddy("", "srv0", false, nil)
	if err != nil {
		return fmt.Errorf("failed to setup Caddy configuration: %w", err)
	}

	log.Println("✅ Caddy configuration initialized successfully")
	return nil
}

// CheckAvailability verifies that the Caddy Admin API is reachable.
func (s *Service) CheckAvailability() error {
	if s.isLocalhost {
		log.Println("Checking local Caddy API availability...")
	} else {
		log.Println("Checking Caddy API availability via SSH (GetConfig /)...")
	}

	// GetConfig("/") is a good way to check for connectivity and basic API functionality.
	_, err := s.client.GetConfig("/")
	if err != nil {
		if s.isLocalhost {
			return fmt.Errorf("failed to connect to local Caddy Admin API (ensure Caddy is running): %w", err)
		}
		return fmt.Errorf("failed to reach Caddy Admin API via SSH: %w", err)
	}

	if s.isLocalhost {
		log.Println("✅ Successfully connected to local Caddy service")
	} else {
		log.Println("Successfully connected to the Caddy admin service.")
	}
	return nil
}

// UpdateReverseProxy uses fastcaddy to configure a reverse proxy route over SSH.
func (s *Service) UpdateReverseProxy(domain string, targetPort int) error {
	proxyTo := fmt.Sprintf("localhost:%d", targetPort)
	log.Printf("Attempting to update Caddy reverse proxy via SSH: %s -> %s", domain, proxyTo)
	// This single call handles the complex task of updating the Caddy config over the SSH tunnel.
	return s.client.AddReverseProxy(domain, proxyTo)
}

// UpdateReverseProxyMultiDomain configures reverse proxies for multiple domains to the same target port
func (s *Service) UpdateReverseProxyMultiDomain(domains []string, targetPort int) error {
	if len(domains) == 0 {
		return fmt.Errorf("domain list cannot be empty")
	}

	proxyTo := fmt.Sprintf("localhost:%d", targetPort)
	log.Printf("Configuring reverse proxy for %d domains, target port: %d", len(domains), targetPort)

	// Configure reverse proxy for each domain
	for _, domain := range domains {
		log.Printf("  Configuring domain: %s -> %s", domain, proxyTo)
		if err := s.client.AddReverseProxy(domain, proxyTo); err != nil {
			return fmt.Errorf("failed to configure reverse proxy for domain '%s': %w", domain, err)
		}
	}

	log.Printf("✅ Successfully configured reverse proxy for %d domains", len(domains))
	return nil
}

// DeleteRoute uses fastcaddy to delete a route over SSH.
func (s *Service) DeleteRoute(domain string) error {
	log.Printf("Attempting to delete Caddy route via SSH: %s", domain)
	return s.client.DeleteRoute(domain)
}

// UpdateSystemRoute configures the reverse proxy for the Deployer system itself (Web UI).
func (s *Service) UpdateSystemRoute(domain string, targetPort int) error {
	proxyTo := fmt.Sprintf("localhost:%d", targetPort)
	log.Printf("Updating System Route (Web UI): %s -> %s", domain, proxyTo)
	return s.client.AddReverseProxy(domain, proxyTo)
}
