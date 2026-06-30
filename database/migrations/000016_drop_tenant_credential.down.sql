-- Reverse: recreate tenants/credentials and re-add columns (best-effort rollback).
CREATE TABLE IF NOT EXISTS tenants (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    uuid VARCHAR(36),
    name VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    monthly_request_limit INT NOT NULL DEFAULT 0,
    monthly_token_limit BIGINT NOT NULL DEFAULT 0
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE relations ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS credential_id BIGINT;

CREATE TABLE IF NOT EXISTS credentials (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    api_key_id BIGINT,
    name VARCHAR(100),
    provider_type VARCHAR(50),
    base_url VARCHAR(500),
    encrypted_key TEXT,
    default_model VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT TRUE
);
