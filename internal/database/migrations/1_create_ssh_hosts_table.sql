-- +migrate Up
CREATE TABLE IF NOT EXISTS ssh_hosts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    addr TEXT NOT NULL,
    port INTEGER NOT NULL,
    user TEXT NOT NULL,
    password TEXT,
    private_key TEXT,
    status TEXT,
    arch TEXT, -- such as linux/amd64, linux/arm64
    initialized_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE IF EXISTS ssh_hosts;
