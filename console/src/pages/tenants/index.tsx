import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2 } from 'lucide-react';
import { PageHeader, Card, Button, Badge, Modal, Input, FormField } from '../../components/UI';
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
    setForm({ name: t.name, monthly_request_limit: t.monthly_request_limit.toString(), monthly_token_limit: t.monthly_token_limit.toString() });
    setShowForm(true);
  };

  const openCreate = () => { setEditId(null); setForm({ ...emptyForm }); setShowForm(true); };

  return (
    <div>
      <PageHeader
        title="租户管理"
        description="管理平台租户与数据隔离"
        action={<Button onClick={openCreate}><Plus className="w-4 h-4" /> 创建租户</Button>}
      />

      <div className="space-y-3">
        {tenants.map(t => (
          <Card key={t.id} className="flex items-center gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700">{t.name}</span>
                <Badge variant={t.active ? 'success' : 'default'}>{t.active ? '启用' : '禁用'}</Badge>
              </div>
              <div className="flex items-center gap-4 mt-1.5 text-xs text-gray-400">
                <span className="font-mono">{t.uuid}</span>
                <span>创建: {t.created_at?.slice(0, 10)}</span>
                <span>月请求限额: {t.monthly_request_limit || '无限'}</span>
                <span>月 Token 限额: {t.monthly_token_limit || '无限'}</span>
              </div>
            </div>
            <div className="flex items-center gap-1 shrink-0">
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

      <Modal
        open={showForm}
        onClose={() => setShowForm(false)}
        title={editId ? '编辑租户' : '创建租户'}
        maxWidth="max-w-sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowForm(false)}>取消</Button>
            <Button loading={saveMut.isPending} onClick={() => saveMut.mutate(form)}>保存</Button>
          </>
        }
      >
        <div className="space-y-3">
          <FormField label="名称">
            <Input value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="租户名称" />
          </FormField>
          {editId && (
            <div className="grid grid-cols-2 gap-3">
              <FormField label="月请求限额" hint="0=无限">
                <Input type="number" min="0" value={form.monthly_request_limit} onChange={e => setForm({ ...form, monthly_request_limit: e.target.value })} className="font-mono" />
              </FormField>
              <FormField label="月 Token 限额">
                <Input type="number" min="0" value={form.monthly_token_limit} onChange={e => setForm({ ...form, monthly_token_limit: e.target.value })} className="font-mono" />
              </FormField>
            </div>
          )}
        </div>
      </Modal>

      <ConfirmDialog open={deleteId !== null} message="确定删除该租户？关联的数据将无法访问。" loading={deleteMut.isPending}
        onConfirm={() => deleteId && deleteMut.mutate(deleteId)} onCancel={() => setDeleteId(null)} />
    </div>
  );
}
