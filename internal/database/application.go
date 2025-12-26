package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// --- Application Table Operations ---
func AddApplication(app *models.Application) error {
	app.ID = uuid.New()
	now := time.Now()
	app.CreatedAt = models.NullableTime{Time: &now}
	app.UpdatedAt = models.NullableTime{Time: &now}
	query := `INSERT INTO applications (id, name, created_at, updated_at) VALUES (:id, :name, :created_at, :updated_at)`
	_, err := DB.NamedExec(query, app)
	return err
}

func GetApplicationByName(name string) (*models.Application, error) {
	var app models.Application
	query := Rebind("SELECT * FROM applications WHERE name = ?")
	err := DB.Get(&app, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}
	return &app, nil
}

// AppInstanceInfo contains simplified information about an application instance and its associated host
type AppInstanceInfo struct {
	HostName   string        `db:"name"`
	Status     string        `db:"status"`
	ActivePort sql.NullInt64 `db:"active_port"`
}

// --- Application Instance Table Operations ---

// GetInstancesForApp retrieves all instances for an application and their host information
func GetInstancesForApp(appID uuid.UUID) ([]AppInstanceInfo, error) {
	var instances []AppInstanceInfo
	query := Rebind(`
		SELECT
			h.name,
			i.status,
			i.active_port
		FROM application_instances i
		JOIN ssh_hosts h ON i.host_id = h.id
		WHERE i.application_id = ?
	`)
	err := DB.Select(&instances, query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to query application instances: %w", err)
	}
	return instances, nil
}

// GetApplicationInstance retrieves a single application instance by application ID and host ID.
func GetApplicationInstance(appID uuid.UUID, hostID uuid.UUID) (*models.ApplicationInstance, error) {
	var instance models.ApplicationInstance
	query := Rebind("SELECT * FROM application_instances WHERE application_id = ? AND host_id = ?")
	err := DB.Get(&instance, query, appID, hostID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceNotFound
		}
		return nil, err
	}
	return &instance, nil
}

// GetApplicationInstanceByID retrieves a single application instance by its ID.
func GetApplicationInstanceByID(instanceID uuid.UUID) (*models.ApplicationInstance, error) {
	var instance models.ApplicationInstance
	query := Rebind("SELECT * FROM application_instances WHERE id = ?")
	err := DB.Get(&instance, query, instanceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceNotFound
		}
		return nil, err
	}
	return &instance, nil
}

func LinkApplicationToHost(instance *models.ApplicationInstance) error {
	instance.ID = uuid.New()
	now := time.Now()
	instance.CreatedAt = models.NullableTime{Time: &now}
	instance.UpdatedAt = models.NullableTime{Time: &now}
	query := `INSERT INTO application_instances (id, application_id, host_id, status, created_at, updated_at) VALUES (:id, :application_id, :host_id, :status, :created_at, :updated_at)`
	_, err := DB.NamedExec(query, instance)
	return err
}

// #gedHostsForApp retrieves all hosts linked to a specific application.
func GetLinkedHostsForApp(appName string) ([]models.SSHHost, error) {
	app, err := GetApplicationByName(appName)
	if err != nil {
		return nil, fmt.Errorf("could not find application '%s': %w", appName, err)
	}

	var hosts []models.SSHHost
	query := Rebind(`
		SELECT h.*
		FROM ssh_hosts h
		INNER JOIN application_instances ai ON h.id = ai.host_id
		WHERE ai.application_id = ?
		ORDER BY h.name ASC
	`)
	err = DB.Select(&hosts, query, app.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked hosts for app '%s': %w", appName, err)
	}

	// Decrypt credentials for each host
	for i := range hosts {
		if err := hosts[i].DecryptCredentials(); err != nil {
			log.Printf("Warning: %v", err)
		}
	}

	return hosts, nil
}

func GetInstance(appName, hostName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
	app, err := GetApplicationByName(appName)
	if err != nil {
		return nil, nil, nil, err
	}
	host, err := GetSSHHostByName(hostName)
	if err != nil {
		return nil, nil, nil, err
	}

	var instance models.ApplicationInstance
	query := Rebind("SELECT * FROM application_instances WHERE application_id = ? AND host_id = ?")
	err = DB.Get(&instance, query, app.ID, host.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, app, host, ErrInstanceNotFound
		}
		return nil, nil, nil, err
	}

	return &instance, app, host, nil
}

// EnsureLocalhostInstance ensures localhost host and app instance exist for server-side deployment
func EnsureLocalhostInstance(appName string) (*models.ApplicationInstance, *models.Application, *models.SSHHost, error) {
	// 1. Get or create application
	app, err := GetApplicationByName(appName)
	if err != nil {
		if errors.Is(err, ErrAppNotFound) {
			// Create new application
			app = &models.Application{Name: appName}
			if err := AddApplication(app); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create application: %w", err)
			}
			log.Printf("✅ Created application: %s", appName)
		} else {
			return nil, nil, nil, err
		}
	}

	// 2. Get or create localhost host
	host, err := GetSSHHostByName("localhost")
	if err != nil {
		if errors.Is(err, ErrHostNotFound) {
			// Create localhost host (virtual host for server-side deployment)
			host = &models.SSHHost{
				Name: "localhost",
				Addr: "127.0.0.1",
				Port: 22,
				User: "localhost", // Placeholder, not used for server-side deployment
			}
			if err := AddSSHHost(host); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create localhost host: %w", err)
			}
			log.Printf("✅ Created localhost host for server-side deployment")
		} else {
			return nil, nil, nil, err
		}
	}

	// 3. Get or create instance
	var instance models.ApplicationInstance
	query := Rebind("SELECT * FROM application_instances WHERE application_id = ? AND host_id = ?")
	err = DB.Get(&instance, query, app.ID, host.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create new instance
			now := time.Now()
			instance = models.ApplicationInstance{
				ID:            uuid.New(),
				ApplicationID: app.ID,
				HostID:        host.ID,
				Status:        "pending",
				CreatedAt:     models.NullableTime{Time: &now},
				UpdatedAt:     models.NullableTime{Time: &now},
			}
			insertQuery := Rebind(`INSERT INTO application_instances 
				(id, application_id, host_id, status, created_at, updated_at) 
				VALUES (?, ?, ?, ?, ?, ?)`)
			_, err = DB.Exec(insertQuery, instance.ID, instance.ApplicationID, instance.HostID,
				instance.Status, instance.CreatedAt, instance.UpdatedAt)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create instance: %w", err)
			}
			log.Printf("✅ Created application instance for %s on localhost", appName)
		} else {
			return nil, nil, nil, err
		}
	}

	return &instance, app, host, nil
}

// UpdateInstancePortsForRollback atomically updates the active and previous ports for an instance.
func UpdateInstancePortsForRollback(instanceID uuid.UUID, newActivePort int, oldActivePort int) error {
	// Ensure oldActivePort is not the same as newActivePort to prevent locking into a state with no previous port
	if newActivePort == oldActivePort {
		oldActivePort = 0 // Or handle as an error, depending on desired logic
	}
	query := Rebind(`UPDATE application_instances SET active_port = ?, previous_active_port = ?, updated_at = ? WHERE id = ?`)
	_, err := DB.Exec(query, newActivePort, oldActivePort, time.Now(), instanceID)
	return err
}

func UpdateApplicationInstanceStatus(id uuid.UUID, status string) error {
	query := Rebind(`UPDATE application_instances SET status = ?, updated_at = ? WHERE id = ?`)
	_, err := DB.Exec(query, status, time.Now(), id)
	return err
}
