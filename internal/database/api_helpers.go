package database

import (
	"database/sql"
	"youfun/shipyard/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
)

// SSHHostRow is an alias for models.SSHHost for API use
type SSHHostRow = models.SSHHost

// ApplicationInstanceRow is an alias for models.ApplicationInstance
type ApplicationInstanceRow = models.ApplicationInstance

// GetSSHHostByID retrieves an SSH host by ID (and decrypts credentials)
func GetSSHHostByID(id uuid.UUID) (*SSHHostRow, error) {
	var host SSHHostRow
	query := Rebind(`
		SELECT 
			id, name, addr, port, "user", password, private_key, host_key,
			COALESCE(status, '') as status, 
			COALESCE(arch, '') as arch, 
			initialized_at, created_at, updated_at 
		FROM ssh_hosts 
		WHERE id = ?
	`)
	err := DB.Get(&host, query, id)
	if err != nil {
		return nil, err
	}

	// Decrypt credentials so handlers can use them for SSH
	if err := host.DecryptCredentials(); err != nil {
		// Log but continue, same as in other parts of the system
		return &host, nil
	}

	return &host, nil
}

// CreateSSHHost creates a new SSH host with pre-encrypted credentials
func CreateSSHHost(name, addr string, port int, user string, password, privateKey, hostKey *string) (*SSHHostRow, error) {
	host := &SSHHostRow{
		ID:         uuid.New(),
		Name:       name,
		Addr:       addr,
		Port:       port,
		User:       user,
		Password:   password,
		PrivateKey: privateKey,
		HostKey:    hostKey,
		Status:     "healthy",
	}

	now := time.Now()
	host.CreatedAt = models.NullableTime{Time: &now}
	host.UpdatedAt = models.NullableTime{Time: &now}

	query := Rebind(`INSERT INTO ssh_hosts (id, name, addr, port, "user", password, private_key, host_key, status, created_at, updated_at) 
	                 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	_, err := DB.Exec(query, host.ID, host.Name, host.Addr, host.Port, host.User, host.Password, host.PrivateKey, host.HostKey, host.Status, now, now)
	if err != nil {
		return nil, err
	}

	return host, nil
}

// UpdateSSHHost updates an SSH host
func UpdateSSHHost(id uuid.UUID, name, addr string, port int, user string, password, privateKey *string) error {
	now := time.Now()

	// Build dynamic update query
	query := Rebind(`UPDATE ssh_hosts SET name = ?, addr = ?, port = ?, "user" = ?, updated_at = ?`)
	args := []interface{}{name, addr, port, user, now}

	if password != nil {
		query = Rebind(`UPDATE ssh_hosts SET name = ?, addr = ?, port = ?, "user" = ?, password = ?, updated_at = ?`)
		args = []interface{}{name, addr, port, user, *password, now}
		if privateKey != nil {
			query = Rebind(`UPDATE ssh_hosts SET name = ?, addr = ?, port = ?, "user" = ?, password = ?, private_key = ?, updated_at = ?`)
			args = []interface{}{name, addr, port, user, *password, *privateKey, now}
		}
	} else if privateKey != nil {
		query = Rebind(`UPDATE ssh_hosts SET name = ?, addr = ?, port = ?, "user" = ?, private_key = ?, updated_at = ?`)
		args = []interface{}{name, addr, port, user, *privateKey, now}
	}

	query += Rebind(" WHERE id = ?")
	args = append(args, id)

	_, err := DB.Exec(query, args...)
	return err
}

// DeleteSSHHost deletes an SSH host
func DeleteSSHHost(id uuid.UUID) error {
	query := Rebind("DELETE FROM ssh_hosts WHERE id = ?")
	_, err := DB.Exec(query, id)
	return err
}

// GetAllApplications retrieves all applications
func GetAllApplications() ([]models.Application, error) {
	var apps []models.Application
	err := DB.Select(&apps, "SELECT * FROM applications ORDER BY name ASC")
	return apps, err
}

// GetApplicationByID retrieves an application by ID
func GetApplicationByID(id uuid.UUID) (*models.Application, error) {
	var app models.Application
	query := Rebind("SELECT * FROM applications WHERE id = ?")
	err := DB.Get(&app, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}
	return &app, nil
}

// GetApplicationInstances retrieves all instances for an application
func GetApplicationInstances(appID uuid.UUID) ([]ApplicationInstanceRow, error) {
	var instances []ApplicationInstanceRow
	query := Rebind("SELECT * FROM application_instances WHERE application_id = ?")
	err := DB.Select(&instances, query, appID)
	return instances, err
}

// GetAllSecrets retrieves all decrypted secrets for an application
func GetAllSecrets(appID uuid.UUID) (map[string]string, error) {
	return GetSecretsForApp(appID)
}

// DeploymentHistoryRow represents a deployment history entry with host name and port
type DeploymentHistoryRow struct {
	ID          uuid.UUID `db:"id"`
	InstanceID  uuid.UUID `db:"instance_id"`
	Version     string    `db:"version"`
	ReleasePath string    `db:"release_path"`
	Status      string    `db:"status"`
	Output      string    `db:"log_output"`
	HostName    string    `db:"host_name"`
	Port        int       `db:"port"`
	CreatedAt   time.Time `db:"created_at"`
}

// GetDeploymentHistoryForApp retrieves deployment history for an application
func GetDeploymentHistoryForApp(appName string) ([]DeploymentHistoryRow, error) {
	var history []DeploymentHistoryRow
	query := Rebind(`
		SELECT dh.id, dh.instance_id, dh.version, dh.release_path, dh.status, 
		       COALESCE(dh.log_output, '') as log_output, h.name as host_name, 
		       COALESCE(dh.port, 0) as port, dh.created_at
		FROM deployment_history dh
		JOIN application_instances ai ON dh.instance_id = ai.id
		JOIN applications a ON ai.application_id = a.id
		JOIN ssh_hosts h ON ai.host_id = h.id
		WHERE a.name = ?
		ORDER BY dh.created_at DESC
		LIMIT 50
	`)
	err := DB.Select(&history, query, appName)
	return history, err
}

// GetDeploymentHistoryByID retrieves a single deployment history entry
func GetDeploymentHistoryByID(id uuid.UUID) (*DeploymentHistoryRow, error) {
	var history DeploymentHistoryRow
	query := Rebind(`
		SELECT dh.id, dh.instance_id, dh.version, dh.release_path, dh.status, 
		       COALESCE(dh.log_output, '') as log_output, h.name as host_name, 
		       COALESCE(dh.port, 0) as port, dh.created_at
		FROM deployment_history dh
		JOIN application_instances ai ON dh.instance_id = ai.id
		JOIN ssh_hosts h ON ai.host_id = h.id
		WHERE dh.id = ?
	`)
	err := DB.Get(&history, query, id)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// CreateDeploymentHistoryWithStatus creates a deployment history with custom status
func CreateDeploymentHistoryWithStatus(instanceID uuid.UUID, version, status, output string) (*models.DeploymentHistory, error) {
	now := time.Now()
	history := &models.DeploymentHistory{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		Version:     version,
		ReleasePath: "",
		Status:      models.DeploymentStatus(status),
		CreatedAt:   models.NullableTime{Time: &now},
	}
	query := `INSERT INTO deployment_history (id, instance_id, version, release_path, status, log_output, created_at) VALUES (:id, :instance_id, :version, :release_path, :status, :log_output, :created_at)`
	_, err := DB.NamedExec(query, map[string]interface{}{
		"id":           history.ID,
		"instance_id":  history.InstanceID,
		"version":      history.Version,
		"release_path": history.ReleasePath,
		"status":       history.Status,
		"log_output":   output,
		"created_at":   history.CreatedAt,
	})
	return history, err
}

// AppendDeploymentHistoryOutput appends to the log output of a deployment
func AppendDeploymentHistoryOutput(id uuid.UUID, output string) error {
	query := Rebind("UPDATE deployment_history SET log_output = COALESCE(log_output, '') || ? WHERE id = ?")
	_, err := DB.Exec(query, output, id)
	return err
}

// UpdateDeploymentHistoryStatusOnly updates only the status of a deployment
func UpdateDeploymentHistoryStatusOnly(id uuid.UUID, status string) error {
	now := time.Now()
	if status == "success" {
		query := Rebind("UPDATE deployment_history SET status = ?, updated_at = ?, deployed_at = ? WHERE id = ?")
		_, err := DB.Exec(query, status, now, now, id)
		return err
	}
	query := Rebind("UPDATE deployment_history SET status = ?, updated_at = ? WHERE id = ?")
	_, err := DB.Exec(query, status, now, id)
	return err
}
