import axios from 'axios';

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
  login: (data: { username: string; password: string }) => api.post('/admin/auth/login', data),
};

export const pipelineApi = {
  list: () => api.get('/admin/pipelines'),
  get: (id: number) => api.get(`/admin/pipelines/${id}`),
  create: (data: any) => api.post('/admin/pipelines', data),
  update: (id: number, data: any) => api.put(`/admin/pipelines/${id}`, data),
  delete: (id: number) => api.delete(`/admin/pipelines/${id}`),
  createStep: (id: number, data: any) => api.post(`/admin/pipelines/${id}/steps`, data),
  updateStep: (id: number, stepId: number, data: any) => api.put(`/admin/pipelines/${id}/steps/${stepId}`, data),
  deleteStep: (id: number, stepId: number) => api.delete(`/admin/pipelines/${id}/steps/${stepId}`),
  reorderSteps: (id: number, stepIds: number[]) => api.put(`/admin/pipelines/${id}/steps/reorder`, { step_ids: stepIds }),
};

export const promptApi = {
  list: () => api.get('/admin/prompts'),
  get: (id: number) => api.get(`/admin/prompts/${id}`),
  create: (data: any) => api.post('/admin/prompts', data),
  update: (id: number, data: any) => api.put(`/admin/prompts/${id}`, data),
  delete: (id: number) => api.delete(`/admin/prompts/${id}`),
};

export const jobApi = {
  list: (params?: any) => api.get('/admin/jobs', { params }),
  submit: (data: any) => api.post('/admin/jobs', data),
  status: (uuid: string) => api.get(`/admin/jobs/${uuid}`),
  recommend: (scene: string) => api.get('/admin/jobs/recommend', { params: { scene } }),
};

export const reviewApi = {
  list: (params?: any) => api.get('/admin/reviews', { params }),
  approve: (id: number) => api.put(`/admin/reviews/${id}/approve`),
  reject: (id: number) => api.put(`/admin/reviews/${id}/reject`),
  modify: (id: number, data: any) => api.put(`/admin/reviews/${id}/modify`, data),
};

export const entityApi = {
  list: (params?: any) => api.get('/admin/entities', { params }),
  get: (id: number) => api.get(`/admin/entities/${id}`),
  getRelations: (id: number) => api.get(`/admin/entities/${id}/relations`),
  search: (keyword: string) => api.get('/admin/entities', { params: { keyword, page_size: 50 } }),
};

export const llmProviderApi = {
  list: () => api.get('/admin/llm-providers'),
  create: (data: any) => api.post('/admin/llm-providers', data),
  update: (id: number, data: any) => api.put(`/admin/llm-providers/${id}`, data),
  delete: (id: number) => api.delete(`/admin/llm-providers/${id}`),
  setDefault: (id: number) => api.put(`/admin/llm-providers/${id}/default`),
  listModels: (name: string) => api.get(`/admin/llm-providers/${name}/models`),
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
  search: (query: string) => api.post('/admin/search', { query }),
};
