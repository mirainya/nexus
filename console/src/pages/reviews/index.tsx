import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Check, X, ChevronDown, ChevronRight, Pencil } from 'lucide-react';
import { PageHeader, Card, StatusBadge, Button, EmptyState, Loading, FilterTabs, Pagination, Input } from '../../components/UI';
import { useToast } from '../../components/Toast';
import { reviewApi } from '../../api';

interface ReviewEntity {
  entity_id: number;
  type: string;
  name: string;
  confidence: number;
  attributes?: Record<string, unknown>;
  aliases?: string[];
}

const statusFilters = [
  { key: 'pending', label: '待审核' },
  { key: 'approved', label: '已通过' },
  { key: 'rejected', label: '已拒绝' },
  { key: 'modified', label: '已修改' },
  { key: '', label: '全部' },
];

export default function ReviewsPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState('pending');
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editEntities, setEditEntities] = useState<ReviewEntity[]>([]);

  const { data, isLoading } = useQuery({
    queryKey: ['reviews', page, status],
    queryFn: () => reviewApi.list({ page, page_size: 20, status: status || undefined }),
  });

  const reviews = (data as any)?.data?.list ?? [];
  const total = (data as any)?.data?.total ?? 0;
  const reload = () => queryClient.invalidateQueries({ queryKey: ['reviews'] });

  const approveMut = useMutation({
    mutationFn: (id: number) => reviewApi.approve(id),
    onSuccess: () => { reload(); toast.success('已通过，实体已确认'); },
    onError: () => toast.error('操作失败'),
  });

  const rejectMut = useMutation({
    mutationFn: (id: number) => reviewApi.reject(id),
    onSuccess: () => { reload(); toast.success('已拒绝，实体已删除'); },
    onError: () => toast.error('操作失败'),
  });

  const modifyMut = useMutation({
    mutationFn: ({ id, entities }: { id: number; entities: ReviewEntity[] }) =>
      reviewApi.modify(id, { entities }),
    onSuccess: () => { reload(); setEditingId(null); toast.success('已修改并确认'); },
    onError: () => toast.error('操作失败'),
  });

  const parseEntities = (r: any): ReviewEntity[] => {
    try {
      const data = typeof r.original_data === 'string' ? JSON.parse(r.original_data) : r.original_data;
      if (Array.isArray(data)) return data;
      return [data];
    } catch {
      return [];
    }
  };

  const startEdit = (r: any) => {
    setEditingId(r.id);
    setEditEntities(parseEntities(r).map(e => ({ ...e })));
  };

  const updateEntity = (idx: number, field: string, value: string) => {
    setEditEntities(prev => prev.map((e, i) => i === idx ? { ...e, [field]: value } : e));
  };

  if (isLoading) return <Loading />;

  return (
    <div>
      <PageHeader title="审核" description={`共 ${total} 条审核`} />

      <FilterTabs
        items={statusFilters}
        value={status}
        onChange={v => { setStatus(v); setPage(1); }}
      />

      {reviews.length === 0 ? (
        <EmptyState message="暂无审核项" />
      ) : (
        <div className="space-y-3">
          {reviews.map((r: any) => {
            const entities = parseEntities(r);
            const isExpanded = expandedId === r.id;
            return (
              <Card key={r.id} className="!p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 cursor-pointer" onClick={() => setExpandedId(isExpanded ? null : r.id)}>
                    {isExpanded ? <ChevronDown className="w-4 h-4 text-gray-400" /> : <ChevronRight className="w-4 h-4 text-gray-400" />}
                    <span className="text-xs font-mono text-gray-400">#{r.id}</span>
                    <StatusBadge status={r.status} />
                    <span className="text-xs text-gray-500">{entities.length} 个实体</span>
                  </div>
                  {r.status === 'pending' && (
                    <div className="flex items-center gap-1">
                      <Button size="sm" variant="ghost" onClick={() => startEdit(r)} title="修改"><Pencil className="w-3.5 h-3.5" /></Button>
                      <Button size="sm" variant="ghost" onClick={() => rejectMut.mutate(r.id)} title="拒绝"><X className="w-4 h-4 text-red-400" /></Button>
                      <Button size="sm" onClick={() => approveMut.mutate(r.id)} title="通过"><Check className="w-4 h-4" /></Button>
                    </div>
                  )}
                </div>

                {isExpanded && editingId !== r.id && (
                  <div className="mt-3 space-y-2">
                    {entities.map((e, i) => (
                      <div key={i} className="flex items-center gap-3 px-3 py-2 bg-surface rounded-lg text-xs">
                        <span className="text-gray-500 w-16">{e.type}</span>
                        <span className="font-medium text-gray-700 flex-1">{e.name}</span>
                        <span className="text-gray-400">{(e.confidence * 100).toFixed(0)}%</span>
                      </div>
                    ))}
                  </div>
                )}

                {editingId === r.id && (
                  <div className="mt-3 space-y-3 border-t border-border-soft pt-3">
                    {editEntities.map((e, i) => (
                      <div key={i} className="px-3 py-3 bg-surface rounded-lg space-y-2">
                        <div className="grid grid-cols-3 gap-2">
                          <div>
                            <label className="text-[10px] text-gray-400">类型</label>
                            <Input value={e.type} onChange={ev => updateEntity(i, 'type', ev.target.value)} className="!py-1.5 !text-xs" />
                          </div>
                          <div>
                            <label className="text-[10px] text-gray-400">名称</label>
                            <Input value={e.name} onChange={ev => updateEntity(i, 'name', ev.target.value)} className="!py-1.5 !text-xs" />
                          </div>
                          <div>
                            <label className="text-[10px] text-gray-400">置信度</label>
                            <Input value={(e.confidence * 100).toFixed(0)} disabled className="!py-1.5 !text-xs opacity-50" />
                          </div>
                        </div>
                      </div>
                    ))}
                    <div className="flex justify-end gap-2">
                      <Button size="sm" variant="secondary" onClick={() => setEditingId(null)}>取消</Button>
                      <Button size="sm" loading={modifyMut.isPending} onClick={() => modifyMut.mutate({ id: r.id, entities: editEntities })}>保存修改</Button>
                    </div>
                  </div>
                )}
              </Card>
            );
          })}
        </div>
      )}

      <Pagination page={page} total={total} pageSize={20} onChange={setPage} />
    </div>
  );
}
