-- +migrate Up
CREATE TABLE deployment_history (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    version TEXT NOT NULL,
    release_path TEXT NOT NULL,
    status TEXT NOT NULL,
    log_output TEXT,
    deployed_at TIMESTAMP,
    port INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(instance_id) REFERENCES application_instances(id)
);

-- +migrate Down
DROP TABLE IF EXISTS deployment_history;
