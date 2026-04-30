-- 租户配额字段
ALTER TABLE tenants ADD COLUMN monthly_request_limit INT NOT NULL DEFAULT 0;
ALTER TABLE tenants ADD COLUMN monthly_token_limit BIGINT NOT NULL DEFAULT 0;
