-- Key-level isolation: add api_key_id to business data tables.
-- One API key = one project; data is scoped to the submitting key.
-- NULL = admin/internal submissions (visible only via admin JWT endpoints).
ALTER TABLE documents ADD COLUMN IF NOT EXISTS api_key_id BIGINT;
ALTER TABLE entities ADD COLUMN IF NOT EXISTS api_key_id BIGINT;
ALTER TABLE relations ADD COLUMN IF NOT EXISTS api_key_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_documents_api_key_id ON documents(api_key_id);
CREATE INDEX IF NOT EXISTS idx_entities_api_key_id ON entities(api_key_id);
CREATE INDEX IF NOT EXISTS idx_relations_api_key_id ON relations(api_key_id);
