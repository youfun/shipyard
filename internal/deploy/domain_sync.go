package deploy

import (
	"bufio"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// SyncDomainsFromConfig syncs domains from the config file to the database
func SyncDomainsFromConfig(instanceID uuid.UUID, domains []string, primaryDomain string) error {
	log.Println("--- Syncing domain config to database ---")

	if len(domains) == 0 {
		log.Println("No domains defined in config file, skipping domain sync")
		return nil
	}

	// Get current domain list from database
	existingDomains, err := database.GetDomainsForInstance(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get existing domains: %w", err)
	}

	// Create a map for quick lookup
	existingMap := make(map[string]bool)
	for _, d := range existingDomains {
		existingMap[d.Hostname] = true
	}

	// Add domains present in config but missing in database
	for _, hostname := range domains {
		if !existingMap[hostname] {
			isPrimary := hostname == primaryDomain
			domain := &models.Domain{
				ApplicationInstanceID: instanceID,
				Hostname:              hostname,
				IsPrimary:             isPrimary,
			}
			if err := database.AddDomain(domain); err != nil {
				return fmt.Errorf("failed to add domain '%s': %w", hostname, err)
			}
			log.Printf("✅ Added new domain: %s (Primary: %v)", hostname, isPrimary)
		}
	}

	// Check for domains in database but not in config (need to warn user)
	configMap := make(map[string]bool)
	for _, hostname := range domains {
		configMap[hostname] = true
	}

	foundMissing := false
	for _, d := range existingDomains {
		if !configMap[d.Hostname] {
			foundMissing = true
			log.Printf("⚠️  Warning: Domain '%s' exists in database but not defined in config file.", d.Hostname)
			log.Printf("    To remove, manually run: shipyard domain remove --host <host> --domain %s", d.Hostname)
		}
	}

	if foundMissing {
		log.Println("You can press Q or Ctrl+C to exit now, otherwise deployment will continue in 5s.")

		// Wait 5 seconds, during which pressing Q (Enter) or receiving SIGINT cancels the operation
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		inputCh := make(chan string, 1)
		go func() {
			reader := bufio.NewReader(os.Stdin)
			s, _ := reader.ReadString('\n')
			inputCh <- strings.TrimSpace(s)
		}()

		select {
		case sig := <-sigCh:
			return fmt.Errorf("Received signal %v, operation cancelled", sig)
		case in := <-inputCh:
			if strings.EqualFold(in, "q") {
				return fmt.Errorf("User cancelled operation")
			}
			// Continue on other input
		case <-time.After(5 * time.Second):
			// Continue after timeout
		}
	}

	// Update primary domain status (if primary domain is specified)
	if primaryDomain != "" {
		if configMap[primaryDomain] {
			if err := database.SetPrimaryDomain(instanceID, primaryDomain); err != nil {
				return fmt.Errorf("failed to set primary domain '%s': %w", primaryDomain, err)
			}
			log.Printf("✅ Primary domain set to: %s", primaryDomain)
		} else {
			log.Printf("⚠️  Warning: Primary domain '%s' is not in the domain list", primaryDomain)
		}
	}

	return nil
}

// GetDomainsForDeploy gets the list of domains needed for deployment
func GetDomainsForDeploy(instanceID uuid.UUID) ([]string, error) {
	domains, err := database.GetDomainsForInstance(instanceID)
	if err != nil {
		return nil, err
	}

	// If there are no domains stored for the instance, return empty list.
	if len(domains) == 0 {
		return nil, nil
	}

	var hostnames []string
	for _, d := range domains {
		hostnames = append(hostnames, d.Hostname)
	}
	return hostnames, nil
}

// SyncDomainsForDeployment syncs domain configuration during deployment
func (d *Deployer) SyncDomainsForDeployment() error {
	// Read domains from config file
	domains := config.AppConfig.Domains
	primaryDomain := config.AppConfig.PrimaryDomain

	// If there are no domains in the config we won't attempt to fallback to a removed Application.Domain field

	// Sync to database
	if len(domains) > 0 {
		// If we have an API client, use it to sync domains instead of direct database access
		if d.APIClient != nil {
			return d.APIClient.SyncDomains(d.Instance.ID.String(), domains, primaryDomain)
		}
		return SyncDomainsFromConfig(d.Instance.ID, domains, primaryDomain)
	}

	return nil
}
