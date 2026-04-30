import { useQuery } from '@tanstack/react-query';
import { ListTodo, CheckCircle, Zap, DollarSign } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import { PageHeader, Card, Loading, StatusBadge } from '../../components/UI';
import { statsApi, jobApi } from '../../api';
import type { DashboardStats } from '../../api/types';

const PIE_COLORS = ['#8b5cf6', '#f472b6', '#34d399', '#fbbf24', '#60a5fa', '#f87171', '#a78bfa', '#fb923c'];

function StatCard({ label, value, icon: Icon, color }: { label: string; value: string | number; icon: React.ElementType; color: string }) {
  return (
    <Card>
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
  );
}

function TrendChart({ data }: { data: DashboardStats['daily_trend'] }) {
  const formatted = data.map((d) => ({ ...d, date: d.date.slice(5) }));
  return (
    <Card>
      <h3 className="text-sm font-medium text-gray-600 mb-4">最近 7 天任务趋势</h3>
      <ResponsiveContainer width="100%" height={260}>
        <LineChart data={formatted}>
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis dataKey="date" tick={{ fontSize: 12 }} />
          <YAxis tick={{ fontSize: 12 }} allowDecimals={false} />
          <Tooltip />
          <Line type="monotone" dataKey="total" name="总数" stroke="#8b5cf6" strokeWidth={2} dot={{ r: 3 }} />
          <Line type="monotone" dataKey="completed" name="完成" stroke="#34d399" strokeWidth={2} dot={{ r: 3 }} />
          <Line type="monotone" dataKey="failed" name="失败" stroke="#f87171" strokeWidth={2} dot={{ r: 3 }} />
        </LineChart>
      </ResponsiveContainer>
    </Card>
  );
}

function EntityPieChart({ data }: { data: DashboardStats['entities']['distribution'] }) {
  if (!data || data.length === 0) {
    return (
      <Card>
        <h3 className="text-sm font-medium text-gray-600 mb-4">实体类型分布</h3>
        <p className="text-sm text-gray-400 py-8 text-center">暂无数据</p>
      </Card>
    );
  }

  const sorted = [...data].sort((a, b) => b.count - a.count);
  const top = sorted.slice(0, 8);
  const rest = sorted.slice(8);
  const merged = rest.length > 0
    ? [...top, { type: '其他', count: rest.reduce((s, r) => s + r.count, 0) }]
    : top;

  return (
    <Card>
      <h3 className="text-sm font-medium text-gray-600 mb-4">实体类型分布</h3>
      <ResponsiveContainer width="100%" height={260}>
        <PieChart>
          <Pie data={merged} dataKey="count" nameKey="type" cx="50%" cy="50%" outerRadius={90} label={(props) => `${props.name ?? ''} ${((props.percent ?? 0) * 100).toFixed(0)}%`}>
            {merged.map((_, i) => (
              <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
            ))}
          </Pie>
          <Tooltip />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </Card>
  );
}

export default function DashboardPage() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: () => statsApi.dashboard(),
  });
  const { data: jobs } = useQuery({
    queryKey: ['recent-jobs'],
    queryFn: () => jobApi.list({ page: 1, page_size: 5 }),
  });

  if (isLoading) return <Loading />;

  const s = (stats as any)?.data as DashboardStats | undefined;
  const successRate = s && s.jobs.total > 0 ? ((s.jobs.completed / s.jobs.total) * 100).toFixed(1) + '%' : '0%';
  const recentJobs = (jobs as any)?.data?.list ?? [];

  return (
    <div>
      <PageHeader title="仪表盘" description="Nexus 实例概览" />

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="总任务" value={s?.jobs.total ?? 0} icon={ListTodo} color="from-nexus-400 to-nexus-500" />
        <StatCard label="成功率" value={successRate} icon={CheckCircle} color="from-emerald-300 to-emerald-400" />
        <StatCard label="总 Token" value={(s?.llm.total_tokens ?? 0).toLocaleString()} icon={Zap} color="from-lavender-300 to-lavender-400" />
        <StatCard label="总费用" value={`$${(s?.llm.total_cost ?? 0).toFixed(4)}`} icon={DollarSign} color="from-sakura-300 to-sakura-400" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-8">
        <TrendChart data={s?.daily_trend ?? []} />
        <EntityPieChart data={s?.entities.distribution ?? []} />
      </div>

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
