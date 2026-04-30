-- Add multi-tenant support

CREATE TABLE IF NOT EXISTS tenants (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    uuid VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_uuid ON tenants(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_deleted_at ON tenants(deleted_at);

-- Default tenant for existing data
INSERT INTO tenants (uuid, name, active) VALUES ('00000000-0000-0000-0000-000000000000', 'default', true)
    ON CONFLICT (uuid) DO NOTHING;

-- Add tenant_id to existing tables (IF NOT EXISTS for idempotency)
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id);
ALTER TABLE documents ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE relations ADD COLUMN IF NOT EXISTS tenant_id BIGINT;
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

-- Migrate existing data to default tenant
UPDATE api_keys SET tenant_id = (SELECT id FROM tenants WHERE uuid = '00000000-0000-0000-0000-000000000000') WHERE tenant_id IS NULL;
UPDATE documents SET tenant_id = (SELECT id FROM tenants WHERE uuid = '00000000-0000-0000-0000-000000000000') WHERE tenant_id IS NULL;
UPDATE jobs SET tenant_id = (SELECT id FROM tenants WHERE uuid = '00000000-0000-0000-0000-000000000000') WHERE tenant_id IS NULL;
UPDATE entities SET tenant_id = (SELECT id FROM tenants WHERE uuid = '00000000-0000-0000-0000-000000000000') WHERE tenant_id IS NULL;
UPDATE relations SET tenant_id = (SELECT id FROM tenants WHERE uuid = '00000000-0000-0000-0000-000000000000') WHERE tenant_id IS NULL;
UPDATE reviews SET tenant_id = (SELECT id FROM tenants WHERE uuid = '00000000-0000-0000-0000-000000000000') WHERE tenant_id IS NULL;

-- Set NOT NULL constraints
ALTER TABLE api_keys ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE documents ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE jobs ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE entities ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE relations ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE reviews ALTER COLUMN tenant_id SET NOT NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_jobs_tenant_id ON jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_entities_tenant_id ON entities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_relations_tenant_id ON relations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reviews_tenant_id ON reviews(tenant_id);
