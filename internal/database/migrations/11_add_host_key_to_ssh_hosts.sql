-- +migrate Up
ALTER TABLE ssh_hosts ADD COLUMN host_key TEXT;

-- +migrate Down
ALTER TABLE ssh_hosts DROP COLUMN host_key;
