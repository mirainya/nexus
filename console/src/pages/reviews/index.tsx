import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Check, X } from 'lucide-react';
import { PageHeader, Card, StatusBadge, Button, EmptyState, Loading } from '../../components/UI';
import { useToast } from '../../components/Toast';
import { reviewApi } from '../../api';

export default function ReviewsPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState('pending');

  const { data, isLoading } = useQuery({
    queryKey: ['reviews', page, status],
    queryFn: () => reviewApi.list({ page, page_size: 20, status: status || undefined }),
  });

  const reviews = (data as any)?.data?.list ?? [];
  const total = (data as any)?.data?.total ?? 0;

  const approveMut = useMutation({
    mutationFn: (id: number) => reviewApi.approve(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['reviews'] }); toast.success('已通过'); },
    onError: () => toast.error('操作失败'),
  });

  const rejectMut = useMutation({
    mutationFn: (id: number) => reviewApi.reject(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['reviews'] }); toast.success('已拒绝'); },
    onError: () => toast.error('操作失败'),
  });

  if (isLoading) return <Loading />;

  return (
    <div>
      <PageHeader title="审核" description={`共 ${total} 条审核`} />

      <div className="flex gap-2 mb-6">
        {['', 'pending', 'approved', 'rejected', 'modified'].map(s => (
          <button
            key={s}
            onClick={() => { setStatus(s); setPage(1); }}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
              status === s ? 'bg-nexus-50 text-nexus-600' : 'text-gray-400 hover:bg-surface-hover'
            }`}
          >
            {s === '' ? '全部' : s === 'pending' ? '待审核' : s === 'approved' ? '已通过' : s === 'rejected' ? '已拒绝' : '已修改'}
          </button>
        ))}
      </div>

      {reviews.length === 0 ? (
        <EmptyState message="暂无审核记录" />
      ) : (
        <div className="space-y-3">
          {reviews.map((r: any) => (
            <Card key={r.id} className="!p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-500">#{r.id}</span>
                  <StatusBadge status={r.status} />
                  {r.reviewer && <span className="text-xs text-gray-400">审核人: {r.reviewer}</span>}
                </div>
                {r.status === 'pending' && (
                  <div className="flex gap-1">
                    <Button size="sm" variant="ghost" onClick={() => approveMut.mutate(r.id)} loading={approveMut.isPending}>
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => rejectMut.mutate(r.id)} loading={rejectMut.isPending}>
                      <X className="w-3.5 h-3.5 text-red-400" />
                    </Button>
                  </div>
                )}
              </div>
              {r.original_data && (
                <pre className="mt-2 text-xs text-gray-500 bg-surface rounded-lg p-3 overflow-x-auto max-h-32">
                  {typeof r.original_data === 'string' ? JSON.stringify(JSON.parse(r.original_data), null, 2) : JSON.stringify(r.original_data, null, 2)}
                </pre>
              )}
            </Card>
          ))}
        </div>
      )}

      {total > 20 && (
        <div className="flex justify-center gap-2 mt-6">
          <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="px-3 py-1.5 rounded-lg text-xs text-gray-500 hover:bg-surface-hover disabled:opacity-30">上一页</button>
          <span className="px-3 py-1.5 text-xs text-gray-400">第 {page} 页</span>
          <button onClick={() => setPage(p => p + 1)} className="px-3 py-1.5 rounded-lg text-xs text-gray-500 hover:bg-surface-hover">下一页</button>
        </div>
      )}
    </div>
  );
}
