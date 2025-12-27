package database

import (
	"youfun/shipyard/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AddDomain adds a new domain to an application instance.
func AddDomain(domain *models.Domain) error {
	domain.ID = uuid.New()
	now := time.Now()
	domain.CreatedAt = models.NullableTime{Time: &now}

	query := `INSERT INTO domains (id, application_instance_id, hostname, is_primary, created_at) 
	          VALUES (:id, :application_instance_id, :hostname, :is_primary, :created_at)`
	_, err := DB.NamedExec(query, domain)
	return err
}

// GetDomainsForInstance retrieves all domains for an application instance.
func GetDomainsForInstance(instanceID uuid.UUID) ([]models.Domain, error) {
	var domains []models.Domain
	query := Rebind("SELECT * FROM domains WHERE application_instance_id = ? ORDER BY is_primary DESC, hostname ASC")
	err := DB.Select(&domains, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query domains: %w", err)
	}
	return domains, nil
}

// DeleteDomain deletes a domain.
func DeleteDomain(instanceID uuid.UUID, hostname string) error {
	query := Rebind("DELETE FROM domains WHERE application_instance_id = ? AND hostname = ?")
	_, err := DB.Exec(query, instanceID, hostname)
	return err
}

// UpdateDomainPrimaryStatus updates the primary domain status.
func UpdateDomainPrimaryStatus(instanceID uuid.UUID, hostname string, isPrimary bool) error {
	query := Rebind("UPDATE domains SET is_primary = ? WHERE application_instance_id = ? AND hostname = ?")
	_, err := DB.Exec(query, isPrimary, instanceID, hostname)
	return err
}

// SetPrimaryDomain sets a domain as primary and marks other domains as non-primary.
func SetPrimaryDomain(instanceID uuid.UUID, hostname string) error {
	tx, err := DB.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// First, set all domains as non-primary.
	clearQuery := Rebind("UPDATE domains SET is_primary = FALSE WHERE application_instance_id = ?")
	_, err = tx.Exec(clearQuery, instanceID)
	if err != nil {
		return fmt.Errorf("failed to clear primary domain status: %w", err)
	}

	// Set the specified domain as primary.
	setQuery := Rebind("UPDATE domains SET is_primary = TRUE WHERE application_instance_id = ? AND hostname = ?")
	_, err = tx.Exec(setQuery, instanceID, hostname)
	if err != nil {
		return fmt.Errorf("failed to set primary domain: %w", err)
	}

	return tx.Commit()
}

// GetDomainByID retrieves a domain by its ID
// GetDomainByID retrieves a domain by its ID
func GetDomainByID(domainID uuid.UUID) (*models.Domain, error) {
	var domain models.Domain
	query := Rebind("SELECT * FROM domains WHERE id = ?")
	err := DB.Get(&domain, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to query domain: %w", err)
	}
	return &domain, nil
}

// UpdateDomain updates a domain's hostname and primary status
func UpdateDomain(domainID uuid.UUID, hostname string, isPrimary bool) error {
	query := Rebind("UPDATE domains SET hostname = ?, is_primary = ? WHERE id = ?")
	_, err := DB.Exec(query, hostname, isPrimary, domainID)
	return err
}

// DeleteDomainByID deletes a domain by its ID
func DeleteDomainByID(domainID uuid.UUID) error {
	query := Rebind("DELETE FROM domains WHERE id = ?")
	_, err := DB.Exec(query, domainID)
	return err
}
