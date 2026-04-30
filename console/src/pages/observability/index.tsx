import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Activity, Cpu, AlertTriangle, Clock, Zap, DollarSign, ChevronDown, ChevronUp } from 'lucide-react';
import {
  LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend,
} from 'recharts';
import { PageHeader, Card, Badge, Loading, EmptyState, Tabs, FilterTabs } from '../../components/UI';
import { statsApi } from '../../api';
import type {
  PipelinePerformance, LLMPerformanceStats, ErrorAnalysis,
} from '../../api/types';

const tabItems = [
  { key: 'pipeline', label: 'Pipeline 性能', icon: Activity },
  { key: 'llm', label: 'LLM 分析', icon: Cpu },
  { key: 'errors', label: '错误追踪', icon: AlertTriangle },
];

const dayFilters = [
  { key: '7', label: '7天' },
  { key: '14', label: '14天' },
  { key: '30', label: '30天' },
];

export default function ObservabilityPage() {
  const [tab, setTab] = useState<Tab>('pipeline');
  const [days, setDays] = useState(7);

  return (
    <div>
      <PageHeader
        title="可观测性"
        description="Pipeline 性能分析、LLM 调用追踪、错误监控"
        action={
          <FilterTabs items={dayFilters} value={String(days)} onChange={v => setDays(Number(v))} />
        }
      />

      <Tabs items={tabItems} value={tab} onChange={v => setTab(v as Tab)} />

      {tab === 'pipeline' && <PipelineTab days={days} />}
      {tab === 'llm' && <LLMTab days={days} />}
      {tab === 'errors' && <ErrorsTab days={days} />}
    </div>
  );
}

// --- Pipeline Tab ---

function PipelineTab({ days }: { days: number }) {
  const { data, isLoading } = useQuery({
    queryKey: ['pipeline-performance', days],
    queryFn: () => statsApi.pipelinePerformance(days),
  });

  if (isLoading) return <Loading />;
  const items = ((data as any)?.data ?? []) as PipelinePerformance[];
  if (items.length === 0) return <EmptyState message="暂无 Pipeline 性能数据" />;

  return (
    <div className="space-y-4">
      {items.map(p => (
        <PipelineCard key={p.pipeline_id} data={p} />
      ))}
    </div>
  );
}

function PipelineCard({ data }: { data: PipelinePerformance }) {
  const [expanded, setExpanded] = useState(false);
  const rateColor = data.success_rate >= 90 ? 'text-emerald-500' : data.success_rate >= 70 ? 'text-amber-500' : 'text-red-500';

  return (
    <Card className="cursor-pointer" onClick={() => setExpanded(!expanded)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium text-gray-700">{data.pipeline_name || `Pipeline #${data.pipeline_id}`}</span>
          <Badge variant="default">{data.total_jobs} 任务</Badge>
        </div>
        <div className="flex items-center gap-6 text-xs">
          <span className="flex items-center gap-1 text-gray-500">
            <Clock className="w-3.5 h-3.5" /> 平均 {formatMs(data.avg_duration_ms)}
          </span>
          <span className="text-gray-400">P95 {formatMs(data.p95_duration_ms)}</span>
          <span className={`font-medium ${rateColor}`}>{data.success_rate.toFixed(1)}%</span>
          {expanded ? <ChevronUp className="w-4 h-4 text-gray-300" /> : <ChevronDown className="w-4 h-4 text-gray-300" />}
        </div>
      </div>

      {expanded && data.steps.length > 0 && (
        <div className="mt-4 pt-4 border-t border-border-soft">
          <ResponsiveContainer width="100%" height={Math.max(data.steps.length * 40, 120)}>
            <BarChart data={data.steps} layout="vertical" margin={{ left: 100, right: 20 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis type="number" tick={{ fontSize: 11 }} />
              <YAxis type="category" dataKey="processor_type" tick={{ fontSize: 11 }} width={90} />
              <Tooltip formatter={(v) => formatMs(Number(v))} />
              <Bar dataKey="avg_duration_ms" name="平均耗时" fill="#8b5cf6" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mt-3">
            {data.steps.map(step => (
              <div key={step.processor_type} className="text-xs bg-gray-50 rounded-lg px-3 py-2">
                <span className="font-medium text-gray-600">{step.processor_type}</span>
                <div className="flex items-center gap-3 mt-1 text-gray-400">
                  <span>{step.avg_tokens} tok</span>
                  <span>${step.avg_cost.toFixed(4)}</span>
                  <span className={step.error_rate > 5 ? 'text-red-400' : ''}>{step.error_rate.toFixed(1)}% err</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </Card>
  );
}

// --- LLM Tab ---

function LLMTab({ days }: { days: number }) {
  const { data, isLoading } = useQuery({
    queryKey: ['llm-performance', days],
    queryFn: () => statsApi.llmPerformance(days),
  });

  if (isLoading) return <Loading />;
  const stats = (data as any)?.data as LLMPerformanceStats | undefined;
  if (!stats) return <EmptyState message="暂无 LLM 性能数据" />;

  const procs = stats.by_processor ?? [];
  const daily = (stats.daily_usage ?? []).map(d => ({ ...d, date: d.date.slice(5) }));

  return (
    <div className="space-y-6">
      {procs.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {procs.map(p => (
            <Card key={p.processor_type}>
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-medium text-gray-700">{p.processor_type}</span>
                {p.error_rate > 5 && <Badge variant="error">{p.error_rate.toFixed(1)}% err</Badge>}
              </div>
              <div className="grid grid-cols-2 gap-3 text-xs">
                <div>
                  <p className="text-gray-400">调用数</p>
                  <p className="text-lg font-semibold text-gray-700">{p.total_calls}</p>
                </div>
                <div>
                  <p className="text-gray-400">平均耗时</p>
                  <p className="text-lg font-semibold text-gray-700">{formatMs(p.avg_duration_ms)}</p>
                </div>
                <div>
                  <p className="text-gray-400 flex items-center gap-1"><Zap className="w-3 h-3" />Token</p>
                  <p className="text-lg font-semibold text-gray-700">{p.total_tokens.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-gray-400 flex items-center gap-1"><DollarSign className="w-3 h-3" />费用</p>
                  <p className="text-lg font-semibold text-gray-700">${p.total_cost.toFixed(4)}</p>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {daily.length > 0 && (
        <Card>
          <h3 className="text-sm font-medium text-gray-600 mb-4">每日 Token / 费用趋势</h3>
          <ResponsiveContainer width="100%" height={260}>
            <LineChart data={daily}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="date" tick={{ fontSize: 12 }} />
              <YAxis yAxisId="tokens" tick={{ fontSize: 12 }} allowDecimals={false} />
              <YAxis yAxisId="cost" orientation="right" tick={{ fontSize: 12 }} />
              <Tooltip />
              <Legend />
              <Line yAxisId="tokens" type="monotone" dataKey="tokens" name="Token" stroke="#8b5cf6" strokeWidth={2} dot={{ r: 3 }} />
              <Line yAxisId="cost" type="monotone" dataKey="cost" name="费用($)" stroke="#f472b6" strokeWidth={2} dot={{ r: 3 }} />
            </LineChart>
          </ResponsiveContainer>
        </Card>
      )}
    </div>
  );
}

// --- Errors Tab ---

function ErrorsTab({ days }: { days: number }) {
  const { data, isLoading } = useQuery({
    queryKey: ['error-analysis', days],
    queryFn: () => statsApi.errors(days),
  });

  if (isLoading) return <Loading />;
  const analysis = (data as any)?.data as ErrorAnalysis | undefined;
  if (!analysis) return <EmptyState message="暂无错误数据" />;

  const trend = (analysis.error_trend ?? []).map(d => ({ ...d, date: d.date.slice(5) }));
  const topErrors = analysis.top_errors ?? [];
  const recent = analysis.recent_failures ?? [];

  return (
    <div className="space-y-6">
      {trend.length > 0 && (
        <Card>
          <h3 className="text-sm font-medium text-gray-600 mb-4">每日失败数趋势</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={trend}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="date" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} allowDecimals={false} />
              <Tooltip />
              <Line type="monotone" dataKey="count" name="失败数" stroke="#f87171" strokeWidth={2} dot={{ r: 3 }} />
            </LineChart>
          </ResponsiveContainer>
        </Card>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <h3 className="text-sm font-medium text-gray-600 mb-4">Top 错误类型</h3>
          {topErrors.length === 0 ? (
            <p className="text-sm text-gray-400 py-4 text-center">无错误记录</p>
          ) : (
            <div className="space-y-2">
              {topErrors.map((e, i) => (
                <div key={i} className="flex items-center justify-between py-2 px-3 rounded-lg bg-gray-50">
                  <span className="text-xs text-gray-600 line-clamp-1 flex-1 mr-3">{e.error}</span>
                  <Badge variant="error">{e.count}</Badge>
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card>
          <h3 className="text-sm font-medium text-gray-600 mb-4">最近失败任务</h3>
          {recent.length === 0 ? (
            <p className="text-sm text-gray-400 py-4 text-center">无失败任务</p>
          ) : (
            <div className="space-y-2 max-h-80 overflow-y-auto">
              {recent.map(f => (
                <div key={f.job_id} className="py-2 px-3 rounded-lg bg-gray-50 text-xs">
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono text-gray-400">{f.uuid.slice(0, 8)}</span>
                    <span className="text-gray-400">{f.created_at}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {f.pipeline && <Badge variant="default">{f.pipeline}</Badge>}
                    <span className="text-red-400 line-clamp-1">{f.error}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}

function formatMs(ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}
