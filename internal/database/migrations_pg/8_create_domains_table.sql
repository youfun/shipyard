-- +migrate Up
CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    application_instance_id TEXT NOT NULL,
    hostname TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(application_instance_id) REFERENCES application_instances(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_domains_hostname_instance ON domains (hostname, application_instance_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_domains_hostname_instance;
DROP TABLE IF EXISTS domains;
