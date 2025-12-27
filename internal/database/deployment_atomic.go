package database

import (
	"youfun/shipyard/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RecordSuccessfulDeployment handles the complex logic of recording a successful deployment.
// It updates deployment_history, creates a deployment_instance, and updates application_instance ports.
func RecordSuccessfulDeployment(deploymentID uuid.UUID, port int, releasePath string, gitCommitSHA string) error {
	// 1. Get the deployment history record to know instance_id and version
	var history models.DeploymentHistory
	err := DB.Get(&history, "SELECT * FROM deployment_history WHERE id = ?", deploymentID)
	if err != nil {
		return fmt.Errorf("failed to find deployment history %s: %w", deploymentID, err)
	}

	// 2. Start a transaction
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	// 3. Update Deployment History
	queryHistory := Rebind("UPDATE deployment_history SET status = 'success', release_path = ?, port = ?, deployed_at = ?, updated_at = ? WHERE id = ?")
	if _, err := tx.Exec(queryHistory, releasePath, port, now, now, deploymentID); err != nil {
		return fmt.Errorf("failed to update deployment history: %w", err)
	}

	// 4. Create Deployment Instance
	// Use the version from history, but allow release_path and git_commit_sha from CLI
	runID := uuid.New()
	queryInstance := Rebind(`INSERT INTO deployment_instances 
		(id, application_instance_id, version, git_commit_sha, release_path, port, status, started_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 'running', ?, ?)`)

	_, err = tx.Exec(queryInstance,
		runID,
		history.InstanceID,
		history.Version,
		gitCommitSHA,
		releasePath,
		port,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment instance: %w", err)
	}

	// 5. Get current active port (to be previous active port)
	var currentInstance models.ApplicationInstance
	if err := tx.Get(&currentInstance, Rebind("SELECT * FROM application_instances WHERE id = ?"), history.InstanceID); err != nil {
		return fmt.Errorf("failed to get application instance: %w", err)
	}

	oldPort := int64(0)
	if currentInstance.ActivePort.Valid {
		oldPort = currentInstance.ActivePort.Int64
	}

	// 6. Update Application Instance ports
	// Set previous_active_port = current active_port
	// Set active_port = new port
	queryAppInst := Rebind("UPDATE application_instances SET active_port = ?, previous_active_port = ?, updated_at = ? WHERE id = ?")
	if _, err := tx.Exec(queryAppInst, port, oldPort, now, history.InstanceID); err != nil {
		return fmt.Errorf("failed to update application instance ports: %w", err)
	}

	// 7. Update status of the NEW deployment instance to 'active'
	// Wait, deployment_instances status logic:
	// 'running' -> initially created
	// 'active' -> when it's taking traffic (which is now)
	// So we should just set it to 'active' initially or update it.
	// Let's update it to match legacy logic (AddDeploymentInstance sets 'running', then later Update to 'active').
	// But here we are doing it all at once when success is reported. So 'active' is appropriate.
	// Actually, the INSERT above set it to 'running'. Let's update it to 'active'.
	queryUpdateStatus := Rebind("UPDATE deployment_instances SET status = 'active' WHERE id = ?")
	if _, err := tx.Exec(queryUpdateStatus, runID); err != nil {
		return fmt.Errorf("failed to mark new instance as active: %w", err)
	}

	// 8. Mark old instance as standby (if any)
	if oldPort > 0 {
		// Find the deployment instance for the old port
		// Note: This logic matches legacy `database.GetLatestDeploymentInstanceByPort` + update
		queryOldRun := Rebind(`SELECT id FROM deployment_instances WHERE application_instance_id = ? AND port = ? ORDER BY started_at DESC LIMIT 1`)
		var oldRunID string
		if err := tx.Get(&oldRunID, queryOldRun, history.InstanceID, oldPort); err == nil {
			// Found it, update status to standby
			_, _ = tx.Exec(Rebind("UPDATE deployment_instances SET status = 'standby' WHERE id = ?"), oldRunID)
		}
	}

	return tx.Commit()
}
