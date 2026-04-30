import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Copy, BarChart3 } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { Card, Button, Badge, Modal, Input, FormField } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { apiKeyApi } from '../../api';
import type { APIKey } from '../../api/types';

const emptyForm = { name: '', expires_at: '', daily_limit: '0', monthly_limit: '0', daily_tokens: '0', monthly_tokens: '0' };

export default function APIKeyPanel() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [usageId, setUsageId] = useState<number | null>(null);

  const { data } = useQuery({ queryKey: ['api-keys'], queryFn: () => apiKeyApi.list() });
  const keys: APIKey[] = (data as any)?.data ?? [];

  const { data: usageData } = useQuery({
    queryKey: ['api-key-usage', usageId],
    queryFn: () => apiKeyApi.usage(usageId!, 30),
    enabled: usageId !== null,
  });
  const usages = (usageData as any)?.data ?? [];

  const reload = () => queryClient.invalidateQueries({ queryKey: ['api-keys'] });

  const saveMut = useMutation({
    mutationFn: (d: typeof emptyForm) => {
      const payload = {
        name: d.name,
        expires_at: d.expires_at || undefined,
        daily_limit: parseInt(d.daily_limit) || 0,
        monthly_limit: parseInt(d.monthly_limit) || 0,
        daily_tokens: parseInt(d.daily_tokens) || 0,
        monthly_tokens: parseInt(d.monthly_tokens) || 0,
      };
      return editId ? apiKeyApi.update(editId, payload) : apiKeyApi.create(payload);
    },
    onSuccess: (res: any) => {
      reload();
      setShowForm(false);
      setEditId(null);
      if (!editId && res?.data?.key) {
        navigator.clipboard.writeText(res.data.key).then(() => toast.success('已创建，Key 已复制到剪贴板'));
      } else {
        toast.success('已保存');
      }
    },
    onError: () => toast.error('保存失败'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => apiKeyApi.delete(id),
    onSuccess: () => { reload(); setDeleteId(null); toast.success('已删除'); },
    onError: () => { setDeleteId(null); toast.error('删除失败'); },
  });

  const toggleMut = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) => apiKeyApi.update(id, { active }),
    onSuccess: () => { reload(); toast.success('已更新'); },
    onError: () => toast.error('操作失败'),
  });

  const openEdit = (k: APIKey) => {
    setEditId(k.id);
    setForm({
      name: k.name,
      expires_at: k.expires_at ? k.expires_at.slice(0, 10) : '',
      daily_limit: k.daily_limit.toString(),
      monthly_limit: k.monthly_limit.toString(),
      daily_tokens: k.daily_tokens.toString(),
      monthly_tokens: k.monthly_tokens.toString(),
    });
    setShowForm(true);
  };

  const openCreate = () => { setEditId(null); setForm(emptyForm); setShowForm(true); };

  return (
    <div>
      <div className="flex justify-end mb-4">
        <Button onClick={openCreate}><Plus className="w-4 h-4" /> 创建 API Key</Button>
      </div>

      <div className="space-y-3">
        {keys.map(k => (
          <Card key={k.id} className="flex items-center gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700">{k.name}</span>
                <Badge variant={k.active ? 'success' : 'default'}>{k.active ? '启用' : '停用'}</Badge>
              </div>
              <div className="flex items-center gap-3 mt-1 text-xs text-gray-400">
                <span className="font-mono">{k.prefix}••••</span>
                {k.expires_at && <span>过期: {k.expires_at.slice(0, 10)}</span>}
                <span>日限: {k.daily_limit || '∞'}</span>
                <span>月限: {k.monthly_limit || '∞'}</span>
              </div>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              <Button variant="ghost" size="sm" onClick={() => setUsageId(usageId === k.id ? null : k.id)} title="用量"><BarChart3 className="w-4 h-4" /></Button>
              <Button variant="ghost" size="sm" onClick={() => { navigator.clipboard.writeText(k.prefix + '••••'); toast.info('前缀已复制'); }} title="复制前缀"><Copy className="w-4 h-4" /></Button>
              <Button variant="ghost" size="sm" onClick={() => toggleMut.mutate({ id: k.id, active: !k.active })}>{k.active ? '停用' : '启用'}</Button>
              <Button variant="ghost" size="sm" onClick={() => openEdit(k)}>编辑</Button>
              <button onClick={() => setDeleteId(k.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </Card>
        ))}
        {keys.length === 0 && <Card><p className="text-sm text-gray-400 text-center py-4">暂无 API Key</p></Card>}
      </div>

      {usageId !== null && usages.length > 0 && (
        <Card className="mt-4">
          <h4 className="text-sm font-medium text-gray-600 mb-3">最近 30 天用量</h4>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={usages}>
              <XAxis dataKey="date" tick={{ fontSize: 11 }} tickFormatter={(v: string) => v.slice(5)} />
              <YAxis tick={{ fontSize: 11 }} allowDecimals={false} />
              <Tooltip />
              <Bar dataKey="requests" name="请求数" fill="#7b93f8" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </Card>
      )}

      <Modal
        open={showForm}
        onClose={() => setShowForm(false)}
        title={editId ? '编辑 API Key' : '创建 API Key'}
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
            <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
          </>
        }
      >
        <div className="space-y-3">
          <FormField label="名称">
            <Input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="API Key 名称" />
          </FormField>
          <FormField label="过期时间" hint="留空则永不过期">
            <Input type="date" value={form.expires_at} onChange={e => setForm({ ...form, expires_at: e.target.value })} />
          </FormField>
          <div className="grid grid-cols-2 gap-3">
            <FormField label="每日请求限制" hint="0=无限">
              <Input type="number" min="0" value={form.daily_limit} onChange={e => setForm({ ...form, daily_limit: e.target.value })} className="font-mono" />
            </FormField>
            <FormField label="每月请求限制">
              <Input type="number" min="0" value={form.monthly_limit} onChange={e => setForm({ ...form, monthly_limit: e.target.value })} className="font-mono" />
            </FormField>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <FormField label="每日 Token 限制">
              <Input type="number" min="0" value={form.daily_tokens} onChange={e => setForm({ ...form, daily_tokens: e.target.value })} className="font-mono" />
            </FormField>
            <FormField label="每月 Token 限制">
              <Input type="number" min="0" value={form.monthly_tokens} onChange={e => setForm({ ...form, monthly_tokens: e.target.value })} className="font-mono" />
            </FormField>
          </div>
        </div>
      </Modal>

      <ConfirmDialog open={deleteId !== null} message="确定删除该 API Key？关联的凭证将无法使用。" loading={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
    </div>
  );
}
