import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { Search, Sparkles, Loader2, Clock, Tag, ChevronDown, ChevronUp, Zap, Type, Blend } from 'lucide-react';
import { PageHeader, Card, Badge, EmptyState } from '../../components/UI';
import { searchApi } from '../../api';

interface SearchItem {
  document_id: number;
  document_uuid: string;
  type: string;
  source_url: string;
  content: string;
  summary: string;
  entities: { id: number; type: string; name: string }[];
  score: number;
  vector_score: number;
  reason: string;
  created_at: string;
}

interface SearchStats {
  total: number;
  vector_hits: number;
  keyword_hits: number;
}

interface SearchResultData {
  items: SearchItem[];
  parsed_query: any;
  reasoning: string;
  stats: SearchStats;
}

type SearchMode = 'hybrid' | 'keyword' | 'vector';

const modeConfig: { key: SearchMode; label: string; icon: typeof Blend; desc: string }[] = [
  { key: 'hybrid', label: '智能混合', icon: Blend, desc: '向量 + 关键词 + AI 精排' },
  { key: 'keyword', label: '关键词', icon: Type, desc: '关键词匹配 + AI 精排' },
  { key: 'vector', label: '语义相似', icon: Zap, desc: '向量语义召回' },
];

function getMatchSource(item: SearchItem): { label: string; variant: 'info' | 'default' | 'success' } {
  const hasVec = item.vector_score > 0;
  const hasKeyword = !hasVec || item.score > 0;
  if (hasVec && hasKeyword && item.score > 0) return { label: '混合匹配', variant: 'info' };
  if (hasVec) return { label: '语义匹配', variant: 'success' };
  return { label: '关键词', variant: 'default' };
}

export default function SearchPage() {
  const [keyword, setKeyword] = useState('');
  const [mode, setMode] = useState<SearchMode>('hybrid');
  const [results, setResults] = useState<SearchItem[]>([]);
  const [parsedQuery, setParsedQuery] = useState<any>(null);
  const [reasoning, setReasoning] = useState('');
  const [stats, setStats] = useState<SearchStats | null>(null);
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [searchTime, setSearchTime] = useState(0);

  const searchMut = useMutation({
    mutationFn: (params: { query: string; mode: SearchMode }) => {
      const start = performance.now();
      return searchApi.search(params.query, params.mode).then((res: any) => {
        setSearchTime(Math.round(performance.now() - start));
        return res;
      });
    },
    onSuccess: (res: any) => {
      const data = res?.data as SearchResultData | undefined;
      setResults(data?.items ?? []);
      setParsedQuery(data?.parsed_query ?? null);
      setReasoning(data?.reasoning ?? '');
      setStats(data?.stats ?? null);
      setExpandedId(null);
    },
  });

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!keyword.trim()) return;
    searchMut.mutate({ query: keyword.trim(), mode });
  };

  return (
    <div>
      <PageHeader title="智能搜索" description="自然语言检索知识库中的文档和素材" />

      <form onSubmit={handleSearch} className="mb-4">
        <div className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            value={keyword}
            onChange={e => setKeyword(e.target.value)}
            placeholder="输入自然语言查询，如：滕总的照片、4月23号的会议报告..."
            className="w-full pl-12 pr-4 py-3 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
          />
          {searchMut.isPending && (
            <Loader2 className="absolute right-4 top-1/2 -translate-y-1/2 w-5 h-5 text-nexus-400 animate-spin" />
          )}
        </div>
      </form>

      {/* 搜索模式切换 */}
      <div className="flex gap-2 mb-6">
        {modeConfig.map(m => {
          const Icon = m.icon;
          const active = mode === m.key;
          return (
            <button
              key={m.key}
              onClick={() => setMode(m.key)}
              className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                active
                  ? 'bg-nexus-50 text-nexus-600 border border-nexus-200'
                  : 'bg-surface-hover text-gray-500 border border-transparent hover:text-gray-700'
              }`}
              title={m.desc}
            >
              <Icon className="w-3.5 h-3.5" />
              {m.label}
            </button>
          );
        })}
      </div>

      {/* 解析结果 */}
      {parsedQuery && (
        <div className="mb-4 flex flex-wrap gap-2 items-center text-xs text-gray-500">
          <span className="font-medium">解析结果：</span>
          {parsedQuery.entity && <Badge variant="info">{parsedQuery.entity}</Badge>}
          {parsedQuery.type && <Badge variant="default">{parsedQuery.type}</Badge>}
          {parsedQuery.date_from && <Badge variant="default">{parsedQuery.date_from}{parsedQuery.date_to && parsedQuery.date_to !== parsedQuery.date_from ? ` ~ ${parsedQuery.date_to}` : ''}</Badge>}
          {parsedQuery.keywords?.map((k: string) => <Badge key={k} variant="default">{k}</Badge>)}
          {parsedQuery.intent && <span className="text-gray-400 ml-2">意图: {parsedQuery.intent}</span>}
        </div>
      )}

      {/* 统计栏 */}
      {stats && searchMut.isSuccess && results.length > 0 && (
        <div className="mb-4 flex items-center gap-4 text-xs text-gray-400">
          <span>共 {stats.total} 条结果</span>
          <span>耗时 {searchTime}ms</span>
          {stats.vector_hits > 0 && <span className="text-nexus-500">{stats.vector_hits} 条语义匹配</span>}
          {stats.keyword_hits > 0 && <span>{stats.keyword_hits} 条关键词匹配</span>}
          {reasoning && (
            <span className="flex items-center gap-1 text-nexus-400">
              <Sparkles className="w-3.5 h-3.5" /> AI 精排已启用
            </span>
          )}
        </div>
      )}

      {/* 结果列表 */}
      {results.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          {results.map(item => {
            const source = getMatchSource(item);
            const expanded = expandedId === item.document_id;
            return (
              <Card
                key={item.document_id}
                className="overflow-hidden cursor-pointer"
                onClick={() => setExpandedId(expanded ? null : item.document_id)}
              >
                {/* 来源标识 */}
                <div className="flex items-center justify-between -mt-1 mb-2">
                  <Badge variant={source.variant}>{source.label}</Badge>
                  {expanded ? <ChevronUp className="w-4 h-4 text-gray-300" /> : <ChevronDown className="w-4 h-4 text-gray-300" />}
                </div>

                {item.type === 'image' && item.source_url && (
                  <div className="h-40 bg-gray-50 -mx-6 mb-3 overflow-hidden">
                    <img src={item.source_url} alt="" className="w-full h-full object-cover" />
                  </div>
                )}

                <div className="space-y-2">
                  {item.summary && <p className="text-sm text-gray-700 line-clamp-2">{item.summary}</p>}
                  {!item.summary && item.content && <p className={`text-sm text-gray-500 ${expanded ? '' : 'line-clamp-2'}`}>{item.content}</p>}

                  {/* 展开时显示完整内容 */}
                  {expanded && item.summary && item.content && (
                    <p className="text-xs text-gray-400 bg-gray-50 rounded-lg px-3 py-2 whitespace-pre-wrap">{item.content}</p>
                  )}

                  {item.entities?.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {(expanded ? item.entities : item.entities.slice(0, 3)).map(e => (
                        <span key={e.id} className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-nexus-50 text-nexus-600 text-[11px]">
                          <Tag className="w-3 h-3" />{e.name}
                        </span>
                      ))}
                      {!expanded && item.entities.length > 3 && (
                        <span className="text-[11px] text-gray-400">+{item.entities.length - 3}</span>
                      )}
                    </div>
                  )}

                  {/* 分数区域 */}
                  <div className="space-y-1.5 pt-1">
                    {item.vector_score > 0 && (
                      <div className="flex items-center gap-2">
                        <span className="text-[11px] text-gray-400 w-14 shrink-0">语义相似</span>
                        <div className="flex-1 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-nexus-400 to-nexus-500 rounded-full transition-all"
                            style={{ width: `${Math.min(item.vector_score * 100, 100)}%` }}
                          />
                        </div>
                        <span className="text-[11px] text-nexus-500 font-medium w-10 text-right">
                          {(item.vector_score * 100).toFixed(0)}%
                        </span>
                      </div>
                    )}

                    <div className="flex items-center justify-between text-[11px] text-gray-400">
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {new Date(item.created_at).toLocaleDateString('zh-CN')}
                      </span>
                      {item.score > 0 && <span className="text-nexus-500 font-medium">{item.score.toFixed(1)}分</span>}
                    </div>
                  </div>

                  {item.reason && (
                    <p className="text-[11px] text-gray-400 bg-gray-50 rounded px-2 py-1">{item.reason}</p>
                  )}
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {searchMut.isSuccess && results.length === 0 && (
        <EmptyState message="未找到相关结果，试试换个关键词或更具体的描述" />
      )}
    </div>
  );
}
