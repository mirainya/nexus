-- Refactor calling system to API-Key-centric model: drop Tenant and Credential layers.
-- Business data (documents/jobs/entities/relations/reviews) becomes platform-global.
-- Billing/quota stays keyed by api_key_id. Future grouping can derive from api_key.

-- 1. Drop credentials table (BYOK removed).
DROP TABLE IF EXISTS credentials;

-- 2. Drop tenant_id foreign keys / columns from dependent tables (drop FK before column).
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_tenant_id_fkey;
ALTER TABLE api_keys DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE documents DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE entities DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE relations DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE reviews DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE jobs DROP COLUMN IF EXISTS tenant_id;

-- 3. Drop credential_id from jobs.
ALTER TABLE jobs DROP COLUMN IF EXISTS credential_id;

-- 4. Drop tenants table last (after all FKs removed).
DROP TABLE IF EXISTS tenants;
