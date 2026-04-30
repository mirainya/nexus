import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Star, Eye, EyeOff } from 'lucide-react';
import { PageHeader, Card, Button, Badge, Tabs, Modal, Input, FormField } from '../../components/UI';
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
];

const emptyForm = { name: '', display_name: '', base_url: '', api_key: '', default_model: '', input_price: '', output_price: '', max_concurrency: '10', active: true };

export default function SettingsPage() {
  const [tab, setTab] = useState('llm');
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

  const openCreate = () => { setEditId(null); setForm(emptyForm); setShowForm(true); };

  return (
    <div>
      <PageHeader title="设置" description="管理 LLM 服务商、API Key 和外部凭证" />

      <Tabs items={tabs} value={tab} onChange={setTab} />

      {tab === 'llm' && (
        <>
          <div className="flex justify-end mb-4">
            <Button onClick={openCreate}><Plus className="w-4 h-4" /> 添加服务商</Button>
          </div>

          <div className="space-y-3">
            {providers.map(p => (
              <Card key={p.id} className="flex items-center gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-gray-700">{p.display_name || p.name}</span>
                    {p.is_default && <Badge variant="info">默认</Badge>}
                    <Badge variant={p.active ? 'success' : 'default'}>{p.active ? '启用' : '停用'}</Badge>
                  </div>
                  <div className="flex items-center gap-4 mt-1.5 text-xs text-gray-400">
                    <span className="font-mono">{p.name}</span>
                    <span>模型: {p.default_model}</span>
                    <span>并发: {p.max_concurrency}</span>
                    {p.api_key && (
                      <span className="flex items-center gap-1 font-mono">
                        Key: {showKeys[p.id] ? p.api_key : '••••••••'}
                        <button onClick={() => setShowKeys(s => ({ ...s, [p.id]: !s[p.id] }))} className="text-gray-300 hover:text-gray-500">
                          {showKeys[p.id] ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                        </button>
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  {!p.is_default && <Button variant="ghost" size="sm" onClick={() => defaultMut.mutate(p.id)} title="设为默认"><Star className="w-4 h-4" /></Button>}
                  <Button variant="ghost" size="sm" onClick={() => openEdit(p)}>编辑</Button>
                  <button onClick={() => setDeleteId(p.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </Card>
            ))}
            {providers.length === 0 && <Card><p className="text-sm text-gray-400 text-center py-4">暂无服务商</p></Card>}
          </div>

          <Modal
            open={showForm}
            onClose={() => setShowForm(false)}
            title={editId ? '编辑服务商' : '添加服务商'}
            maxWidth="max-w-lg"
            footer={
              <>
                <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
                <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
              </>
            }
          >
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <FormField label="标识名"><Input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="openai" /></FormField>
                <FormField label="显示名"><Input value={form.display_name} onChange={e => setForm({ ...form, display_name: e.target.value })} placeholder="OpenAI" /></FormField>
              </div>
              <FormField label="Base URL"><Input value={form.base_url} onChange={e => setForm({ ...form, base_url: e.target.value })} placeholder="https://api.openai.com/v1" /></FormField>
              <FormField label="API Key" hint={editId ? '留空则不修改' : undefined}>
                <Input type="password" value={form.api_key} onChange={e => setForm({ ...form, api_key: e.target.value })} placeholder="sk-..." className="font-mono" />
              </FormField>
              <FormField label="默认模型"><Input value={form.default_model} onChange={e => setForm({ ...form, default_model: e.target.value })} placeholder="gpt-4o" /></FormField>
              <div className="grid grid-cols-2 gap-3">
                <FormField label="输入价格 ($/1M tokens)"><Input type="number" step="0.01" value={form.input_price} onChange={e => setForm({ ...form, input_price: e.target.value })} placeholder="0.00" className="font-mono" /></FormField>
                <FormField label="输出价格 ($/1M tokens)"><Input type="number" step="0.01" value={form.output_price} onChange={e => setForm({ ...form, output_price: e.target.value })} placeholder="0.00" className="font-mono" /></FormField>
              </div>
              <FormField label="最大并发数"><Input type="number" min="1" max="100" value={form.max_concurrency} onChange={e => setForm({ ...form, max_concurrency: e.target.value })} className="font-mono" /></FormField>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={form.active} onChange={e => setForm({ ...form, active: e.target.checked })} className="w-4 h-4 rounded border-gray-300 text-nexus-500 focus:ring-nexus-300" />
                <span className="text-sm text-gray-600">启用</span>
              </label>
            </div>
          </Modal>

          <ConfirmDialog open={deleteId !== null} message="确定删除该服务商配置？" loading={deleteMut.isPending}
            onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
        </>
      )}

      {tab === 'apikeys' && <APIKeyPanel />}
      {tab === 'credentials' && <CredentialPanel />}
    </div>
  );
}
