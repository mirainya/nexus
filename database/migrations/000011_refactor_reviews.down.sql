DROP INDEX IF EXISTS idx_reviews_document_id;
ALTER TABLE reviews DROP COLUMN IF EXISTS document_id;
