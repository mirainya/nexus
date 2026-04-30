import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2 } from 'lucide-react';
import { PageHeader, Card, Button, Badge } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { tenantApi } from '../../api';
import type { Tenant } from '../../api/types';

const emptyForm = { name: '', monthly_request_limit: '0', monthly_token_limit: '0' };

export default function TenantsPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<number | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const { data } = useQuery({ queryKey: ['tenants'], queryFn: () => tenantApi.list() });
  const tenants: Tenant[] = (data as any)?.data ?? [];

  const reload = () => queryClient.invalidateQueries({ queryKey: ['tenants'] });

  const saveMut = useMutation({
    mutationFn: (d: typeof emptyForm) => {
      if (editId) {
        return tenantApi.update(editId, {
          name: d.name,
          monthly_request_limit: parseInt(d.monthly_request_limit) || 0,
          monthly_token_limit: parseInt(d.monthly_token_limit) || 0,
        });
      }
      return tenantApi.create({ name: d.name });
    },
    onSuccess: () => { reload(); setShowForm(false); setEditId(null); toast.success('已保存'); },
    onError: () => toast.error('保存失败'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => tenantApi.delete(id),
    onSuccess: () => { reload(); setDeleteId(null); toast.success('已删除'); },
    onError: () => { setDeleteId(null); toast.error('删除失败'); },
  });

  const toggleMut = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) => tenantApi.update(id, { active }),
    onSuccess: () => { reload(); toast.success('已更新'); },
    onError: () => toast.error('操作失败'),
  });

  const openEdit = (t: Tenant) => {
    setEditId(t.id);
    setForm({
      name: t.name,
      monthly_request_limit: t.monthly_request_limit.toString(),
      monthly_token_limit: t.monthly_token_limit.toString(),
    });
    setShowForm(true);
  };

  const openCreate = () => { setEditId(null); setForm({ ...emptyForm }); setShowForm(true); };

  const inputCls = 'w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300';

  return (
    <div>
      <PageHeader title="租户管理" description="管理平台租户与数据隔离" />

      <div className="flex justify-end mb-4">
        <Button onClick={openCreate}><Plus className="w-4 h-4" /> 创建租户</Button>
      </div>

      <div className="space-y-3">
        {tenants.map(t => (
          <Card key={t.id} className="flex items-center gap-4">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700">{t.name}</span>
                <Badge variant={t.active ? 'success' : 'default'}>{t.active ? '启用' : '禁用'}</Badge>
              </div>
              <div className="flex items-center gap-4 mt-1.5">
                <span className="text-xs text-gray-400 font-mono">{t.uuid}</span>
                <span className="text-xs text-gray-400">创建: {t.created_at?.slice(0, 10)}</span>
                <span className="text-xs text-gray-400">月请求限额: {t.monthly_request_limit || '无限'}</span>
                <span className="text-xs text-gray-400">月 Token 限额: {t.monthly_token_limit || '无限'}</span>
              </div>
            </div>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" onClick={() => toggleMut.mutate({ id: t.id, active: !t.active })}>{t.active ? '禁用' : '启用'}</Button>
              <Button variant="ghost" size="sm" onClick={() => openEdit(t)}>编辑</Button>
              <button onClick={() => setDeleteId(t.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          </Card>
        ))}
        {tenants.length === 0 && <Card><p className="text-sm text-gray-400 text-center py-4">暂无租户</p></Card>}
      </div>

      {showForm && (
        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={() => setShowForm(false)}>
          <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-sm shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-semibold text-gray-800 mb-4">{editId ? '编辑租户' : '创建租户'}</h3>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">名称</label>
                <input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="租户名称" className={inputCls} />
              </div>
              {editId && (
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-500 mb-1.5">月请求限额 <span className="text-gray-300">(0=无限)</span></label>
                    <input type="number" min="0" value={form.monthly_request_limit} onChange={e => setForm({ ...form, monthly_request_limit: e.target.value })} className={`${inputCls} font-mono`} />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-500 mb-1.5">月 Token 限额</label>
                    <input type="number" min="0" value={form.monthly_token_limit} onChange={e => setForm({ ...form, monthly_token_limit: e.target.value })} className={`${inputCls} font-mono`} />
                  </div>
                </div>
              )}
            </div>
            <div className="flex justify-end gap-2 mt-5">
              <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
              <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog open={deleteId !== null} message="确定删除该租户？关联的数据将无法访问。" loading={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
    </div>
  );
}
