import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Trash2, Settings2 } from 'lucide-react';
import { PageHeader, Card, Button, Badge, EmptyState, Loading } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { pipelineApi } from '../../api';

export default function PipelinesPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const toast = useToast();
  const [showCreate, setShowCreate] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null);
  const [form, setForm] = useState({ name: '', description: '' });

  const { data, isLoading } = useQuery({ queryKey: ['pipelines'], queryFn: () => pipelineApi.list() });
  const pipelines = (data as any)?.data ?? [];

  const createMut = useMutation({
    mutationFn: (d: any) => pipelineApi.create(d),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipelines'] });
      setShowCreate(false);
      setForm({ name: '', description: '' });
      toast.success('流水线创建成功');
    },
    onError: () => toast.error('创建失败'),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => pipelineApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipelines'] });
      setDeleteTarget(null);
      toast.success('已删除');
    },
    onError: () => { setDeleteTarget(null); toast.error('删除失败'); },
  });

  if (isLoading) return <Loading />;

  return (
    <div>
      <PageHeader
        title="流水线"
        description="管理数据处理流水线"
        action={<Button onClick={() => setShowCreate(true)}><Plus className="w-4 h-4" /> 新建流水线</Button>}
      />

      {showCreate && (
        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
          <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-md shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-800 mb-4">新建流水线</h3>
            <div className="space-y-3">
              <input
                placeholder="流水线名称"
                value={form.name}
                onChange={e => setForm({ ...form, name: e.target.value })}
                className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 focus:ring-2 focus:ring-nexus-100"
              />
              <textarea
                placeholder="描述"
                value={form.description}
                onChange={e => setForm({ ...form, description: e.target.value })}
                className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 focus:ring-2 focus:ring-nexus-100 h-24 resize-none"
              />
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <Button variant="secondary" onClick={() => setShowCreate(false)}>取消</Button>
              <Button onClick={() => createMut.mutate({ ...form, active: true })} disabled={!form.name} loading={createMut.isPending}>创建</Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        title="删除流水线"
        message="确定要删除这条流水线吗？此操作不可撤销。"
        confirmText="删除"
        loading={deleteMut.isPending}
        onConfirm={() => deleteTarget && deleteMut.mutate(deleteTarget)}
        onCancel={() => setDeleteTarget(null)}
      />

      {pipelines.length === 0 ? (
        <EmptyState message="暂无流水线，创建第一个吧" />
      ) : (
        <div className="grid grid-cols-2 gap-4">
          {pipelines.map((p: any) => (
            <Card key={p.id} onClick={() => navigate(`/pipelines/${p.id}`)} className="group">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-nexus-100 to-lavender-100 flex items-center justify-center">
                    <Settings2 className="w-4 h-4 text-nexus-500" />
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-gray-800">{p.name}</h4>
                    <p className="text-xs text-gray-400 mt-0.5">{p.steps?.length ?? 0} 个步骤</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={p.active ? 'success' : 'default'}>{p.active ? '启用' : '停用'}</Badge>
                  <button
                    onClick={(e) => { e.stopPropagation(); setDeleteTarget(p.id); }}
                    className="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
              {p.description && <p className="text-xs text-gray-400 mt-3 line-clamp-2">{p.description}</p>}
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
