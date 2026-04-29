-- Add pricing columns to llm_providers for cost tracking
ALTER TABLE llm_providers ADD COLUMN input_price DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE llm_providers ADD COLUMN output_price DOUBLE PRECISION NOT NULL DEFAULT 0;
