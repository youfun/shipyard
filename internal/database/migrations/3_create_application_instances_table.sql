-- +migrate Up
CREATE TABLE IF NOT EXISTS application_instances (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    host_id TEXT NOT NULL,
    status TEXT NOT NULL,
    active_port INTEGER,
    previous_active_port INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id),
    FOREIGN KEY (host_id) REFERENCES ssh_hosts(id)
);

-- +migrate Down
DROP TABLE IF EXISTS application_instances;
