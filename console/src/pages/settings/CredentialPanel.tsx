import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2 } from 'lucide-react';
import { Card, Button, Badge, Modal, Input, FormField, Select } from '../../components/UI';
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
    setForm({ api_key_id: c.api_key_id.toString(), name: c.name, provider_type: c.provider_type, base_url: c.base_url, api_key: '', default_model: c.default_model });
    setShowForm(true);
  };

  const openCreate = () => { setEditId(null); setForm(emptyForm); setShowForm(true); };
  const getKeyName = (id: number) => apiKeys.find(k => k.id === id)?.name || `#${id}`;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-500">按 API Key 筛选:</span>
          <Select value={filterKeyId ?? ''} onChange={e => setFilterKeyId(e.target.value ? parseInt(e.target.value) : undefined)} className="!w-auto !py-1.5 !text-xs">
            <option value="">全部</option>
            {apiKeys.map(k => <option key={k.id} value={k.id}>{k.name}</option>)}
          </Select>
        </div>
        <Button onClick={openCreate}><Plus className="w-4 h-4" /> 添加凭证</Button>
      </div>

      <div className="space-y-3">
        {creds.map(c => (
          <Card key={c.id} className="flex items-center gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700">{c.name}</span>
                <Badge variant={c.active ? 'success' : 'default'}>{c.active ? '启用' : '停用'}</Badge>
                <Badge>{c.provider_type}</Badge>
              </div>
              <div className="flex items-center gap-3 mt-1 text-xs text-gray-400">
                <span>API Key: {getKeyName(c.api_key_id)}</span>
                <span>模型: {c.default_model}</span>
              </div>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              <Button variant="ghost" size="sm" onClick={() => toggleMut.mutate({ id: c.id, active: !c.active })}>{c.active ? '停用' : '启用'}</Button>
              <Button variant="ghost" size="sm" onClick={() => openEdit(c)}>编辑</Button>
              <button onClick={() => setDeleteId(c.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </Card>
        ))}
        {creds.length === 0 && <Card><p className="text-sm text-gray-400 text-center py-4">暂无凭证</p></Card>}
      </div>

      <Modal
        open={showForm}
        onClose={() => setShowForm(false)}
        title={editId ? '编辑凭证' : '添加凭证'}
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
            <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
          </>
        }
      >
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <FormField label="关联 API Key">
              <Select value={form.api_key_id} onChange={e => setForm({ ...form, api_key_id: e.target.value })}>
                <option value="">选择 API Key</option>
                {apiKeys.map(k => <option key={k.id} value={k.id}>{k.name}</option>)}
              </Select>
            </FormField>
            <FormField label="凭证名称">
              <Input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="My OpenAI" />
            </FormField>
          </div>
          <FormField label="服务商类型">
            <Select value={form.provider_type} onChange={e => setForm({ ...form, provider_type: e.target.value })}>
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic</option>
              <option value="doubao">豆包</option>
            </Select>
          </FormField>
          <FormField label="Base URL">
            <Input value={form.base_url} onChange={e => setForm({ ...form, base_url: e.target.value })} placeholder="https://api.openai.com/v1" />
          </FormField>
          <FormField label="API Key" hint={editId ? '留空则不修改' : undefined}>
            <Input type="password" value={form.api_key} onChange={e => setForm({ ...form, api_key: e.target.value })} placeholder="sk-..." className="font-mono" />
          </FormField>
          <FormField label="默认模型">
            <Input value={form.default_model} onChange={e => setForm({ ...form, default_model: e.target.value })} placeholder="gpt-4o" />
          </FormField>
        </div>
      </Modal>

      <ConfirmDialog open={deleteId !== null} message="确定删除该凭证？" loading={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
    </div>
  );
}
