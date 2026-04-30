import axios from 'axios';
import type {
  LoginRequest,
  LoginResponse,
  Pipeline,
  PipelineStep,
  PromptTemplate,
  Job,
  JobSubmitRequest,
  JobSubmitResponse,
  Review,
  Entity,
  Relation,
  RecommendItem,
  LLMProvider,
  LLMProviderCreateRequest,
  ModelInfo,
  PaginatedData,
  DashboardStats,
  GraphData,
  APIKey,
  APIKeyCreateRequest,
  Credential,
  CredentialCreateRequest,
  APIKeyUsage,
  Tenant,
  TenantCreateRequest,
  TenantUpdateRequest,
} from './types';

const api = axios.create({
  baseURL: '/api',
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err.response?.data || err);
  }
);

export default api;

export const authApi = {
  login: (data: LoginRequest) => api.post<unknown, LoginResponse>('/admin/auth/login', data),
};

export const pipelineApi = {
  list: () => api.get<unknown, Pipeline[]>('/admin/pipelines'),
  get: (id: number) => api.get<unknown, Pipeline>(`/admin/pipelines/${id}`),
  create: (data: Partial<Pipeline>) => api.post<unknown, Pipeline>('/admin/pipelines', data),
  update: (id: number, data: Partial<Pipeline>) => api.put<unknown, Pipeline>(`/admin/pipelines/${id}`, data),
  delete: (id: number) => api.delete(`/admin/pipelines/${id}`),
  createStep: (id: number, data: Partial<PipelineStep>) => api.post<unknown, PipelineStep>(`/admin/pipelines/${id}/steps`, data),
  updateStep: (id: number, stepId: number, data: Partial<PipelineStep>) => api.put<unknown, PipelineStep>(`/admin/pipelines/${id}/steps/${stepId}`, data),
  deleteStep: (id: number, stepId: number) => api.delete(`/admin/pipelines/${id}/steps/${stepId}`),
  reorderSteps: (id: number, stepIds: number[]) => api.put(`/admin/pipelines/${id}/steps/reorder`, { step_ids: stepIds }),
};

export const promptApi = {
  list: () => api.get<unknown, PromptTemplate[]>('/admin/prompts'),
  get: (id: number) => api.get<unknown, PromptTemplate>(`/admin/prompts/${id}`),
  create: (data: Partial<PromptTemplate>) => api.post<unknown, PromptTemplate>('/admin/prompts', data),
  update: (id: number, data: Partial<PromptTemplate>) => api.put<unknown, PromptTemplate>(`/admin/prompts/${id}`, data),
  delete: (id: number) => api.delete(`/admin/prompts/${id}`),
};

export const jobApi = {
  list: (params?: { page?: number; page_size?: number; status?: string }) => api.get<unknown, PaginatedData<Job>>('/admin/jobs', { params }),
  submit: (data: JobSubmitRequest) => api.post<unknown, JobSubmitResponse>('/admin/jobs', data),
  status: (uuid: string) => api.get<unknown, Job>(`/admin/jobs/${uuid}`),
  recommend: (scene: string) => api.get<unknown, RecommendItem[]>('/admin/jobs/recommend', { params: { scene } }),
};

export const reviewApi = {
  list: (params?: { page?: number; page_size?: number; status?: string }) => api.get<unknown, PaginatedData<Review>>('/admin/reviews', { params }),
  approve: (id: number) => api.put(`/admin/reviews/${id}/approve`),
  reject: (id: number) => api.put(`/admin/reviews/${id}/reject`),
  modify: (id: number, data: { modified_data: unknown }) => api.put(`/admin/reviews/${id}/modify`, data),
};

export const entityApi = {
  list: (params?: { page?: number; page_size?: number; keyword?: string; type?: string }) => api.get<unknown, PaginatedData<Entity>>('/admin/entities', { params }),
  get: (id: number) => api.get<unknown, Entity>(`/admin/entities/${id}`),
  getRelations: (id: number) => api.get<unknown, Relation[]>(`/admin/entities/${id}/relations`),
  search: (keyword: string) => api.get<unknown, PaginatedData<Entity>>('/admin/entities', { params: { keyword, page_size: 50 } }),
};

export const llmProviderApi = {
  list: () => api.get<unknown, LLMProvider[]>('/admin/llm-providers'),
  create: (data: LLMProviderCreateRequest) => api.post<unknown, LLMProvider>('/admin/llm-providers', data),
  update: (id: number, data: Partial<LLMProviderCreateRequest>) => api.put<unknown, LLMProvider>(`/admin/llm-providers/${id}`, data),
  delete: (id: number) => api.delete(`/admin/llm-providers/${id}`),
  setDefault: (id: number) => api.put(`/admin/llm-providers/${id}/default`),
  listModels: (name: string) => api.get<unknown, ModelInfo[]>(`/admin/llm-providers/${name}/models`),
};

export const uploadApi = {
  upload: (file: File, path?: string) => {
    const form = new FormData();
    form.append('file', file);
    if (path) form.append('path', path);
    return api.post('/admin/upload', form);
  },
};

export const searchApi = {
  search: (query: string, mode?: string) => api.post('/admin/search', { query, mode }),
};

export const statsApi = {
  dashboard: () => api.get<unknown, DashboardStats>('/admin/stats'),
  pipelinePerformance: (days?: number) => api.get('/admin/stats/pipeline-performance', { params: { days } }),
  llmPerformance: (days?: number) => api.get('/admin/stats/llm-performance', { params: { days } }),
  errors: (days?: number) => api.get('/admin/stats/errors', { params: { days } }),
};

export const graphApi = {
  getData: (limit?: number) => api.get<unknown, GraphData>('/admin/graph', { params: { limit } }),
};

export const apiKeyApi = {
  list: () => api.get<unknown, APIKey[]>('/admin/api-keys'),
  create: (data: APIKeyCreateRequest) => api.post<unknown, APIKey>('/admin/api-keys', data),
  update: (id: number, data: Partial<APIKeyCreateRequest & { active?: boolean }>) => api.put<unknown, APIKey>(`/admin/api-keys/${id}`, data),
  delete: (id: number) => api.delete(`/admin/api-keys/${id}`),
  usage: (id: number, days?: number) => api.get<unknown, APIKeyUsage[]>(`/admin/api-keys/${id}/usage`, { params: { days } }),
};

export const credentialApi = {
  list: (apiKeyId?: number) => api.get<unknown, Credential[]>('/admin/credentials', { params: apiKeyId ? { api_key_id: apiKeyId } : {} }),
  create: (data: CredentialCreateRequest) => api.post<unknown, Credential>('/admin/credentials', data),
  update: (id: number, data: Partial<CredentialCreateRequest & { active?: boolean }>) => api.put<unknown, Credential>(`/admin/credentials/${id}`, data),
  delete: (id: number) => api.delete(`/admin/credentials/${id}`),
};

export const tenantApi = {
  list: () => api.get<unknown, Tenant[]>('/admin/tenants'),
  create: (data: TenantCreateRequest) => api.post<unknown, Tenant>('/admin/tenants', data),
  update: (id: number, data: TenantUpdateRequest) => api.put<unknown, Tenant>(`/admin/tenants/${id}`, data),
  delete: (id: number) => api.delete(`/admin/tenants/${id}`),
};
