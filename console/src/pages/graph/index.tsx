import { useEffect, useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import Graph from 'graphology';
import { SigmaContainer, useLoadGraph, useRegisterEvents, useSigma, ControlsContainer, ZoomControl, FullScreenControl } from '@react-sigma/core';
import forceAtlas2 from 'graphology-layout-forceatlas2';
import '@react-sigma/core/lib/style.css';
import { PageHeader, Card, Loading } from '../../components/UI';
import { graphApi } from '../../api';
import type { GraphData } from '../../api/types';

const TYPE_COLORS: Record<string, string> = {
  person: '#8b5cf6',
  organization: '#f472b6',
  location: '#34d399',
  event: '#fbbf24',
  product: '#60a5fa',
  concept: '#f87171',
  technology: '#a78bfa',
  date: '#fb923c',
};

function getColor(type: string): string {
  return TYPE_COLORS[type.toLowerCase()] ?? '#94a3b8';
}

function GraphLoader({ data }: { data: GraphData }) {
  const loadGraph = useLoadGraph();

  useEffect(() => {
    const graph = new Graph();
    data.nodes.forEach((node) => {
      graph.addNode(String(node.id), {
        label: node.label,
        size: 8 + node.confidence * 8,
        color: getColor(node.type),
        x: Math.random() * 100,
        y: Math.random() * 100,
        nodeType: node.type,
      });
    });
    data.edges.forEach((edge, i) => {
      const src = String(edge.source);
      const tgt = String(edge.target);
      if (graph.hasNode(src) && graph.hasNode(tgt)) {
        graph.addEdge(src, tgt, { label: edge.type, size: 1, key: `e-${i}` });
      }
    });
    forceAtlas2.assign(graph, { iterations: 100, settings: { gravity: 1, scalingRatio: 2 } });
    loadGraph(graph);
  }, [data, loadGraph]);

  return null;
}

function GraphEvents({ onClickNode }: { onClickNode: (id: string) => void }) {
  const registerEvents = useRegisterEvents();
  const sigma = useSigma();

  useEffect(() => {
    registerEvents({
      clickNode: (event) => onClickNode(event.node),
      enterNode: () => { document.body.style.cursor = 'pointer'; },
      leaveNode: () => { document.body.style.cursor = 'default'; },
    });
    return () => { document.body.style.cursor = 'default'; };
  }, [registerEvents, sigma, onClickNode]);

  return null;
}

function NodeDetail({ nodeId, data }: { nodeId: string | null; data: GraphData }) {
  if (!nodeId) return null;
  const node = data.nodes.find((n) => String(n.id) === nodeId);
  if (!node) return null;
  const relations = data.edges.filter((e) => String(e.source) === nodeId || String(e.target) === nodeId);

  return (
    <div className="absolute top-4 right-4 w-72 bg-white rounded-2xl border border-border-soft shadow-lg p-4 z-10">
      <div className="flex items-center gap-2 mb-3">
        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: getColor(node.type) }} />
        <span className="font-medium text-gray-800">{node.label}</span>
      </div>
      <div className="text-xs text-gray-500 space-y-1">
        <p>类型: {node.type}</p>
        <p>置信度: {(node.confidence * 100).toFixed(0)}%</p>
        <p>关系数: {relations.length}</p>
      </div>
      {relations.length > 0 && (
        <div className="mt-3 border-t border-border-soft pt-2">
          <p className="text-xs text-gray-400 mb-1">关系</p>
          <div className="space-y-1 max-h-40 overflow-y-auto">
            {relations.slice(0, 10).map((r, i) => {
              const otherId = String(r.source) === nodeId ? String(r.target) : String(r.source);
              const other = data.nodes.find((n) => String(n.id) === otherId);
              return (
                <div key={i} className="text-xs text-gray-600 flex items-center gap-1">
                  <span className="text-gray-400">{r.type}</span>
                  <span>→</span>
                  <span>{other?.label ?? otherId}</span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

function Legend({ types }: { types: string[] }) {
  return (
    <div className="absolute bottom-4 left-4 bg-white/90 rounded-xl border border-border-soft px-3 py-2 z-10">
      <div className="flex flex-wrap gap-3">
        {types.map((t) => (
          <div key={t} className="flex items-center gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: getColor(t) }} />
            <span className="text-xs text-gray-500">{t}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function GraphPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['graph'],
    queryFn: () => graphApi.getData(200),
  });
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  const graphData = (data as any)?.data as GraphData | undefined;
  const types = [...new Set(graphData?.nodes.map((n) => n.type) ?? [])];

  const handleClickNode = useCallback((id: string) => {
    setSelectedNode((prev) => (prev === id ? null : id));
  }, []);

  if (isLoading) return <Loading />;

  if (!graphData || graphData.nodes.length === 0) {
    return (
      <div>
        <PageHeader title="知识图谱" description="实体关系可视化" />
        <Card>
          <p className="text-sm text-gray-400 py-8 text-center">暂无图谱数据</p>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <PageHeader title="知识图谱" description={`${graphData.nodes.length} 个实体，${graphData.edges.length} 条关系`} />
      <Card className="relative p-0 overflow-hidden" style={{ height: 'calc(100vh - 180px)' }}>
        <SigmaContainer
          style={{ width: '100%', height: '100%' }}
          settings={{
            defaultEdgeColor: '#e2e8f0',
            defaultEdgeType: 'arrow',
            labelRenderedSizeThreshold: 8,
            labelFont: 'system-ui, sans-serif',
            labelSize: 12,
            labelColor: { color: '#374151' },
          }}
        >
          <GraphLoader data={graphData} />
          <GraphEvents onClickNode={handleClickNode} />
          <ControlsContainer position="bottom-right">
            <ZoomControl />
            <FullScreenControl />
          </ControlsContainer>
        </SigmaContainer>
        <Legend types={types} />
        <NodeDetail nodeId={selectedNode} data={graphData} />
      </Card>
    </div>
  );
}
