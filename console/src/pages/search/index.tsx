import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { Search, Sparkles, Loader2, Clock, Tag } from 'lucide-react';
import { PageHeader, Card, Badge, EmptyState } from '../../components/UI';
import { searchApi } from '../../api';

interface SearchItem {
  document_id: number;
  type: string;
  source_url: string;
  content: string;
  summary: string;
  entities: { id: number; type: string; name: string }[];
  score: number;
  reason: string;
  created_at: string;
}

export default function SearchPage() {
  const [keyword, setKeyword] = useState('');
  const [results, setResults] = useState<SearchItem[]>([]);
  const [parsedQuery, setParsedQuery] = useState<any>(null);
  const [reasoning, setReasoning] = useState('');

  const searchMut = useMutation({
    mutationFn: (query: string) => searchApi.search(query),
    onSuccess: (res: any) => {
      const data = res?.data;
      setResults(data?.items ?? []);
      setParsedQuery(data?.parsed_query ?? null);
      setReasoning(data?.reasoning ?? '');
    },
  });

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!keyword.trim()) return;
    searchMut.mutate(keyword.trim());
  };

  return (
    <div>
      <PageHeader title="智能搜索" description="自然语言检索知识库中的文档和素材" />

      <form onSubmit={handleSearch} className="mb-6">
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

      {reasoning && (
        <div className="mb-4 text-xs text-gray-400 flex items-center gap-1">
          <Sparkles className="w-3.5 h-3.5" /> AI 精排已启用
        </div>
      )}

      {results.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          {results.map(item => (
            <Card key={item.document_id} className="overflow-hidden">
              {item.type === 'image' && item.source_url && (
                <div className="h-40 bg-gray-50 -mx-4 -mt-4 mb-3 overflow-hidden">
                  <img src={item.source_url} alt="" className="w-full h-full object-cover" />
                </div>
              )}
              <div className="space-y-2">
                {item.summary && <p className="text-sm text-gray-700 line-clamp-2">{item.summary}</p>}
                {!item.summary && item.content && <p className="text-sm text-gray-500 line-clamp-2">{item.content}</p>}

                {item.entities?.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {item.entities.map(e => (
                      <span key={e.id} className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-nexus-50 text-nexus-600 text-[11px]">
                        <Tag className="w-3 h-3" />{e.name}
                      </span>
                    ))}
                  </div>
                )}

                <div className="flex items-center justify-between text-[11px] text-gray-400">
                  <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" />
                    {new Date(item.created_at).toLocaleDateString('zh-CN')}
                  </span>
                  {item.score > 0 && <span className="text-nexus-500 font-medium">{item.score.toFixed(1)}分</span>}
                </div>

                {item.reason && (
                  <p className="text-[11px] text-gray-400 bg-gray-50 rounded px-2 py-1">{item.reason}</p>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}

      {searchMut.isSuccess && results.length === 0 && (
        <EmptyState message="未找到相关结果，试试换个关键词或更具体的描述" />
      )}
    </div>
  );
}

