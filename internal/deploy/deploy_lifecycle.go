package deploy

import (
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"fmt"
	"log"
	"time"
)

// startNewVersion starts the new version of the application on a free port.
// This encapsulates the logic of directory creation, symlinking, permission fixing, and service starting.
func (d *Deployer) startNewVersion(releasePath string) (*models.DeploymentInstance, error) {
	log.Println("🌱 Starting new version...")
	greenPort, err := d.findFreePort()
	if err != nil {
		return nil, err
	}
	log.Printf("Found free port on remote host: %d", greenPort)

	instancesDir := fmt.Sprintf("/var/www/%s/instances", d.AppName)
	if err := d.executeRemoteCommand(fmt.Sprintf("mkdir -p %s", instancesDir), false); err != nil {
		return nil, err
	}
	if err := d.executeRemoteCommand(fmt.Sprintf("ln -sfn %s %s/%d", releasePath, instancesDir, greenPort), false); err != nil {
		return nil, err
	}
	if err := d.executeRemoteCommand(fmt.Sprintf("(chown -h phoenix:phoenix %s/%d || true) && (chown -R phoenix:phoenix %s || true)", instancesDir, greenPort, releasePath), true); err != nil {
		return nil, err
	}
	if err := d.executeRemoteCommand(fmt.Sprintf("systemctl start %s@%d", d.AppName, greenPort), true); err != nil {
		return nil, err
	}

	now := time.Now()
	run := &models.DeploymentInstance{
		ApplicationInstanceID: d.Instance.ID,
		Version:               d.Version,
		GitCommitSHA:          d.GitCommitSHA,
		ReleasePath:           releasePath,
		Port:                  greenPort,
		Status:                "running",
		StartedAt:             models.NullableTime{Time: &now},
		CreatedAt:             models.NullableTime{Time: &now},
	}
	// Note: We do NOT save to database here anymore, to support CLI mode (which has no DB connection).
	// The caller is responsible for persistence (saving to DB or calling API).

	return run, nil
}

// performHealthCheck performs a health check on the application running on the given port.
func (d *Deployer) performHealthCheck(port int) error {
	log.Printf("Executing service health check (Port: %d)", port)
	// Use systemctl is-active to check if service started successfully and is running.
	healthCheckCmd := fmt.Sprintf("systemctl is-active --quiet %s@%d", d.AppName, port)

	for i := 0; i < 10; i++ {
		if err := d.executeRemoteCommand(healthCheckCmd, false); err == nil {
			log.Println("✅ Service health check passed (systemd service is active).")
			return nil
		}

		log.Printf("Health check attempt %d/10 failed, retrying in 2 seconds...", i+1)
		time.Sleep(2 * time.Second)
	}

	// If loop ends without success, log detailed info and return error
	log.Println("Last health check attempt failed, outputting detailed status of systemd unit:")
	debugCmd := fmt.Sprintf("systemctl status %s@%d", d.AppName, port)
	_ = d.executeRemoteCommand(debugCmd, true) // Run once to log output

	return fmt.Errorf("Service failed to enter 'active' state after multiple attempts")
}

// stopOldVersion stops the old version of the application.
// This replaces the inline logic for stopping old ports.
func (d *Deployer) stopOldVersion(greenPort int) error {
	oldPort := 0
	if d.Instance.ActivePort.Valid && d.Instance.ActivePort.Int64 > 0 {
		oldPort = int(d.Instance.ActivePort.Int64)
	}

	if err := database.UpdateInstancePortsForRollback(d.Instance.ID, greenPort, oldPort); err != nil {
		return fmt.Errorf("failed to update active_port and previous_active_port in database: %w", err)
	}
	log.Printf("Database updated: active_port -> %d, previous_active_port -> %d", greenPort, oldPort)

	if oldPort > 0 {
		log.Println("--- 11. Handling old version ---")
		if oldRun, err := database.GetLatestDeploymentInstanceByPort(d.Instance.ID, oldPort); err == nil {
			_ = database.UpdateDeploymentInstanceStatus(oldRun.ID, "standby", nil)
			log.Printf("Old version (Port %d) marked as 'standby'. Stopping service in 5s to save resources...", oldPort)
			time.Sleep(5 * time.Second)
			_ = d.executeRemoteCommand(fmt.Sprintf("systemctl disable %s@%d", d.AppName, oldPort), true)
			_ = d.executeRemoteCommand(fmt.Sprintf("systemctl stop %s@%d", d.AppName, oldPort), true)
			log.Printf("✅ Old version (Port %d) service stopped, but files are kept for quick rollback.", oldPort)
		}
	}
	return nil
}

// cleanupStaleInstances cleans up stale deployment instances.
func (d *Deployer) cleanupStaleInstances(greenPort int) error {
	oldPort := 0
	if d.Instance.ActivePort.Valid && d.Instance.ActivePort.Int64 > 0 {
		oldPort = int(d.Instance.ActivePort.Int64)
	}

	staleInstances, err := database.GetStaleDeploymentInstances(d.Instance.ID, greenPort, oldPort)
	if err != nil {
		log.Printf("⚠️ Failed to get stale instance list: %v", err)
		return nil
	}

	if len(staleInstances) > 0 {
		log.Printf("Found %d stale instances, cleaning up...", len(staleInstances))
		for _, stale := range staleInstances {
			log.Printf("Stopping stale instance (Port %d)...", stale.Port)
			d.executeRemoteCommand(fmt.Sprintf("systemctl disable %s@%d || true", d.AppName, stale.Port), true)
			d.executeRemoteCommand(fmt.Sprintf("systemctl stop %s@%d || true", d.AppName, stale.Port), true)
			d.executeRemoteCommand(fmt.Sprintf("rm -f /var/www/%s/instances/%d || true", d.AppName, stale.Port), false)
			st := time.Now()
			_ = database.UpdateDeploymentInstanceStatus(stale.ID, "stopped", &st)
		}
		log.Println("✅ Stale instance cleanup completed.")
	} else {
		log.Println("No stale instances to clean up.")
	}

	return nil
}
