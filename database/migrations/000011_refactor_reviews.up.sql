-- 审核系统重构：添加文档级审核支持
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS document_id BIGINT REFERENCES documents(id);
CREATE INDEX IF NOT EXISTS idx_reviews_document_id ON reviews(document_id);
