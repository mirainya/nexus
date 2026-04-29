import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Plus, Trash2, Edit3, MessageSquareText } from 'lucide-react';
import { PageHeader, Card, Button, EmptyState, Loading } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { promptApi } from '../../api';

export default function PromptsPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<any>(null);
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null);
  const [form, setForm] = useState({ name: '', description: '', content: '', variables: '{}' });

  const { data, isLoading } = useQuery({ queryKey: ['prompts'], queryFn: () => promptApi.list() });
  const prompts = (data as any)?.data ?? [];

  const createMut = useMutation({
    mutationFn: (d: any) => editing ? promptApi.update(editing.id, d) : promptApi.create(d),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['prompts'] });
      closeForm();
      toast.success(editing ? '已更新' : '创建成功');
    },
    onError: () => toast.error(editing ? '更新失败' : '创建失败'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => promptApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['prompts'] });
      setDeleteTarget(null);
      toast.success('已删除');
    },
    onError: () => { setDeleteTarget(null); toast.error('删除失败'); },
  });

  const closeForm = () => { setShowForm(false); setEditing(null); setForm({ name: '', description: '', content: '', variables: '{}' }); };

  const openEdit = (p: any) => {
    setEditing(p);
    setForm({ name: p.name, description: p.description, content: p.content, variables: JSON.stringify(p.variables ?? {}, null, 2) });
    setShowForm(true);
  };

  if (isLoading) return <Loading />;

  return (
    <div>
      <PageHeader
        title="提示词模板"
        description="管理流水线步骤中使用的提示词"
        action={<Button onClick={() => { setEditing(null); setShowForm(true); }}><Plus className="w-4 h-4" /> 新建提示词</Button>}
      />

      {showForm && (
        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={closeForm}>
          <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-2xl shadow-xl max-h-[80vh] overflow-y-auto animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-800 mb-4">{editing ? '编辑' : '新建'}提示词模板</h3>
            <div className="space-y-3">
              <input
                placeholder="名称"
                value={form.name}
                onChange={e => setForm({ ...form, name: e.target.value })}
                className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
              />
              <input
                placeholder="描述"
                value={form.description}
                onChange={e => setForm({ ...form, description: e.target.value })}
                className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
              />
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">内容</label>
                <textarea
                  value={form.content}
                  onChange={e => setForm({ ...form, content: e.target.value })}
                  className="w-full px-4 py-3 rounded-xl border border-border-soft bg-surface text-sm font-mono focus:outline-none focus:border-nexus-300 h-48 resize-none"
                  placeholder="输入提示词模板... 使用 {{variable}} 作为变量占位符"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">变量 (JSON)</label>
                <textarea
                  value={form.variables}
                  onChange={e => setForm({ ...form, variables: e.target.value })}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm font-mono focus:outline-none focus:border-nexus-300 h-20 resize-none"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <Button variant="secondary" onClick={closeForm}>取消</Button>
              <Button
                loading={createMut.isPending}
                onClick={() => {
                  let vars = {};
                  try { vars = JSON.parse(form.variables); } catch {}
                  createMut.mutate({ ...form, variables: vars });
                }}
                disabled={!form.name || !form.content}
              >
                {editing ? '更新' : '创建'}
              </Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        title="删除提示词"
        message="确定要删除这个提示词模板吗？使用该模板的步骤将受到影响。"
        confirmText="删除"
        loading={deleteMut.isPending}
        onConfirm={() => deleteTarget && deleteMut.mutate(deleteTarget)}
        onCancel={() => setDeleteTarget(null)}
      />

      {prompts.length === 0 ? (
        <EmptyState message="暂无提示词模板" />
      ) : (
        <div className="space-y-3">
          {prompts.map((p: any) => (
            <Card key={p.id}>
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-sakura-100 to-lavender-100 flex items-center justify-center">
                    <MessageSquareText className="w-4 h-4 text-sakura-500" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-gray-800">{p.name}</h4>
                    <p className="text-xs text-gray-400 mt-0.5">{p.description} · v{p.version}</p>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  <button onClick={() => openEdit(p)} className="p-1.5 rounded-lg hover:bg-nexus-50 text-gray-300 hover:text-nexus-500 transition-all">
                    <Edit3 className="w-3.5 h-3.5" />
                  </button>
                  <button onClick={() => setDeleteTarget(p.id)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all">
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
              <details className="mt-3">
                <summary className="text-xs text-gray-400 cursor-pointer hover:text-gray-600">查看提示词内容</summary>
                <pre className="mt-2 text-xs text-gray-500 bg-surface rounded-lg p-3 overflow-auto max-h-64 whitespace-pre-wrap">{p.content}</pre>
              </details>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
