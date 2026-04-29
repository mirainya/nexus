import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2 } from 'lucide-react';
import { Card, Button, Badge } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { credentialApi, apiKeyApi } from '../../api';
import type { Credential, APIKey } from '../../api/types';

const emptyForm = { api_key_id: '', name: '', provider_type: 'openai', base_url: '', api_key: '', default_model: '' };

export default function CredentialPanel() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [filterKeyId, setFilterKeyId] = useState<number | undefined>(undefined);

  const { data } = useQuery({ queryKey: ['credentials', filterKeyId], queryFn: () => credentialApi.list(filterKeyId) });
  const creds: Credential[] = (data as any)?.data ?? [];

  const { data: keysData } = useQuery({ queryKey: ['api-keys'], queryFn: () => apiKeyApi.list() });
  const apiKeys: APIKey[] = (keysData as any)?.data ?? [];

  const reload = () => queryClient.invalidateQueries({ queryKey: ['credentials'] });

  const saveMut = useMutation({
    mutationFn: (d: typeof emptyForm) => {
      const payload = {
        api_key_id: parseInt(d.api_key_id),
        name: d.name,
        provider_type: d.provider_type,
        base_url: d.base_url,
        api_key: d.api_key,
        default_model: d.default_model,
      };
      return editId ? credentialApi.update(editId, payload) : credentialApi.create(payload);
    },
    onSuccess: () => { reload(); setShowForm(false); setEditId(null); toast.success('已保存'); },
    onError: () => toast.error('保存失败'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => credentialApi.delete(id),
    onSuccess: () => { reload(); setDeleteId(null); toast.success('已删除'); },
    onError: () => { setDeleteId(null); toast.error('删除失败'); },
  });

  const toggleMut = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) => credentialApi.update(id, { active }),
    onSuccess: () => { reload(); toast.success('已更新'); },
    onError: () => toast.error('操作失败'),
  });

  const openEdit = (c: Credential) => {
    setEditId(c.id);
    setForm({
      api_key_id: c.api_key_id.toString(),
      name: c.name,
      provider_type: c.provider_type,
      base_url: c.base_url,
      api_key: '',
      default_model: c.default_model,
    });
    setShowForm(true);
  };

  const openCreate = () => { setEditId(null); setForm(emptyForm); setShowForm(true); };

  const getKeyName = (id: number) => apiKeys.find(k => k.id === id)?.name || `#${id}`;

  const inputCls = 'w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300';
  const labelCls = 'block text-xs font-medium text-gray-500 mb-1.5';

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <select value={filterKeyId ?? ''} onChange={e => setFilterKeyId(e.target.value ? parseInt(e.target.value) : undefined)}
            className="px-3 py-1.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300">
            <option value="">全部 API Key</option>
            {apiKeys.map(k => <option key={k.id} value={k.id}>{k.name}</option>)}
          </select>
        </div>
        <Button onClick={openCreate}><Plus className="w-4 h-4" /> 添加凭证</Button>
      </div>

      <div className="space-y-3">
        {creds.map(c => (
          <Card key={c.id} className="flex items-center gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700">{c.name}</span>
                <Badge variant={c.active ? 'success' : 'default'}>{c.active ? '启用' : '禁用'}</Badge>
                <Badge variant="info">{c.provider_type}</Badge>
              </div>
              <div className="flex items-center gap-4 mt-1.5">
                <span className="text-xs text-gray-400">API Key: {getKeyName(c.api_key_id)}</span>
                <span className="text-xs text-gray-400">模型: {c.default_model || '-'}</span>
                <span className="text-xs text-gray-400 truncate max-w-xs">URL: {c.base_url}</span>
                <code className="text-xs text-gray-400 font-mono">{c.api_key}</code>
              </div>
            </div>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" onClick={() => toggleMut.mutate({ id: c.id, active: !c.active })}>{c.active ? '禁用' : '启用'}</Button>
              <Button variant="ghost" size="sm" onClick={() => openEdit(c)}>编辑</Button>
              <button onClick={() => setDeleteId(c.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </Card>
        ))}
        {creds.length === 0 && <Card><p className="text-sm text-gray-400 text-center py-4">暂无凭证</p></Card>}
      </div>

      {showForm && (
        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={() => setShowForm(false)}>
          <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-md shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-semibold text-gray-800 mb-4">{editId ? '编辑凭证' : '添加凭证'}</h3>
            <div className="space-y-3">
              <div>
                <label className={labelCls}>所属 API Key</label>
                <select value={form.api_key_id} onChange={e => setForm({ ...form, api_key_id: e.target.value })}
                  disabled={!!editId} className={`${inputCls} disabled:opacity-50`}>
                  <option value="">请选择</option>
                  {apiKeys.filter(k => k.active).map(k => <option key={k.id} value={k.id}>{k.name}</option>)}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className={labelCls}>凭证名称</label>
                  <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="My OpenAI" className={inputCls} />
                </div>
                <div>
                  <label className={labelCls}>服务商类型</label>
                  <select value={form.provider_type} onChange={e => setForm({ ...form, provider_type: e.target.value })} className={inputCls}>
                    <option value="openai">OpenAI</option>
                    <option value="anthropic">Anthropic</option>
                    <option value="doubao">豆包</option>
                  </select>
                </div>
              </div>
              <div>
                <label className={labelCls}>Base URL</label>
                <input value={form.base_url} onChange={e => setForm({ ...form, base_url: e.target.value })} placeholder="https://api.openai.com/v1" className={inputCls} />
              </div>
              <div>
                <label className={labelCls}>API Key {editId && <span className="text-gray-300">(留空则不修改)</span>}</label>
                <input type="password" value={form.api_key} onChange={e => setForm({ ...form, api_key: e.target.value })} placeholder="sk-..." className={`${inputCls} font-mono`} />
              </div>
              <div>
                <label className={labelCls}>默认模型</label>
                <input value={form.default_model} onChange={e => setForm({ ...form, default_model: e.target.value })} placeholder="gpt-4o" className={inputCls} />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-5">
              <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
              <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog open={deleteId !== null} message="确定删除该凭证？" loading={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
    </div>
  );
}
