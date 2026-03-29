"use client";

import { useState } from "react";

type LineageNode = {
  id: string;
  type: string;
  label: string;
  metadata?: Record<string, unknown>;
};

type LineageEdge = {
  from_id: string;
  to_id: string;
  edge_type: string;
};

type LineageGraph = {
  root_type: string;
  root_id: string;
  nodes: LineageNode[];
  edges: LineageEdge[];
};

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const nodeTypeColors: Record<string, string> = {
  ingestion_batch: "border-blue-500/40 bg-blue-500/5",
  raw_record: "border-blue-500/30 bg-blue-500/5",
  position: "border-cyan-500/40 bg-cyan-500/5",
  transaction: "border-cyan-500/40 bg-cyan-500/5",
  recon_result: "border-amber-500/40 bg-amber-500/5",
  exception: "border-red-500/40 bg-red-500/5",
  trust_score: "border-emerald-500/40 bg-emerald-500/5",
};

const nodeTypeLabels: Record<string, string> = {
  ingestion_batch: "BATCH",
  raw_record: "RAW",
  position: "POS",
  transaction: "TXN",
  recon_result: "RECON",
  exception: "EXCPT",
  trust_score: "TRUST",
};

export default function LineagePage() {
  const [entityType, setEntityType] = useState("position");
  const [entityID, setEntityID] = useState("");
  const [graph, setGraph] = useState<LineageGraph | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!entityID.trim()) return;

    setLoading(true);
    setError("");
    setGraph(null);

    try {
      const res = await fetch(
        `${API_BASE}/api/v1/lineage/${entityType}/${entityID.trim()}`,
        { cache: "no-store" }
      );
      if (!res.ok) {
        const data = await res.json();
        setError(data.error || `Error ${res.status}`);
        return;
      }
      const data: LineageGraph = await res.json();
      setGraph(data);
    } catch (err) {
      setError("Failed to connect to API");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-1">
        Data Lineage
      </h2>
      <p className="text-xs text-zinc-500 mb-4">
        Trace any record from raw source through normalization,
        reconciliation, exceptions, and trust scoring.
      </p>

      <form onSubmit={handleSearch} className="flex gap-2 mb-4">
        <select
          value={entityType}
          onChange={(e) => setEntityType(e.target.value)}
          className="bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm text-zinc-200"
        >
          <option value="position">Position</option>
          <option value="transaction">Transaction</option>
        </select>
        <input
          type="text"
          value={entityID}
          onChange={(e) => setEntityID(e.target.value)}
          placeholder="Entity ID (UUID)"
          className="flex-1 bg-zinc-800 border border-zinc-700 rounded px-3 py-1.5 text-sm text-zinc-200 font-mono placeholder:text-zinc-600"
        />
        <button
          type="submit"
          disabled={loading}
          className="bg-zinc-700 hover:bg-zinc-600 text-zinc-200 px-4 py-1.5 rounded text-sm disabled:opacity-50"
        >
          {loading ? "Loading..." : "Trace"}
        </button>
      </form>

      {error && (
        <div className="bg-red-900/20 border border-red-500/30 rounded p-3 mb-4">
          <p className="text-sm text-red-400">{error}</p>
        </div>
      )}

      {graph && <LineageGraphView graph={graph} />}
    </div>
  );
}

function LineageGraphView({ graph }: { graph: LineageGraph }) {
  if (graph.nodes.length === 0) {
    return (
      <div className="bg-zinc-800 border border-zinc-700 rounded p-6 text-center">
        <p className="text-sm text-zinc-500">
          No lineage data found for this entity.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="text-xs text-zinc-500 mb-2">
        {graph.nodes.length} nodes, {graph.edges.length} edges
      </div>

      {graph.nodes.map((node) => (
        <div
          key={node.id}
          className={`border rounded p-3 ${nodeTypeColors[node.type] || "border-zinc-700 bg-zinc-800"}`}
        >
          <div className="flex items-center gap-2 mb-1">
            <span className="text-[10px] font-mono font-bold text-zinc-500 bg-zinc-700/50 px-1.5 py-0.5 rounded">
              {nodeTypeLabels[node.type] || node.type.toUpperCase()}
            </span>
            <span className="text-sm text-zinc-200">{node.label}</span>
          </div>
          <div className="text-[10px] font-mono text-zinc-600 mb-1.5">
            {node.id}
          </div>
          {node.metadata && Object.keys(node.metadata).length > 0 && (
            <div className="grid grid-cols-2 gap-x-4 gap-y-0.5">
              {Object.entries(node.metadata).map(([k, v]) => (
                <div key={k} className="flex justify-between">
                  <span className="text-[10px] text-zinc-600">{k}</span>
                  <span className="text-[10px] text-zinc-400 font-mono truncate ml-2">
                    {String(v)}
                  </span>
                </div>
              ))}
            </div>
          )}

          {/* Show outgoing edges */}
          {graph.edges
            .filter((e) => e.from_id === node.id)
            .map((e) => {
              const target = graph.nodes.find((n) => n.id === e.to_id);
              return (
                <div
                  key={`${e.from_id}-${e.to_id}-${e.edge_type}`}
                  className="mt-1.5 flex items-center gap-1.5 text-[10px] text-zinc-500"
                >
                  <span className="text-zinc-600">&#8627;</span>
                  <span className="text-zinc-500">{e.edge_type}</span>
                  <span className="text-zinc-600">→</span>
                  <span className="font-mono text-zinc-400">
                    {target
                      ? `${nodeTypeLabels[target.type] || target.type} ${target.id.slice(0, 8)}`
                      : e.to_id.slice(0, 8)}
                  </span>
                </div>
              );
            })}
        </div>
      ))}
    </div>
  );
}
