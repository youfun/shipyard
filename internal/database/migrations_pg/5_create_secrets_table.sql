-- +migrate Up
CREATE TABLE IF NOT EXISTS secrets (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(application_id) REFERENCES applications(id),
    UNIQUE(application_id, key)
);

-- +migrate Down
DROP TABLE IF EXISTS secrets;
