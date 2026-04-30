import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { PageHeader, Card, Badge, EmptyState, Loading, FilterTabs, Pagination } from '../../components/UI';
import { entityApi } from '../../api';

const typeFilters = [
  { key: '', label: '全部' },
  { key: 'person', label: '人物' },
  { key: 'company', label: '公司' },
  { key: 'event', label: '事件' },
  { key: 'location', label: '地点' },
];

export default function EntitiesPage() {
  const [page, setPage] = useState(1);
  const [type, setType] = useState('');
  const [selected, setSelected] = useState<number | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['entities', page, type],
    queryFn: () => entityApi.list({ page, page_size: 20, type: type || undefined }),
  });

  const entities = (data as any)?.data?.list ?? [];
  const total = (data as any)?.data?.total ?? 0;

  const { data: relData } = useQuery({
    queryKey: ['relations', selected],
    queryFn: () => entityApi.getRelations(selected!),
    enabled: !!selected,
  });
  const relations = (relData as any)?.data ?? [];

  if (isLoading) return <Loading />;

  return (
    <div>
      <PageHeader title="实体" description={`共 ${total} 个实体`} />

      <FilterTabs
        items={typeFilters}
        value={type}
        onChange={v => { setType(v); setPage(1); }}
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 space-y-2">
          {entities.length === 0 ? (
            <EmptyState message="暂无实体" />
          ) : (
            entities.map((e: any) => (
              <Card
                key={e.id}
                onClick={() => setSelected(e.id)}
                className={`!p-4 ${selected === e.id ? 'border-nexus-300 bg-nexus-50/30' : ''}`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-gray-800">{e.name}</span>
                    <Badge variant="info">{e.type}</Badge>
                    {e.confirmed && <Badge variant="success">已确认</Badge>}
                  </div>
                  <span className="text-xs text-gray-400">{Math.round(e.confidence * 100)}%</span>
                </div>
                {(() => { const a = Array.isArray(e.aliases) ? e.aliases : []; return a.length > 0 && <p className="text-xs text-gray-400 mt-1">别名: {a.join(', ')}</p>; })()}
              </Card>
            ))
          )}

          <Pagination page={page} total={total} pageSize={20} onChange={setPage} />
        </div>

        <div>
          <Card>
            <h4 className="text-sm font-medium text-gray-600 mb-3">关系</h4>
            {!selected ? (
              <p className="text-xs text-gray-400">选择一个实体查看关系</p>
            ) : relations.length === 0 ? (
              <p className="text-xs text-gray-400">暂无关系</p>
            ) : (
              <div className="space-y-2">
                {relations.map((r: any) => (
                  <div key={r.id} className="p-2 rounded-lg bg-surface text-xs">
                    <div className="flex items-center gap-1.5">
                      <span className="font-medium text-gray-700">{r.from_entity?.name}</span>
                      <span className="text-nexus-500">→</span>
                      <span className="text-gray-500">{r.type}</span>
                      <span className="text-nexus-500">→</span>
                      <span className="font-medium text-gray-700">{r.to_entity?.name}</span>
                    </div>
                    <span className="text-gray-400">{Math.round(r.confidence * 100)}%</span>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
