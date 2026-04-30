import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Copy, BarChart3 } from 'lucide-react';
import { Card, Button, Badge } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { apiKeyApi, tenantApi } from '../../api';
import type { APIKey, Tenant } from '../../api/types';

const emptyForm = { name: '', tenant_id: '', expires_at: '', daily_limit: '0', monthly_limit: '0', daily_tokens: '0', monthly_tokens: '0' };

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

  const { data: tenantData } = useQuery({ queryKey: ['tenants'], queryFn: () => tenantApi.list() });
  const tenants: Tenant[] = (tenantData as any)?.data ?? [];

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
        tenant_id: parseInt(d.tenant_id) || 0,
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
      tenant_id: k.tenant_id.toString(),
      expires_at: k.expires_at ? k.expires_at.slice(0, 10) : '',
      daily_limit: k.daily_limit.toString(),
      monthly_limit: k.monthly_limit.toString(),
      daily_tokens: k.daily_tokens.toString(),
      monthly_tokens: k.monthly_tokens.toString(),
    });
    setShowForm(true);
  };

  const openCreate = () => { setEditId(null); setForm(emptyForm); setShowForm(true); };

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key).then(() => toast.success('已复制'));
  };

  const inputCls = 'w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300';
  const labelCls = 'block text-xs font-medium text-gray-500 mb-1.5';

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
                <Badge variant={k.active ? 'success' : 'default'}>{k.active ? '启用' : '禁用'}</Badge>
                <span className="text-xs text-gray-400">租户: {tenants.find(t => t.id === k.tenant_id)?.name || k.tenant_id}</span>
                {k.expires_at && <span className="text-xs text-gray-400">过期: {k.expires_at.slice(0, 10)}</span>}
              </div>
              <div className="flex items-center gap-3 mt-1.5">
                <code className="text-xs text-gray-400 font-mono">{k.key}</code>
                <button onClick={() => copyKey(k.key)} className="text-gray-300 hover:text-gray-500"><Copy className="w-3 h-3" /></button>
              </div>
              <div className="flex items-center gap-4 mt-1">
                <span className="text-xs text-gray-400">日限: {k.daily_limit || '无限'} 次</span>
                <span className="text-xs text-gray-400">月限: {k.monthly_limit || '无限'} 次</span>
                <span className="text-xs text-gray-400">日 Token: {k.daily_tokens || '无限'}</span>
                <span className="text-xs text-gray-400">月 Token: {k.monthly_tokens || '无限'}</span>
              </div>
            </div>
            <div className="flex items-center gap-1">
              <button onClick={() => setUsageId(usageId === k.id ? null : k.id)} className={`p-1.5 rounded-lg hover:bg-nexus-50 transition-all ${usageId === k.id ? 'text-nexus-500' : 'text-gray-300 hover:text-nexus-500'}`} title="用量">
                <BarChart3 className="w-4 h-4" />
              </button>
              <Button variant="ghost" size="sm" onClick={() => toggleMut.mutate({ id: k.id, active: !k.active })}>{k.active ? '禁用' : '启用'}</Button>
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
          <h4 className="text-sm font-medium text-gray-700 mb-3">近 30 天用量</h4>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead><tr className="text-gray-400 border-b border-border-soft">
                <th className="text-left py-2 font-medium">日期</th>
                <th className="text-right py-2 font-medium">请求数</th>
                <th className="text-right py-2 font-medium">Tokens</th>
              </tr></thead>
              <tbody>
                {usages.map((u: any) => (
                  <tr key={u.date} className="border-b border-border-soft/50">
                    <td className="py-1.5 text-gray-600">{u.date}</td>
                    <td className="py-1.5 text-right text-gray-600">{u.requests.toLocaleString()}</td>
                    <td className="py-1.5 text-right text-gray-600">{u.tokens.toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {showForm && (
        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={() => setShowForm(false)}>
          <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-md shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-semibold text-gray-800 mb-4">{editId ? '编辑 API Key' : '创建 API Key'}</h3>
            <div className="space-y-3">
              <div>
                <label className={labelCls}>名称</label>
                <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="My Project" className={inputCls} />
              </div>
              <div>
                <label className={labelCls}>所属租户</label>
                <select value={form.tenant_id} onChange={e => setForm({ ...form, tenant_id: e.target.value })} className={inputCls} disabled={!!editId}>
                  <option value="">请选择租户</option>
                  {tenants.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                </select>
              </div>
              <div>
                <label className={labelCls}>过期时间 <span className="text-gray-300">(留空则永不过期)</span></label>
                <input type="date" value={form.expires_at} onChange={e => setForm({ ...form, expires_at: e.target.value })} className={inputCls} />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className={labelCls}>每日请求限制 <span className="text-gray-300">(0=无限)</span></label>
                  <input type="number" min="0" value={form.daily_limit} onChange={e => setForm({ ...form, daily_limit: e.target.value })} className={`${inputCls} font-mono`} />
                </div>
                <div>
                  <label className={labelCls}>每月请求限制</label>
                  <input type="number" min="0" value={form.monthly_limit} onChange={e => setForm({ ...form, monthly_limit: e.target.value })} className={`${inputCls} font-mono`} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className={labelCls}>每日 Token 限制</label>
                  <input type="number" min="0" value={form.daily_tokens} onChange={e => setForm({ ...form, daily_tokens: e.target.value })} className={`${inputCls} font-mono`} />
                </div>
                <div>
                  <label className={labelCls}>每月 Token 限制</label>
                  <input type="number" min="0" value={form.monthly_tokens} onChange={e => setForm({ ...form, monthly_tokens: e.target.value })} className={`${inputCls} font-mono`} />
                </div>
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-5">
              <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
              <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog open={deleteId !== null} message="确定删除该 API Key？关联的凭证将无法使用。" loading={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
    </div>
  );
}
