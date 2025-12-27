package database

import (
	"database/sql"
	"errors"
	"fmt"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// GetLatestDeploymentHistory retrieves the latest deployment history for a given application.
func GetLatestDeploymentHistory(appID uuid.UUID) (*models.DeploymentHistory, error) {
	var history models.DeploymentHistory
	query := Rebind(`
		SELECT dh.*
		FROM deployment_history dh
		JOIN application_instances ai ON dh.instance_id = ai.id
		WHERE ai.application_id = ?
		ORDER BY dh.created_at DESC
		LIMIT 1
	`)
	err := DB.Get(&history, query, appID)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// --- deployment_instances Table Operations ---

func AddDeploymentInstance(run *models.DeploymentInstance) error {
	run.ID = uuid.New()
	now := time.Now()
	run.CreatedAt = models.NullableTime{Time: &now}
	run.StartedAt = models.NullableTime{Time: &now}
	query := `INSERT INTO deployment_instances (id, application_instance_id, version, git_commit_sha, release_path, port, status, started_at, created_at)
	          VALUES (:id, :application_instance_id, :version, :git_commit_sha, :release_path, :port, :status, :started_at, :created_at)`
	_, err := DB.NamedExec(query, run)
	return err
}

func UpdateDeploymentInstanceStatus(id uuid.UUID, status string, stoppedAt *time.Time) error {
	if stoppedAt != nil {
		query := Rebind("UPDATE deployment_instances SET status = ?, stopped_at = ? WHERE id = ?")
		_, err := DB.Exec(query, status, *stoppedAt, id)
		return err
	}
	query := Rebind("UPDATE deployment_instances SET status = ? WHERE id = ?")
	_, err := DB.Exec(query, status, id)
	return err
}

func GetRecentDeploymentInstances(appInstanceID uuid.UUID, limit int) ([]models.DeploymentInstance, error) {
	var runs []models.DeploymentInstance
	query := Rebind(`SELECT * FROM deployment_instances WHERE application_instance_id = ? ORDER BY started_at DESC LIMIT ?`)
	err := DB.Select(&runs, query, appInstanceID, limit)
	return runs, err
}

// GetDeploymentHistoryForInstance retrieves all deployment history for a given application_instance.
// Ordered by time in descending order, with the newest first.
func GetDeploymentHistoryForInstance(instanceID uuid.UUID, limit int) ([]models.DeploymentInstance, error) {
	var instances []models.DeploymentInstance
	query := Rebind(`SELECT * FROM deployment_instances WHERE application_instance_id = ? ORDER BY started_at DESC LIMIT ?`)
	err := DB.Select(&instances, query, instanceID, limit)
	return instances, err
}

// GetLastDeploymentInstance retrieves the last deployment instance for a given application_instance.
// Regardless of success or failure.
func GetLastDeploymentInstance(appInstanceID uuid.UUID) (*models.DeploymentInstance, error) {
	var instance models.DeploymentInstance
	query := Rebind(`SELECT * FROM deployment_instances WHERE application_instance_id = ? ORDER BY started_at DESC LIMIT 1`)
	err := DB.Get(&instance, query, appInstanceID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no deployment records found")
	}
	return &instance, err
}

func GetLatestDeploymentInstanceByPort(appInstanceID uuid.UUID, port int) (*models.DeploymentInstance, error) {
	var run models.DeploymentInstance
	query := Rebind(`SELECT * FROM deployment_instances WHERE application_instance_id = ? AND port = ? ORDER BY started_at DESC LIMIT 1`)
	err := DB.Get(&run, query, appInstanceID, port)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no instance record found for application_instance_id=%s port=%d", appInstanceID, port)
	}
	return &run, err
}

// GetStaleDeploymentInstances finds instances that are not active or standby, so they can be cleaned up.
func GetStaleDeploymentInstances(instanceID uuid.UUID, activePort int, standbyPort int) ([]models.DeploymentInstance, error) {
	var instances []models.DeploymentInstance
	// We need to handle the case where standbyPort might be 0 (first deployment)
	var args []interface{}
	args = append(args, instanceID, activePort)

	query := `SELECT * FROM deployment_instances
			  WHERE application_instance_id = ?
			  AND port != ?`

	if standbyPort > 0 {
		query += " AND port != ?"
		args = append(args, standbyPort)
	}
	query += " AND status NOT IN ('stopped', 'failed')"
	query = Rebind(query)

	err := DB.Select(&instances, query, args...)
	return instances, err
}

// --- deployment_history Table Operations ---
func CreateDeploymentHistory(instanceID uuid.UUID, version, releasePath string) (*models.DeploymentHistory, error) {
	now := time.Now()
	history := &models.DeploymentHistory{
		ID:          uuid.New(),
		InstanceID:  instanceID,
		Version:     version,
		ReleasePath: releasePath,
		Status:      models.DeploymentStatusPending,
		CreatedAt:   models.NullableTime{Time: &now},
	}
	query := `INSERT INTO deployment_history (id, instance_id, version, release_path, status, created_at) VALUES (:id, :instance_id, :version, :release_path, :status, :created_at)`
	_, err := DB.NamedExec(query, history)
	return history, err
}

func UpdateDeploymentHistoryStatus(id uuid.UUID, status models.DeploymentStatus, logOutput string) error {
	now := time.Now()
	var err error
	if status == models.DeploymentStatusSuccess {
		query := Rebind("UPDATE deployment_history SET status = ?, log_output = ?, updated_at = ?, deployed_at = ? WHERE id = ?")
		_, err = DB.Exec(query, status, logOutput, now, now, id)
	} else {
		query := Rebind("UPDATE deployment_history SET status = ?, log_output = ?, updated_at = ? WHERE id = ?")
		_, err = DB.Exec(query, status, logOutput, now, id)
	}
	return err
}

// GetLastSuccessfulHostForApp retrieves the host name and time of the last successful deployment for a given app.
func GetLastSuccessfulHostForApp(appID uuid.UUID) (string, time.Time, error) {
	var result struct {
		HostName   string              `db:"name"`
		DeployedAt models.NullableTime `db:"deployed_at"`
	}
	query := Rebind(`
		SELECT
			h.name,
			dh.deployed_at
		FROM deployment_history dh
		JOIN application_instances ai ON dh.instance_id = ai.id
		JOIN ssh_hosts h ON ai.host_id = h.id
		WHERE ai.application_id = ? AND dh.status = 'success' AND dh.deployed_at IS NOT NULL
		ORDER BY dh.deployed_at DESC
		LIMIT 1
	`)
	err := DB.Get(&result, query, appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// It's not an error if there's no previous deployment, return zero values
			return "", time.Time{}, nil
		}
		return "", time.Time{}, err
	}

	if result.DeployedAt.Time == nil {
		return result.HostName, time.Time{}, nil
	}
	return result.HostName, *result.DeployedAt.Time, nil
}

// GetLatestDeploymentInstanceForApp retrieves the latest deployment instance for a given application across all hosts.
// It returns the deployment instance and the associated host name.
func GetLatestDeploymentInstanceForApp(appName string) (*models.DeploymentInstance, string, error) {
	var result struct {
		models.DeploymentInstance
		HostName string `db:"host_name"`
	}

	query := Rebind(`
		SELECT di.*, h.name as host_name
		FROM deployment_instances di
		JOIN application_instances ai ON di.application_instance_id = ai.id
		JOIN applications a ON ai.application_id = a.id
		JOIN ssh_hosts h ON ai.host_id = h.id
		WHERE a.name = ?
		ORDER BY di.started_at DESC
		LIMIT 1
	`)
	err := DB.Get(&result, query, appName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("no deployment instances found for app '%s'", appName)
		}
		return nil, "", err
	}

	return &result.DeploymentInstance, result.HostName, nil
}

// GetLatestDeploymentHistoryForApp retrieves the latest deployment history for a given application by name.
func GetLatestDeploymentHistoryForApp(appName string) (*models.DeploymentHistory, error) {
	var history models.DeploymentHistory
	query := Rebind(`
		SELECT dh.*
		FROM deployment_history dh
		JOIN application_instances ai ON dh.instance_id = ai.id
		JOIN applications a ON ai.application_id = a.id
		WHERE a.name = ?
		ORDER BY dh.created_at DESC
		LIMIT 1
	`)
	err := DB.Get(&history, query, appName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetDeploymentsCount returns the total number of deployments in the system
func GetDeploymentsCount() (int, error) {
	var count int
	err := DB.Get(&count, "SELECT COUNT(*) FROM deployment_history")
	return count, err
}

// RecentDeploymentRow represents a recent deployment with full context for dashboard
type RecentDeploymentRow struct {
	ID           uuid.UUID  `db:"id"`
	AppName      string     `db:"app_name"`
	HostName     string     `db:"host_name"`
	HostAddr     string     `db:"host_addr"`
	Version      string     `db:"version"`
	GitCommitSHA string     `db:"git_commit_sha"`
	Status       string     `db:"status"`
	Port         int        `db:"port"`
	DeployedAt   *time.Time `db:"deployed_at"`
}

// GetRecentDeploymentsGlobal retrieves the most recent deployments across all applications
func GetRecentDeploymentsGlobal(limit int) ([]RecentDeploymentRow, error) {
	var deployments []RecentDeploymentRow
	query := Rebind(`
		SELECT 
			dh.id,
			a.name as app_name,
			h.name as host_name,
			h.addr as host_addr,
			dh.version,
			COALESCE((
				SELECT di.git_commit_sha 
				FROM deployment_instances di 
				WHERE di.application_instance_id = ai.id 
				  AND di.version = dh.version 
				ORDER BY di.created_at DESC 
				LIMIT 1
			), '') as git_commit_sha,
			dh.status,
			COALESCE(dh.port, 0) as port,
			dh.deployed_at
		FROM deployment_history dh
		JOIN application_instances ai ON dh.instance_id = ai.id
		JOIN applications a ON ai.application_id = a.id
		JOIN ssh_hosts h ON ai.host_id = h.id
		WHERE dh.deployed_at IS NOT NULL
		ORDER BY dh.deployed_at DESC
		LIMIT ?
	`)
	err := DB.Select(&deployments, query, limit)
	return deployments, err
}
