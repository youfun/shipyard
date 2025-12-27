-- +migrate Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    two_factor_secret TEXT DEFAULT '',
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE auth_devices (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE two_factor_recovery_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    code TEXT NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE device_requests (
    device_code TEXT PRIMARY KEY,
    user_code TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, authorized, rejected
    expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE device_requests;
DROP TABLE two_factor_recovery_codes;
DROP TABLE auth_devices;
DROP TABLE users;
