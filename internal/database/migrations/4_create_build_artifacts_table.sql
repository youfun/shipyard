-- +migrate Up
CREATE TABLE IF NOT EXISTS build_artifacts (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    version TEXT NOT NULL,
    git_commit_sha TEXT NOT NULL,
    md5_hash TEXT NOT NULL UNIQUE,
    local_path TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS build_artifacts;
