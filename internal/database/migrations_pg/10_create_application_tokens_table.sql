-- +migrate Up
CREATE TABLE IF NOT EXISTS application_tokens (
    id UUID PRIMARY KEY,
    application_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    token_hash VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NULL,
    last_used_at TIMESTAMP NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

CREATE INDEX idx_application_tokens_application_id ON application_tokens(application_id);
CREATE INDEX idx_application_tokens_is_active ON application_tokens(is_active);

-- +migrate Down
DROP TABLE IF EXISTS application_tokens;
