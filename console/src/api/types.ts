export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedData<T> {
  list: T[];
  total: number;
}

// Auth
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  username: string;
}

// Pipeline
export interface Pipeline {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  active: boolean;
  steps?: PipelineStep[];
}

export interface PipelineStep {
  id: number;
  pipeline_id: number;
  sort_order: number;
  processor_type: string;
  prompt_template_id?: number;
  config?: Record<string, unknown>;
  condition: string;
  on_error: string;
  max_retry: number;
  parallel_group: number;
  prompt_template?: PromptTemplate;
}

// Prompt Template
export interface PromptTemplate {
  id: number;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  content: string;
  variables?: Record<string, unknown>;
  version: number;
}

// Document
export interface Document {
  id: number;
  uuid: string;
  type: string;
  content: string;
  source_url: string;
  metadata?: Record<string, unknown>;
  status: string;
  file_path: string;
}

// Job
export interface JobSubmitRequest {
  content?: string;
  type: string;
  source_url?: string;
  pipeline_id: number;
  callback_url?: string;
  skip_cache?: boolean;
  metadata?: Record<string, unknown>;
}

export interface JobSubmitResponse {
  job_id: string;
  status: string;
  cached?: boolean;
  result?: unknown;
}

export interface Job {
  id: number;
  uuid: string;
  document_id: number;
  pipeline_id: number;
  status: string;
  content_hash: string;
  result?: unknown;
  callback_url: string;
  current_step: number;
  total_steps: number;
  error?: string;
  created_at: string;
  updated_at: string;
  document?: Document;
  step_logs?: JobStepLog[];
}

export interface JobStepLog {
  id: number;
  job_id: number;
  step_order: number;
  processor_type: string;
  status: string;
  started_at?: string;
  finished_at?: string;
  error?: string;
  tokens: number;
  cost: number;
}

// Entity
export interface Entity {
  id: number;
  uuid: string;
  type: string;
  name: string;
  aliases?: string[];
  attributes?: Record<string, unknown>;
  confidence: number;
  confirmed: boolean;
  source_id: number;
  evidence?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// Relation
export interface Relation {
  id: number;
  uuid: string;
  from_entity_id: number;
  to_entity_id: number;
  type: string;
  metadata?: Record<string, unknown>;
  confidence: number;
  confirmed: boolean;
  source_id: number;
  from_entity?: Entity;
  to_entity?: Entity;
}

// Review
export interface Review {
  id: number;
  entity_id?: number;
  relation_id?: number;
  status: string;
  original_data?: unknown;
  modified_data?: unknown;
  reviewer: string;
  created_at: string;
  updated_at: string;
}

// LLM Provider
export interface LLMProvider {
  id: number;
  name: string;
  display_name: string;
  base_url: string;
  default_model: string;
  input_price: number;
  output_price: number;
  max_concurrency: number;
  active: boolean;
  is_default: boolean;
}

export interface LLMProviderCreateRequest {
  name: string;
  display_name?: string;
  base_url: string;
  api_key: string;
  default_model: string;
  input_price?: number;
  output_price?: number;
  max_concurrency?: number;
  active?: boolean;
}

export interface ModelInfo {
  id: string;
  owned_by?: string;
}

// Recommend
export interface RecommendItem {
  document_id: number;
  source_url: string;
  content: string;
  scene: string;
  score: number;
  reason: string;
  tags?: string[];
}

// Search
export interface SearchRequest {
  query: string;
}

// Upload
export interface UploadResponse {
  url: string;
  path: string;
}
