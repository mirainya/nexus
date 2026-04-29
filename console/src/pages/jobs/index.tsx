import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { PageHeader, Card, StatusBadge, EmptyState, Loading } from '../../components/UI';
import { jobApi } from '../../api';

const stepStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '等待', color: 'text-gray-400' },
  running: { label: '执行中', color: 'text-nexus-600' },
  completed: { label: '完成', color: 'text-green-600' },
  failed: { label: '失败', color: 'text-red-500' },
  skipped: { label: '跳过', color: 'text-amber-500' },
};

function StepProgress({ job }: { job: any }) {
  const total = job.total_steps || 0;
  const current = job.current_step || 0;
  const logs: any[] = job.step_logs || [];

  if (total === 0) return null;

  const percent = job.status === 'completed' ? 100 : Math.round((current / total) * 100);

  return (
    <div className="mt-3 space-y-2">
      <div className="flex items-center justify-between text-xs">
        <span className="text-gray-500">进度 {current}/{total}</span>
        <span className="text-gray-400">{percent}%</span>
      </div>
      <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-500 ${
            job.status === 'failed' ? 'bg-red-400' : job.status === 'completed' ? 'bg-green-500' : 'bg-nexus-500'
          }`}
          style={{ width: `${percent}%` }}
        />
      </div>
      {logs.length > 0 && (
        <div className="space-y-1 mt-2">
          {logs.map((log: any, i: number) => {
            const info = stepStatusMap[log.status] || stepStatusMap.pending;
            const duration = log.started_at && log.finished_at
              ? ((new Date(log.finished_at).getTime() - new Date(log.started_at).getTime()) / 1000).toFixed(1)
              : null;
            return (
              <div key={i} className="flex items-center gap-2 text-xs py-1 px-2 rounded-md bg-surface">
                <span className="w-5 text-gray-300 text-right">{log.step_order + 1}</span>
                <span className="font-mono text-gray-600 flex-1">{log.processor_type}</span>
                <span className={info.color}>{info.label}</span>
                {duration && <span className="text-gray-400">{duration}s</span>}
                {log.tokens > 0 && <span className="text-gray-400">{log.tokens}t</span>}
                {log.error && <span className="text-red-400 truncate max-w-[200px]" title={log.error}>{log.error}</span>}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default function JobsPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState('');

  const hasRunning = (data: any) => {
    const jobs = data?.data?.list ?? [];
    return jobs.some((j: any) => j.status === 'running' || j.status === 'pending');
  };

  const { data, isLoading } = useQuery({
    queryKey: ['jobs', page, status],
    queryFn: () => jobApi.list({ page, page_size: 20, status: status || undefined }),
    refetchInterval: (query) => hasRunning(query.state.data) ? 3000 : false,
  });

  const jobs = (data as any)?.data?.list ?? [];
  const total = (data as any)?.data?.total ?? 0;
  const [expanded, setExpanded] = useState<number | null>(null);

  if (isLoading) return <Loading />;

  return (
    <div>
      <PageHeader title="任务" description={`共 ${total} 个任务`} />

      {/* Filters */}
      <div className="flex gap-2 mb-6">
        {['', 'pending', 'running', 'completed', 'failed'].map(s => (
          <button
            key={s}
            onClick={() => { setStatus(s); setPage(1); }}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
              status === s ? 'bg-nexus-50 text-nexus-600' : 'text-gray-400 hover:bg-surface-hover'
            }`}
          >
            {s === '' ? '全部' : s === 'pending' ? '等待中' : s === 'running' ? '运行中' : s === 'completed' ? '已完成' : '失败'}
          </button>
        ))}
      </div>

      {jobs.length === 0 ? (
        <EmptyState message="暂无任务" />
      ) : (
        <div className="space-y-2">
          {jobs.map((job: any) => (
            <Card key={job.id} onClick={() => setExpanded(expanded === job.id ? null : job.id)} className="!p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-xs font-mono text-gray-400">{job.uuid?.slice(0, 8)}</span>
                  <StatusBadge status={job.status} />
                  {job.status === 'running' && job.total_steps > 0 && (
                    <span className="text-xs text-gray-400">{job.current_step}/{job.total_steps}</span>
                  )}
                </div>
                <span className="text-xs text-gray-400">{new Date(job.created_at).toLocaleString()}</span>
              </div>
              {expanded === job.id && (
                <>
                  <StepProgress job={job} />
                  {job.result && (
                    <pre className="mt-3 text-xs text-gray-500 bg-surface rounded-lg p-3 overflow-x-auto max-h-64">
                      {typeof job.result === 'string' ? JSON.stringify(JSON.parse(job.result), null, 2) : JSON.stringify(job.result, null, 2)}
                    </pre>
                  )}
                  {job.error && (
                    <p className="mt-3 text-xs text-red-500 bg-red-50 rounded-lg p-3">{job.error}</p>
                  )}
                </>
              )}
            </Card>
          ))}
        </div>
      )}

      {/* Pagination */}
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
