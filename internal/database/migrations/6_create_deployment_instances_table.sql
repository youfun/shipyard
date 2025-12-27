-- +migrate Up
CREATE TABLE IF NOT EXISTS deployment_instances (
    id TEXT PRIMARY KEY,
    application_instance_id TEXT NOT NULL,
    version TEXT NOT NULL,
    git_commit_sha TEXT NOT NULL,
    release_path TEXT NOT NULL,
    port INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'stopped', -- e.g., running, stopped, active, standby
    started_at DATETIME,
    stopped_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_instance_id) REFERENCES application_instances(id)
);
CREATE INDEX IF NOT EXISTS idx_deployment_instances_appinst ON deployment_instances(application_instance_id);
CREATE INDEX IF NOT EXISTS idx_deployment_instances_port ON deployment_instances(port);

-- +migrate Down
DROP TABLE IF EXISTS deployment_instances;