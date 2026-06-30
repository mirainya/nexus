-- Widen api_keys.key column.
-- generateKey() produces "nxk_" + hex(32 bytes) = 68 chars, but the column was
-- VARCHAR(64). SQLite (dev) ignores varchar length so it was never caught;
-- PostgreSQL (prod) enforces it and rejects inserts with SQLSTATE 22001.
ALTER TABLE api_keys ALTER COLUMN key TYPE VARCHAR(128);
