CREATE SCHEMA IF NOT EXISTS accounts;
CREATE TABLE IF NOT EXISTS accounts.user_secrets (id SERIAL, user_id text PRIMARY KEY, secret_id text, created_at timestamp, updated_at timestamp, deleted_at timestamp);
CREATE TABLE IF NOT EXISTS accounts.user_configs (id SERIAL PRIMARY KEY, user_id VARCHAR(255) UNIQUE, attributes JSONB NOT NULL DEFAULT '{}'::JSONB, created_at timestamp, updated_at timestamp, deleted_at timestamp);
CREATE TABLE IF NOT EXISTS accounts.history_records (id SERIAL PRIMARY KEY, user_id VARCHAR(255), event_type text, message text, created_at timestamp, updated_at timestamp, deleted_at timestamp);
CREATE TABLE IF NOT EXISTS accounts.backups (id SERIAL PRIMARY KEY, user_id VARCHAR(255),credentials bytea, created_at timestamp, updated_at timestamp, deleted_at timestamp);
CREATE TABLE IF NOT EXISTS accounts.presentation_requests (id SERIAL PRIMARY KEY, user_id VARCHAR(255),request_id text, proof_request_id text, created_at timestamp, updated_at timestamp, deleted_at timestamp, ttl integer);
CREATE TABLE IF NOT EXISTS accounts.user_connections (id SERIAL PRIMARY KEY, user_id text, remote_did text, created_at timestamp, updated_at timestamp, deleted_at timestamp);