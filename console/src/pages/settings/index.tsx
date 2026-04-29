import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Star, Eye, EyeOff } from 'lucide-react';
import { PageHeader, Card, Button, Badge } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { llmProviderApi } from '../../api';
import APIKeyPanel from './APIKeyPanel';
import CredentialPanel from './CredentialPanel';

interface LLMProvider {
  id: number;
  name: string;
  display_name: string;
  base_url: string;
  api_key?: string;
  default_model: string;
  input_price: number;
  output_price: number;
  max_concurrency: number;
  active: boolean;
  is_default: boolean;
}

const tabs = [
  { key: 'llm', label: 'LLM 服务商' },
  { key: 'apikeys', label: 'API Key' },
  { key: 'credentials', label: '外部凭证' },
] as const;

type TabKey = typeof tabs[number]['key'];

const emptyForm = { name: '', display_name: '', base_url: '', api_key: '', default_model: '', input_price: '', output_price: '', max_concurrency: '10', active: true };

export default function SettingsPage() {
  const [tab, setTab] = useState<TabKey>('llm');
  const queryClient = useQueryClient();
  const toast = useToast();
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [showKeys, setShowKeys] = useState<Record<number, boolean>>({});

  const { data } = useQuery({ queryKey: ['llm-providers'], queryFn: () => llmProviderApi.list() });
  const providers: LLMProvider[] = (data as any)?.data ?? [];

  const reload = () => queryClient.invalidateQueries({ queryKey: ['llm-providers'] });

  const saveMut = useMutation({
    mutationFn: (d: any) => {
      const payload = { ...d, input_price: parseFloat(d.input_price) || 0, output_price: parseFloat(d.output_price) || 0, max_concurrency: parseInt(d.max_concurrency) || 10 };
      return editId ? llmProviderApi.update(editId, payload) : llmProviderApi.create(payload);
    },
    onSuccess: () => { reload(); setShowForm(false); setEditId(null); toast.success('已保存'); },
    onError: () => toast.error('保存失败'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => llmProviderApi.delete(id),
    onSuccess: () => { reload(); setDeleteId(null); toast.success('已删除'); },
    onError: () => { setDeleteId(null); toast.error('删除失败'); },
  });

  const defaultMut = useMutation({
    mutationFn: (id: number) => llmProviderApi.setDefault(id),
    onSuccess: () => { reload(); toast.success('已设为默认'); },
    onError: () => toast.error('操作失败'),
  });

  const openEdit = (p: LLMProvider) => {
    setEditId(p.id);
    setForm({ name: p.name, display_name: p.display_name, base_url: p.base_url, api_key: '', default_model: p.default_model, input_price: p.input_price?.toString() || '', output_price: p.output_price?.toString() || '', max_concurrency: p.max_concurrency?.toString() || '10', active: p.active });
    setShowForm(true);
  };

  const openCreate = () => {
    setEditId(null);
    setForm(emptyForm);
    setShowForm(true);
  };

  return (
    <div>
      <PageHeader title="设置" description="系统配置与开放平台管理" />

      <div className="flex gap-1 mb-6 bg-surface-hover rounded-xl p-1">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-all ${tab === t.key ? 'bg-white text-gray-800 shadow-sm' : 'text-gray-500 hover:text-gray-700'}`}>
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'llm' && (
        <>
          <div className="flex justify-end mb-4">
            <Button onClick={openCreate}><Plus className="w-4 h-4" /> 添加服务商</Button>
          </div>
          <div className="space-y-3">
            {providers.map(p => (
              <Card key={p.id} className="flex items-center gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-gray-700">{p.display_name || p.name}</span>
                    <Badge variant={p.active ? 'success' : 'default'}>{p.active ? '启用' : '禁用'}</Badge>
                    {p.is_default && <Badge variant="info">默认</Badge>}
                  </div>
                  <div className="flex items-center gap-4 mt-1.5">
                    <span className="text-xs text-gray-400">模型: {p.default_model}</span>
                    {(p.input_price > 0 || p.output_price > 0) && <span className="text-xs text-gray-400">价格: ${p.input_price}/{p.output_price} /1M tokens</span>}
                    <span className="text-xs text-gray-400 truncate max-w-xs">URL: {p.base_url}</span>
                    <button onClick={() => setShowKeys(prev => ({ ...prev, [p.id]: !prev[p.id] }))} className="text-xs text-gray-300 hover:text-gray-500 flex items-center gap-1">
                      {showKeys[p.id] ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                      {showKeys[p.id] ? (p.api_key || '••••••') : 'Key'}
                    </button>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  {!p.is_default && p.active && (
                    <button onClick={() => defaultMut.mutate(p.id)} className="p-1.5 rounded-lg hover:bg-amber-50 text-gray-300 hover:text-amber-500 transition-all" title="设为默认">
                      <Star className="w-4 h-4" />
                    </button>
                  )}
                  <Button variant="ghost" size="sm" onClick={() => openEdit(p)}>编辑</Button>
                  <button onClick={() => setDeleteId(p.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </Card>
            ))}
            {providers.length === 0 && (
              <Card><p className="text-sm text-gray-400 text-center py-4">暂无服务商配置</p></Card>
            )}
          </div>

          {showForm && (
            <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={() => setShowForm(false)}>
              <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-md shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
                <h3 className="text-sm font-semibold text-gray-800 mb-4">{editId ? '编辑服务商' : '添加服务商'}</h3>
                <div className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-500 mb-1.5">标识名</label>
                      <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="doubao" disabled={!!editId}
                        className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 disabled:opacity-50" />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-500 mb-1.5">显示名称</label>
                      <input value={form.display_name} onChange={e => setForm({ ...form, display_name: e.target.value })} placeholder="豆包"
                        className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300" />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500 mb-1.5">Base URL</label>
                    <input value={form.base_url} onChange={e => setForm({ ...form, base_url: e.target.value })} placeholder="https://api.openai.com/v1"
                      className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500 mb-1.5">API Key {editId && <span className="text-gray-300">(留空则不修改)</span>}</label>
                    <input type="password" value={form.api_key} onChange={e => setForm({ ...form, api_key: e.target.value })} placeholder="sk-..."
                      className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 font-mono" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500 mb-1.5">默认模型</label>
                    <input value={form.default_model} onChange={e => setForm({ ...form, default_model: e.target.value })} placeholder="gpt-4o"
                      className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300" />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-500 mb-1.5">输入价格 ($/1M tokens)</label>
                      <input type="number" step="0.01" value={form.input_price} onChange={e => setForm({ ...form, input_price: e.target.value })} placeholder="0.00"
                        className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 font-mono" />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-500 mb-1.5">输出价格 ($/1M tokens)</label>
                      <input type="number" step="0.01" value={form.output_price} onChange={e => setForm({ ...form, output_price: e.target.value })} placeholder="0.00"
                        className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 font-mono" />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500 mb-1.5">最大并发数</label>
                    <input type="number" min="1" max="100" value={form.max_concurrency} onChange={e => setForm({ ...form, max_concurrency: e.target.value })} placeholder="10"
                      className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 font-mono" />
                  </div>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" checked={form.active} onChange={e => setForm({ ...form, active: e.target.checked })}
                      className="w-4 h-4 rounded border-gray-300 text-nexus-500 focus:ring-nexus-300" />
                    <span className="text-sm text-gray-600">启用</span>
                  </label>
                </div>
                <div className="flex justify-end gap-2 mt-5">
                  <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
                  <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
                </div>
              </div>
            </div>
          )}

          <ConfirmDialog open={deleteId !== null} message="确定删除该服务商配置？" loading={deleteMut.isPending}
            onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
        </>
      )}

      {tab === 'apikeys' && <APIKeyPanel />}
      {tab === 'credentials' && <CredentialPanel />}
    </div>
  );
}
