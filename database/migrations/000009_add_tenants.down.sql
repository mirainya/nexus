DROP INDEX IF EXISTS idx_reviews_tenant_id;
DROP INDEX IF EXISTS idx_relations_tenant_id;
DROP INDEX IF EXISTS idx_entities_tenant_id;
DROP INDEX IF EXISTS idx_jobs_tenant_id;
DROP INDEX IF EXISTS idx_documents_tenant_id;
DROP INDEX IF EXISTS idx_api_keys_tenant_id;

ALTER TABLE reviews DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE relations DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE entities DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE jobs DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE documents DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE api_keys DROP COLUMN IF EXISTS tenant_id;

DROP TABLE IF EXISTS tenants;
