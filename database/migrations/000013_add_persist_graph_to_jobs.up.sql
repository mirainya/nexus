-- Add persist_graph column to jobs (controls whether entities/relations are persisted)
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS persist_graph BOOLEAN NOT NULL DEFAULT FALSE;
