import { useQuery, useMutation } from '@tanstack/react-query';
import { useState, useRef, useCallback } from 'react';
import { Play, Loader2, Upload, X, CheckCircle2, XCircle, Circle } from 'lucide-react';
import { PageHeader, Card, Button, Badge } from '../../components/UI';
import { useToast } from '../../components/Toast';
import { pipelineApi, jobApi, uploadApi } from '../../api';

interface StepStatus {
  order: number;
  processor: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  error?: string;
}

export default function PlaygroundPage() {
  const toast = useToast();
  const [content, setContent] = useState('');
  const [type, setType] = useState('text');
  const [pipelineId, setPipelineId] = useState<number | null>(null);
  const [result, setResult] = useState<any>(null);
  const [processing, setProcessing] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [sourceUrl, setSourceUrl] = useState('');
  const [previewUrl, setPreviewUrl] = useState('');
  const [dragOver, setDragOver] = useState(false);
  const [steps, setSteps] = useState<StepStatus[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  const { data: pipelinesData } = useQuery({ queryKey: ['pipelines'], queryFn: () => pipelineApi.list() });
  const pipelines = (pipelinesData as any)?.data ?? [];

  const selectedPipeline = pipelines.find((p: any) => p.id === pipelineId);
  const pipelineSteps: { sort_order: number; processor_type: string }[] = selectedPipeline?.steps ?? [];

  const connectSSE = useCallback((jobUuid: string) => {
    const token = localStorage.getItem('token') || '';
    const es = new EventSource(`/api/admin/jobs/${jobUuid}/events?token=${token}`);
    eventSourceRef.current = es;

    es.onmessage = (e) => {
      try {
        const evt = JSON.parse(e.data);
        if (evt.type === 'step_start') {
          setSteps(prev => prev.map(s =>
            s.order === evt.step ? { ...s, status: 'running' } : s
          ));
        } else if (evt.type === 'step_end') {
          setSteps(prev => prev.map(s =>
            s.order === evt.step
              ? { ...s, status: evt.error ? 'failed' : 'completed', error: evt.error }
              : s
          ));
        } else if (evt.type === 'completed') {
          es.close();
          eventSourceRef.current = null;
          jobApi.status(jobUuid).then((res: any) => {
            const job = res?.data;
            setResult(typeof job?.result === 'string' ? JSON.parse(job.result) : job?.result ?? job);
            setProcessing(false);
          });
        } else if (evt.type === 'failed') {
          es.close();
          eventSourceRef.current = null;
          setProcessing(false);
          toast.error('任务失败: ' + (evt.error || '未知错误'));
          setResult({ error: evt.error });
        }
      } catch { /* ignore parse errors */ }
    };

    es.onerror = () => {
      es.close();
      eventSourceRef.current = null;
      // fallback to polling
      fallbackPoll(jobUuid);
    };
  }, [toast]);

  const fallbackPoll = async (uuid: string) => {
    for (let i = 0; i < 60; i++) {
      await new Promise(r => setTimeout(r, 3000));
      try {
        const res = await jobApi.status(uuid) as any;
        const job = res?.data;
        if (job?.step_logs) {
          setSteps(prev => prev.map(s => {
            const log = job.step_logs.find((l: any) => l.step_order === s.order);
            if (log) return { ...s, status: log.status, error: log.error };
            return s;
          }));
        }
        if (job?.status === 'completed' || job?.status === 'partial') {
          setResult(typeof job.result === 'string' ? JSON.parse(job.result) : job.result ?? job);
          setProcessing(false);
          return;
        }
        if (job?.status === 'failed') {
          setResult({ error: job.error });
          setProcessing(false);
          toast.error('任务失败: ' + (job.error || '未知错误'));
          return;
        }
      } catch { /* continue */ }
    }
    setProcessing(false);
    toast.error('轮询超时');
  };

  const submitMut = useMutation({
    mutationFn: (data: any) => jobApi.submit(data),
    onSuccess: (res: any) => {
      const data = res?.data;
      const jobUuid = data?.job_id;
      if (!jobUuid) { toast.error('提交失败'); return; }

      if (data?.cached) {
        toast.success('命中缓存，直接返回结果');
        const r = typeof data.result === 'string' ? JSON.parse(data.result) : data.result;
        setResult(r);
        return;
      }

      toast.success('任务已提交');
      setProcessing(true);
      setSteps(pipelineSteps.map(s => ({
        order: s.sort_order,
        processor: s.processor_type,
        status: 'pending' as const,
      })));
      connectSSE(jobUuid);
    },
    onError: () => toast.error('提交失败'),
  });

  const handleUpload = useCallback(async (file: File) => {
    if (!file.type.startsWith('image/') && type === 'image') {
      toast.error('请选择图片文件');
      return;
    }
    setUploading(true);
    try {
      const res = await uploadApi.upload(file) as any;
      const url = res?.data?.url;
      if (url) {
        setSourceUrl(url);
        setPreviewUrl(url);
        toast.success('上传成功');
      } else {
        toast.error('上传失败');
      }
    } catch {
      toast.error('上传失败');
    } finally {
      setUploading(false);
    }
  }, [type, toast]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) handleUpload(file);
  }, [handleUpload]);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) handleUpload(file);
    e.target.value = '';
  };

  const clearUpload = () => {
    setSourceUrl('');
    setPreviewUrl('');
  };

  const handleSubmit = () => {
    if (type === 'image' && !sourceUrl.trim()) { toast.error('请上传图片或粘贴图片 URL'); return; }
    if (type !== 'image' && !content.trim()) { toast.error('请填写内容'); return; }
    if (!pipelineId) { toast.error('请选择流水线'); return; }
    if (eventSourceRef.current) { eventSourceRef.current.close(); eventSourceRef.current = null; }
    setResult(null);
    setSteps([]);
    const payload: any = { type, pipeline_id: pipelineId, skip_cache: true };
    if (type === 'image') {
      payload.source_url = sourceUrl.trim();
      if (content.trim()) payload.content = content.trim();
    } else {
      payload.content = content.trim();
    }
    submitMut.mutate(payload);
  };

  return (
    <div>
      <PageHeader title="测试" description="提交内容到流水线，查看提取结果" />

      <div className="grid grid-cols-2 gap-6">
        {/* Input */}
        <Card>
          <h3 className="text-sm font-medium text-gray-700 mb-3">输入</h3>
          <div className="space-y-3">
            <div className="flex gap-3">
              <select
                value={type}
                onChange={e => { setType(e.target.value); if (e.target.value === 'text') { setPreviewUrl(''); setSourceUrl(''); } }}
                className="px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
              >
                <option value="text">文本</option>
                <option value="image">图片</option>
              </select>
              <select
                value={pipelineId ?? ''}
                onChange={e => setPipelineId(Number(e.target.value) || null)}
                className="flex-1 px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
              >
                <option value="">选择流水线</option>
                {pipelines.map((p: any) => <option key={p.id} value={p.id}>{p.name}</option>)}
              </select>
            </div>

            {type === 'image' && (
              <div>
                <input ref={fileInputRef} type="file" accept="image/*" className="hidden" onChange={handleFileSelect} />
                {previewUrl ? (
                  <div className="relative rounded-xl border border-border-soft overflow-hidden bg-gray-50">
                    <img src={previewUrl} alt="preview" className="w-full max-h-48 object-contain" />
                    <button
                      onClick={clearUpload}
                      className="absolute top-2 right-2 w-6 h-6 rounded-full bg-black/50 text-white flex items-center justify-center hover:bg-black/70 transition-colors"
                    >
                      <X className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ) : (
                  <div
                    onDragOver={e => { e.preventDefault(); setDragOver(true); }}
                    onDragLeave={() => setDragOver(false)}
                    onDrop={handleDrop}
                    onClick={() => fileInputRef.current?.click()}
                    className={`flex flex-col items-center justify-center gap-2 py-10 rounded-xl border-2 border-dashed cursor-pointer transition-all ${
                      dragOver ? 'border-nexus-400 bg-nexus-50/50' : 'border-border-soft hover:border-nexus-300 hover:bg-surface-hover'
                    }`}
                  >
                    {uploading ? (
                      <Loader2 className="w-8 h-8 text-nexus-400 animate-spin" />
                    ) : (
                      <>
                        <div className="w-10 h-10 rounded-full bg-nexus-50 flex items-center justify-center">
                          <Upload className="w-5 h-5 text-nexus-400" />
                        </div>
                        <p className="text-sm text-gray-400">拖拽图片到此处，或点击选择</p>
                        <p className="text-xs text-gray-300">支持 JPG、PNG、WebP</p>
                      </>
                    )}
                  </div>
                )}
              </div>
            )}

            {type === 'image' && !previewUrl && (
              <input
                value={sourceUrl}
                onChange={e => { setSourceUrl(e.target.value); setPreviewUrl(e.target.value); }}
                placeholder="或直接粘贴图片 URL"
                className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
              />
            )}

            <textarea
              value={content}
              onChange={e => setContent(e.target.value)}
              placeholder={type === 'image' ? '备注信息（可选）' : '输入要处理的文本内容...'}
              className={`w-full px-4 py-3 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 resize-none ${
                type === 'image' ? 'h-20' : 'h-64'
              }`}
            />
            <Button onClick={handleSubmit} disabled={submitMut.isPending || processing} className="w-full">
              {processing ? <><Loader2 className="w-4 h-4 animate-spin" /> 处理中...</> : <><Play className="w-4 h-4" /> 执行</>}
            </Button>
          </div>
        </Card>

        {/* Output */}
        <Card>
          <h3 className="text-sm font-medium text-gray-700 mb-3">结果</h3>
          {!result && !processing && steps.length === 0 && <p className="text-sm text-gray-400">提交内容后查看结果</p>}

          {steps.length > 0 && (
            <div className="mb-4 space-y-1.5">
              <h4 className="text-xs font-medium text-gray-500 mb-2">流水线进度</h4>
              {steps.map(s => (
                <div key={s.order} className="flex items-center gap-2 text-sm">
                  {s.status === 'completed' && <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0" />}
                  {s.status === 'running' && <Loader2 className="w-4 h-4 text-nexus-500 animate-spin shrink-0" />}
                  {s.status === 'failed' && <XCircle className="w-4 h-4 text-red-500 shrink-0" />}
                  {s.status === 'pending' && <Circle className="w-4 h-4 text-gray-300 shrink-0" />}
                  <span className={`${s.status === 'running' ? 'text-nexus-600 font-medium' : s.status === 'completed' ? 'text-gray-600' : s.status === 'failed' ? 'text-red-600' : 'text-gray-400'}`}>
                    {s.processor}
                  </span>
                  {s.error && <span className="text-[11px] text-red-400 truncate max-w-[200px]">{s.error}</span>}
                </div>
              ))}
            </div>
          )}
          {result && (
            <div className="space-y-4 overflow-y-auto max-h-[calc(100vh-220px)]">
              {result.error && <div className="text-sm text-red-500 bg-red-50 p-3 rounded-xl">{result.error}</div>}

              {result.metadata?.processors_used?.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-gray-500 mb-2">处理流程</h4>
                  <div className="flex flex-wrap gap-1.5">
                    {result.metadata.processors_used.map((p: string, i: number) => (
                      <span key={i} className="inline-flex items-center gap-1 px-2 py-1 rounded-lg bg-nexus-50 text-nexus-600 text-xs">
                        <span className="w-4 h-4 rounded-full bg-nexus-200 text-nexus-700 text-[10px] flex items-center justify-center font-medium">{i + 1}</span>
                        {p}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {result.content?.summary && (
                <div>
                  <h4 className="text-xs font-medium text-gray-500 mb-2">摘要</h4>
                  <p className="text-sm text-gray-600 bg-surface rounded-lg p-3">{result.content.summary}</p>
                </div>
              )}

              {result.content?.raw_text && (
                <details>
                  <summary className="text-xs text-gray-400 cursor-pointer hover:text-gray-600">识别原文</summary>
                  <pre className="mt-2 text-xs text-gray-500 bg-surface rounded-lg p-3 whitespace-pre-wrap max-h-32 overflow-auto">{result.content.raw_text}</pre>
                </details>
              )}

              {result.extras?.image_assessment?.use_cases?.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-gray-500 mb-2">AI 应用场景评估</h4>
                  <div className="space-y-2">
                    {result.extras.image_assessment.use_cases.map((uc: any, i: number) => (
                      <div key={i} className="bg-surface rounded-lg p-3">
                        <div className="flex items-center justify-between mb-1.5">
                          <div className="flex items-center gap-2">
                            <span className="text-sm font-medium text-gray-700">{uc.scene}</span>
                            <Badge variant={uc.suitable ? 'success' : 'default'}>{uc.suitable ? '适合' : '不适合'}</Badge>
                          </div>
                          <span className="text-xs font-medium text-gray-500">{uc.score}/10</span>
                        </div>
                        <div className="w-full h-1.5 bg-gray-100 rounded-full overflow-hidden mb-1.5">
                          <div className={`h-full rounded-full ${uc.score >= 6 ? 'bg-emerald-400' : uc.score >= 4 ? 'bg-amber-400' : 'bg-red-300'}`} style={{ width: `${uc.score * 10}%` }} />
                        </div>
                        <p className="text-[11px] text-gray-400">{uc.reason}</p>
                        {uc.tags?.length > 0 && (
                          <div className="mt-1 flex flex-wrap gap-1">
                            {uc.tags.map((t: string) => <span key={t} className="text-[11px] text-nexus-600 bg-nexus-50 px-1.5 py-0.5 rounded">{t}</span>)}
                          </div>
                        )}
                      </div>
                    ))}
                    {result.extras.image_assessment.overall && (
                      <p className="text-xs text-gray-500 bg-gray-50 rounded-lg p-2">{result.extras.image_assessment.overall}</p>
                    )}
                  </div>
                </div>
              )}

              {result.entities?.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-gray-500 mb-2">实体 ({result.entities.length})</h4>
                  <div className="space-y-2">
                    {result.entities.map((e: any, i: number) => (
                      <div key={i} className="bg-surface rounded-lg p-3">
                        <div className="flex items-center gap-2">
                          <Badge variant="info">{e.type}</Badge>
                          <span className="text-sm font-medium text-gray-700">{e.name}</span>
                          <span className="text-gray-400 text-xs">{(e.confidence * 100).toFixed(0)}%</span>
                        </div>
                        {e.attributes && Object.keys(e.attributes).filter(k => k !== 'existing_id').length > 0 && (
                          <div className="mt-1.5 flex flex-wrap gap-1">
                            {Object.entries(e.attributes).filter(([k]) => k !== 'existing_id').map(([k, v]) => (
                              <span key={k} className="text-[11px] text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded">{k}: {String(v)}</span>
                            ))}
                          </div>
                        )}
                        {e.evidence?.detail && <p className="mt-1 text-[11px] text-gray-400">{e.evidence.detail}</p>}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {result.relations?.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-gray-500 mb-2">关系 ({result.relations.length})</h4>
                  <div className="space-y-1.5">
                    {result.relations.map((r: any, i: number) => (
                      <div key={i} className="text-sm text-gray-600">
                        <span className="font-medium">{r.from}</span>
                        <span className="mx-1.5 text-nexus-500">{r.type}</span>
                        <span className="font-medium">{r.to}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <details className="text-xs">
                <summary className="text-gray-400 cursor-pointer hover:text-gray-600">原始 JSON</summary>
                <pre className="mt-2 text-gray-500 bg-gray-50 p-3 rounded-xl overflow-auto max-h-60 whitespace-pre-wrap">{JSON.stringify(result, null, 2)}</pre>
              </details>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
