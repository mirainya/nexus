-- Initial schema for Nexus
-- Generated from GORM models

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin'
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(100),
    key VARCHAR(64) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at);

CREATE TABLE IF NOT EXISTS llm_providers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    base_url VARCHAR(500),
    api_key VARCHAR(500),
    default_model VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_llm_providers_name ON llm_providers(name);
CREATE INDEX IF NOT EXISTS idx_llm_providers_deleted_at ON llm_providers(deleted_at);

CREATE TABLE IF NOT EXISTS prompt_templates (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(100),
    description VARCHAR(500),
    content TEXT,
    variables JSONB,
    version INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_prompt_templates_deleted_at ON prompt_templates(deleted_at);

CREATE TABLE IF NOT EXISTS pipelines (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pipelines_name ON pipelines(name);
CREATE INDEX IF NOT EXISTS idx_pipelines_deleted_at ON pipelines(deleted_at);

CREATE TABLE IF NOT EXISTS pipeline_steps (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    pipeline_id BIGINT NOT NULL REFERENCES pipelines(id),
    sort_order INT NOT NULL DEFAULT 0,
    processor_type VARCHAR(50),
    prompt_template_id BIGINT REFERENCES prompt_templates(id),
    config JSONB,
    condition VARCHAR(500),
    on_error VARCHAR(20) NOT NULL DEFAULT 'stop',
    max_retry INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_pipeline_steps_pipeline_id ON pipeline_steps(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_steps_deleted_at ON pipeline_steps(deleted_at);

CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    uuid VARCHAR(36) NOT NULL,
    type VARCHAR(50),
    content TEXT,
    source_url VARCHAR(1024),
    metadata JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_path VARCHAR(512)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_uuid ON documents(uuid);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(type);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_deleted_at ON documents(deleted_at);

CREATE TABLE IF NOT EXISTS entities (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    uuid VARCHAR(36) NOT NULL,
    type VARCHAR(50),
    name VARCHAR(255),
    aliases JSONB,
    attributes JSONB,
    confidence DECIMAL(5,4),
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    source_id BIGINT,
    evidence JSONB
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_entities_uuid ON entities(uuid);
CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);
CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name);
CREATE INDEX IF NOT EXISTS idx_entities_source_id ON entities(source_id);
CREATE INDEX IF NOT EXISTS idx_entities_deleted_at ON entities(deleted_at);

CREATE TABLE IF NOT EXISTS relations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    uuid VARCHAR(36) NOT NULL,
    from_entity_id BIGINT NOT NULL REFERENCES entities(id),
    to_entity_id BIGINT NOT NULL REFERENCES entities(id),
    type VARCHAR(100),
    metadata JSONB,
    confidence DECIMAL(5,4),
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    source_id BIGINT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_relations_uuid ON relations(uuid);
CREATE INDEX IF NOT EXISTS idx_relations_from_entity_id ON relations(from_entity_id);
CREATE INDEX IF NOT EXISTS idx_relations_to_entity_id ON relations(to_entity_id);
CREATE INDEX IF NOT EXISTS idx_relations_type ON relations(type);
CREATE INDEX IF NOT EXISTS idx_relations_source_id ON relations(source_id);
CREATE INDEX IF NOT EXISTS idx_relations_deleted_at ON relations(deleted_at);

CREATE TABLE IF NOT EXISTS jobs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    uuid VARCHAR(36) NOT NULL,
    document_id BIGINT NOT NULL REFERENCES documents(id),
    pipeline_id BIGINT NOT NULL REFERENCES pipelines(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result JSONB,
    callback_url VARCHAR(1024),
    current_step INT NOT NULL DEFAULT 0,
    total_steps INT NOT NULL DEFAULT 0,
    error TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_jobs_uuid ON jobs(uuid);
CREATE INDEX IF NOT EXISTS idx_jobs_document_id ON jobs(document_id);
CREATE INDEX IF NOT EXISTS idx_jobs_pipeline_id ON jobs(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_deleted_at ON jobs(deleted_at);

CREATE TABLE IF NOT EXISTS job_step_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    job_id BIGINT NOT NULL REFERENCES jobs(id),
    step_order INT,
    processor_type VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error TEXT,
    tokens INT NOT NULL DEFAULT 0,
    cost DOUBLE PRECISION NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_job_step_logs_job_id ON job_step_logs(job_id);
CREATE INDEX IF NOT EXISTS idx_job_step_logs_deleted_at ON job_step_logs(deleted_at);

CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    entity_id BIGINT REFERENCES entities(id),
    relation_id BIGINT REFERENCES relations(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    original_data JSONB,
    modified_data JSONB,
    reviewer VARCHAR(100)
);
CREATE INDEX IF NOT EXISTS idx_reviews_entity_id ON reviews(entity_id);
CREATE INDEX IF NOT EXISTS idx_reviews_relation_id ON reviews(relation_id);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
CREATE INDEX IF NOT EXISTS idx_reviews_deleted_at ON reviews(deleted_at);
