import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { ArrowLeft, Plus, GripVertical, Trash2, Pencil, Check, X } from 'lucide-react';
import { DndContext, closestCenter, PointerSensor, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { PageHeader, Card, Button, Badge, Loading } from '../../components/UI';
import { ConfirmDialog } from '../../components/ConfirmDialog';
import { useToast } from '../../components/Toast';
import { pipelineApi, promptApi, llmProviderApi } from '../../api';

const processorTypes = ['llm_extract', 'llm_review', 'embedding', 'ocr', 'face', 'classifier', 'context_loader', 'entity_align', 'image_assess'];
const llmProcessors = ['llm_extract', 'llm_review', 'embedding'];

interface StepFormData {
  processor_type: string;
  prompt_template_id: number | null;
  config: string;
  condition: string;
  provider: string;
  model: string;
}

function parseConfigWithLLM(configStr: string): { provider: string; model: string; rest: Record<string, any> } {
  try {
    const obj = JSON.parse(configStr || '{}');
    const { provider = '', model = '', ...rest } = obj;
    return { provider, model, rest };
  } catch {
    return { provider: '', model: '', rest: {} };
  }
}

function buildConfigStr(provider: string, model: string, restStr: string): string {
  try {
    const rest = JSON.parse(restStr || '{}');
    const obj: Record<string, any> = { ...rest };
    if (provider) obj.provider = provider;
    if (model) obj.model = model;
    return JSON.stringify(obj);
  } catch {
    const obj: Record<string, any> = {};
    if (provider) obj.provider = provider;
    if (model) obj.model = model;
    return JSON.stringify(obj);
  }
}

function ProviderModelSelect({ provider, model, onProviderChange, onModelChange }: {
  provider: string;
  model: string;
  onProviderChange: (v: string) => void;
  onModelChange: (v: string) => void;
}) {
  const { data: providersData } = useQuery({ queryKey: ['llm-providers'], queryFn: () => llmProviderApi.list() });
  const providers: any[] = (providersData as any)?.data ?? [];

  const { data: modelsData } = useQuery({
    queryKey: ['llm-models', provider],
    queryFn: () => llmProviderApi.listModels(provider),
    enabled: !!provider,
  });
  const models: any[] = (modelsData as any)?.data ?? [];

  return (
    <div className="grid grid-cols-2 gap-3">
      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1.5">LLM 服务商</label>
        <select
          value={provider}
          onChange={e => { onProviderChange(e.target.value); onModelChange(''); }}
          className="w-full px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
        >
          <option value="">默认</option>
          {providers.filter(p => p.active).map(p => (
            <option key={p.name} value={p.name}>{p.display_name || p.name}</option>
          ))}
        </select>
      </div>
      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1.5">模型</label>
        <div className="relative">
          <input
            list="model-options"
            value={model}
            onChange={e => onModelChange(e.target.value)}
            placeholder={provider ? '选择或输入模型' : '使用默认模型'}
            className="w-full px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
          />
          <datalist id="model-options">
            {models.map(m => <option key={m.id} value={m.id} />)}
          </datalist>
        </div>
      </div>
    </div>
  );
}

function SortableStepCard({ step, index, editingStep, editStepForm, setEditStepForm, setEditingStep, openEditStep, setDeleteStepId, updateStepMut, buildSubmitData, prompts }: any) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: step.id });
  const style = { transform: CSS.Transform.toString(transform), transition, opacity: isDragging ? 0.5 : 1 };

  return (
    <div ref={setNodeRef} style={style}>
      <Card>
        {editingStep === step.id ? (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">处理器</label>
                <select
                  value={editStepForm.processor_type}
                  onChange={e => setEditStepForm({ ...editStepForm, processor_type: e.target.value })}
                  className="w-full px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
                >
                  {processorTypes.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">提示词模板</label>
                <select
                  value={editStepForm.prompt_template_id ?? ''}
                  onChange={e => setEditStepForm({ ...editStepForm, prompt_template_id: e.target.value ? Number(e.target.value) : null })}
                  className="w-full px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
                >
                  <option value="">无（使用默认）</option>
                  {prompts.map((p: any) => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
            </div>
            {llmProcessors.includes(editStepForm.processor_type) && (
              <ProviderModelSelect
                provider={editStepForm.provider}
                model={editStepForm.model}
                onProviderChange={v => setEditStepForm({ ...editStepForm, provider: v })}
                onModelChange={v => setEditStepForm({ ...editStepForm, model: v })}
              />
            )}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">条件</label>
                <input
                  value={editStepForm.condition}
                  onChange={e => setEditStepForm({ ...editStepForm, condition: e.target.value })}
                  placeholder="例如 type=image"
                  className="w-full px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">其他配置 (JSON)</label>
                <input
                  value={editStepForm.config}
                  onChange={e => setEditStepForm({ ...editStepForm, config: e.target.value })}
                  className="w-full px-3 py-2 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 font-mono"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setEditingStep(null)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600 transition-all">
                <X className="w-4 h-4" />
              </button>
              <button
                onClick={() => updateStepMut.mutate({ stepId: step.id, data: buildSubmitData(editStepForm) })}
                className="p-1.5 rounded-lg hover:bg-nexus-50 text-nexus-400 hover:text-nexus-600 transition-all"
              >
                <Check className="w-4 h-4" />
              </button>
            </div>
          </div>
        ) : (
          <div className="flex items-center gap-4">
            <div {...attributes} {...listeners}>
              <GripVertical className="w-4 h-4 text-gray-300 cursor-grab" />
            </div>
            <div className="w-8 h-8 rounded-lg bg-nexus-50 flex items-center justify-center text-xs font-medium text-nexus-600">
              {index + 1}
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-gray-700">{step.processor_type}</span>
                {step.prompt_template && <Badge variant="info">{step.prompt_template.name}</Badge>}
                {step.condition && <Badge variant="default">条件: {step.condition}</Badge>}
                {step.config?.provider && <Badge variant="default">{step.config.provider}{step.config.model ? ` / ${step.config.model}` : ''}</Badge>}
              </div>
            </div>
            <button
              onClick={() => openEditStep(step)}
              className="p-1.5 rounded-lg hover:bg-nexus-50 text-gray-300 hover:text-nexus-400 transition-all"
            >
              <Pencil className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setDeleteStepId(step.id)}
              className="p-1.5 rounded-lg hover:bg-red-50 text-gray-300 hover:text-red-400 transition-all"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          </div>
        )}
      </Card>
    </div>
  );
}

export default function PipelineDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [showAddStep, setShowAddStep] = useState(false);
  const [deleteStepId, setDeleteStepId] = useState<number | null>(null);
  const [editingStep, setEditingStep] = useState<number | null>(null);
  const [editStepForm, setEditStepForm] = useState<StepFormData>({ processor_type: '', prompt_template_id: null, config: '{}', condition: '', provider: '', model: '' });
  const [editingPipeline, setEditingPipeline] = useState(false);
  const [pipelineForm, setPipelineForm] = useState({ name: '', description: '' });
  const [stepForm, setStepForm] = useState<StepFormData>({ processor_type: 'llm_extract', prompt_template_id: null, config: '{}', condition: '', provider: '', model: '' });

  const { data, isLoading } = useQuery({ queryKey: ['pipeline', id], queryFn: () => pipelineApi.get(Number(id)) });
  const pipeline = (data as any)?.data;

  const { data: promptsData } = useQuery({ queryKey: ['prompts'], queryFn: () => promptApi.list() });
  const prompts = (promptsData as any)?.data ?? [];

  const addStepMut = useMutation({
    mutationFn: (d: any) => pipelineApi.createStep(Number(id), d),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline', id] });
      setShowAddStep(false);
      toast.success('步骤已添加');
    },
    onError: () => toast.error('添加失败'),
  });

  const updatePipelineMut = useMutation({
    mutationFn: (d: any) => pipelineApi.update(Number(id), d),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline', id] });
      setEditingPipeline(false);
      toast.success('已保存');
    },
    onError: () => toast.error('保存失败'),
  });

  const updateStepMut = useMutation({
    mutationFn: ({ stepId, data }: { stepId: number; data: any }) => pipelineApi.updateStep(Number(id), stepId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline', id] });
      setEditingStep(null);
      toast.success('步骤已更新');
    },
    onError: () => toast.error('更新失败'),
  });

  const deleteStepMut = useMutation({
    mutationFn: (stepId: number) => pipelineApi.deleteStep(Number(id), stepId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pipeline', id] });
      setDeleteStepId(null);
      toast.success('步骤已删除');
    },
    onError: () => { setDeleteStepId(null); toast.error('删除失败'); },
  });

  const reorderMut = useMutation({
    mutationFn: (stepIds: number[]) => pipelineApi.reorderSteps(Number(id), stepIds),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['pipeline', id] }),
    onError: () => toast.error('排序失败'),
  });

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }));

  if (isLoading) return <Loading />;
  if (!pipeline) return null;

  const steps = pipeline.steps ?? [];

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = steps.findIndex((s: any) => s.id === active.id);
    const newIndex = steps.findIndex((s: any) => s.id === over.id);
    if (oldIndex === -1 || newIndex === -1) return;
    const reordered = [...steps];
    const [moved] = reordered.splice(oldIndex, 1);
    reordered.splice(newIndex, 0, moved);
    reorderMut.mutate(reordered.map((s: any) => s.id));
  };

  const buildSubmitData = (form: StepFormData) => {
    const configStr = llmProcessors.includes(form.processor_type)
      ? buildConfigStr(form.provider, form.model, form.config)
      : form.config;
    return {
      processor_type: form.processor_type,
      prompt_template_id: form.prompt_template_id,
      config: JSON.parse(configStr || '{}'),
      condition: form.condition,
    };
  };

  const openEditStep = (step: any) => {
    const rawConfig = step.config ? JSON.stringify(step.config) : '{}';
    const { provider, model, rest } = parseConfigWithLLM(rawConfig);
    setEditingStep(step.id);
    setEditStepForm({
      processor_type: step.processor_type,
      prompt_template_id: step.prompt_template_id ?? null,
      config: JSON.stringify(rest),
      condition: step.condition || '',
      provider,
      model,
    });
  };

  return (
    <div>
      <button onClick={() => navigate('/pipelines')} className="flex items-center gap-1.5 text-sm text-gray-400 hover:text-gray-600 mb-4 transition-colors">
        <ArrowLeft className="w-4 h-4" /> 返回
      </button>

      {editingPipeline ? (
        <Card className="mb-6">
          <div className="space-y-3">
            <input
              value={pipelineForm.name}
              onChange={e => setPipelineForm({ ...pipelineForm, name: e.target.value })}
              className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm font-medium focus:outline-none focus:border-nexus-300"
              placeholder="流水线名称"
            />
            <textarea
              value={pipelineForm.description}
              onChange={e => setPipelineForm({ ...pipelineForm, description: e.target.value })}
              className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 h-20 resize-none"
              placeholder="描述"
            />
            <div className="flex justify-end gap-2">
              <Button variant="secondary" onClick={() => setEditingPipeline(false)}>取消</Button>
              <Button loading={updatePipelineMut.isPending} onClick={() => updatePipelineMut.mutate(pipelineForm)}>保存</Button>
            </div>
          </div>
        </Card>
      ) : (
        <PageHeader
          title={pipeline.name}
          description={pipeline.description}
          action={
            <div className="flex gap-2">
              <Button variant="secondary" onClick={() => { setPipelineForm({ name: pipeline.name, description: pipeline.description || '' }); setEditingPipeline(true); }}>
                <Pencil className="w-4 h-4" /> 编辑
              </Button>
              <Button onClick={() => setShowAddStep(true)}><Plus className="w-4 h-4" /> 添加步骤</Button>
            </div>
          }
        />
      )}

      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={steps.map((s: any) => s.id)} strategy={verticalListSortingStrategy}>
          <div className="space-y-3">
            {steps.map((step: any, i: number) => (
              <SortableStepCard key={step.id} step={step} index={i} editingStep={editingStep} editStepForm={editStepForm} setEditStepForm={setEditStepForm} setEditingStep={setEditingStep} openEditStep={openEditStep} setDeleteStepId={setDeleteStepId} updateStepMut={updateStepMut} buildSubmitData={buildSubmitData} prompts={prompts} />
            ))}
          </div>
        </SortableContext>
      </DndContext>

      {steps.length === 0 && (
        <Card className="text-center py-12">
          <p className="text-sm text-gray-400">暂无步骤，添加第一个处理步骤吧</p>
        </Card>
      )}

      <ConfirmDialog
        open={deleteStepId !== null}
        title="删除步骤"
        message="确定要删除这个处理步骤吗？"
        confirmText="删除"
        loading={deleteStepMut.isPending}
        onConfirm={() => deleteStepId && deleteStepMut.mutate(deleteStepId)}
        onCancel={() => setDeleteStepId(null)}
      />

      {showAddStep && (
        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={() => setShowAddStep(false)}>
          <div className="bg-white rounded-2xl border border-border-soft p-6 w-full max-w-md shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-gray-800 mb-4">添加步骤</h3>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">处理器</label>
                <select
                  value={stepForm.processor_type}
                  onChange={e => setStepForm({ ...stepForm, processor_type: e.target.value })}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
                >
                  {processorTypes.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">提示词模板</label>
                <select
                  value={stepForm.prompt_template_id ?? ''}
                  onChange={e => setStepForm({ ...stepForm, prompt_template_id: e.target.value ? Number(e.target.value) : null })}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
                >
                  <option value="">无（使用默认）</option>
                  {prompts.map((p: any) => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
              {llmProcessors.includes(stepForm.processor_type) && (
                <ProviderModelSelect
                  provider={stepForm.provider}
                  model={stepForm.model}
                  onProviderChange={v => setStepForm({ ...stepForm, provider: v })}
                  onModelChange={v => setStepForm({ ...stepForm, model: v })}
                />
              )}
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">条件</label>
                <input
                  placeholder='例如 type=image'
                  value={stepForm.condition}
                  onChange={e => setStepForm({ ...stepForm, condition: e.target.value })}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1.5">其他配置 (JSON)</label>
                <textarea
                  value={stepForm.config}
                  onChange={e => setStepForm({ ...stepForm, config: e.target.value })}
                  className="w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 h-20 resize-none font-mono"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <Button variant="secondary" onClick={() => setShowAddStep(false)}>取消</Button>
              <Button loading={addStepMut.isPending} onClick={() => addStepMut.mutate({
                ...buildSubmitData(stepForm),
                sort_order: steps.length,
              })}>添加</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
