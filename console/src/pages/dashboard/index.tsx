import { useQuery } from '@tanstack/react-query';
import { GitBranch, ListTodo, CheckCircle, Database } from 'lucide-react';
import { PageHeader, Card, StatusBadge } from '../../components/UI';
import { pipelineApi, jobApi, reviewApi, entityApi } from '../../api';

export default function DashboardPage() {
  const { data: pipelines } = useQuery({ queryKey: ['pipelines'], queryFn: () => pipelineApi.list() });
  const { data: jobs } = useQuery({ queryKey: ['jobs'], queryFn: () => jobApi.list({ page: 1, page_size: 5 }) });
  const { data: reviews } = useQuery({ queryKey: ['reviews'], queryFn: () => reviewApi.list({ page: 1, page_size: 1, status: 'pending' }) });
  const { data: entities } = useQuery({ queryKey: ['entities'], queryFn: () => entityApi.list({ page: 1, page_size: 1 }) });

  const stats = [
    { label: '流水线', value: (pipelines as any)?.data?.length ?? 0, icon: GitBranch, color: 'from-nexus-400 to-nexus-500' },
    { label: '总任务数', value: (jobs as any)?.data?.total ?? 0, icon: ListTodo, color: 'from-lavender-300 to-lavender-400' },
    { label: '待审核', value: (reviews as any)?.data?.total ?? 0, icon: CheckCircle, color: 'from-sakura-300 to-sakura-400' },
    { label: '实体', value: (entities as any)?.data?.total ?? 0, icon: Database, color: 'from-emerald-300 to-emerald-400' },
  ];

  const recentJobs = (jobs as any)?.data?.list ?? [];

  return (
    <div>
      <PageHeader title="仪表盘" description="Nexus 实例概览" />

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {stats.map(({ label, value, icon: Icon, color }) => (
          <Card key={label}>
            <div className="flex items-center gap-4">
              <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center`}>
                <Icon className="w-5 h-5 text-white" />
              </div>
              <div>
                <p className="text-2xl font-semibold text-gray-800">{value}</p>
                <p className="text-xs text-gray-400">{label}</p>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* Recent Jobs */}
      <Card>
        <h3 className="text-sm font-medium text-gray-600 mb-4">最近任务</h3>
        {recentJobs.length === 0 ? (
          <p className="text-sm text-gray-400 py-4 text-center">暂无任务</p>
        ) : (
          <div className="space-y-2">
            {recentJobs.map((job: any) => (
              <div key={job.id} className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-surface-hover transition-colors">
                <div className="flex items-center gap-3">
                  <span className="text-xs font-mono text-gray-400">{job.uuid?.slice(0, 8)}</span>
                  <StatusBadge status={job.status} />
                </div>
                <span className="text-xs text-gray-400">{new Date(job.created_at).toLocaleString()}</span>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
}
