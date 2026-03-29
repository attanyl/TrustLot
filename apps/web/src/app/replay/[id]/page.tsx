"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

type ReplayCaseDetail = {
  id: string;
  description: string;
  status: string;
  source_exception_id?: string;
  entity_type: string;
  expected_match_status: string;
  expected_reason_code: string;
  created_at: string;
  internal_snapshot: Record<string, unknown> | null;
  custodian_snapshot: Record<string, unknown> | null;
  config_snapshot: Record<string, unknown> | null;
};

type ReplayRunResult = {
  case_id: string;
  actual_match_status: string;
  actual_reason_code: string;
  expected_match_status: string;
  expected_reason_code: string;
  changed: boolean;
  regression: boolean;
  improvement: boolean;
  summary: string;
};

async function fetchCase(id: string): Promise<ReplayCaseDetail | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/replay-cases/${id}`, {
      cache: "no-store",
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

async function runReplay(id: string): Promise<ReplayRunResult | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/replay-cases/${id}/run`, {
      method: "POST",
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export default function ReplayCaseDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [detail, setDetail] = useState<ReplayCaseDetail | null>(null);
  const [result, setResult] = useState<ReplayRunResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [running, setRunning] = useState(false);

  useEffect(() => {
    fetchCase(id).then((d) => {
      setDetail(d);
      setLoading(false);
    });
  }, [id]);

  const handleRun = async () => {
    setRunning(true);
    const r = await runReplay(id);
    setResult(r);
    setRunning(false);
    // Refresh detail to pick up status change.
    const d = await fetchCase(id);
    if (d) setDetail(d);
  };

  if (loading) {
    return (
      <div className="text-sm text-zinc-500">Loading replay case...</div>
    );
  }

  if (!detail) {
    return (
      <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
        <p className="text-sm text-zinc-400">
          Replay case not found. Ensure the API is running.
        </p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-4">
        <Link
          href="/replay"
          className="text-xs text-zinc-500 hover:text-zinc-300"
        >
          &larr; Back to replay cases
        </Link>
      </div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-zinc-200">
          Replay Case Detail
        </h2>
        <button
          onClick={handleRun}
          disabled={running}
          className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 disabled:bg-zinc-700 text-white text-xs font-medium rounded"
        >
          {running ? "Running..." : "Run Replay"}
        </button>
      </div>

      {/* Case metadata */}
      <div className="grid grid-cols-2 gap-4 mb-4">
        <Field label="ID" value={detail.id} mono />
        <Field label="Status" value={detail.status} />
        <Field label="Entity Type" value={detail.entity_type} />
        <Field
          label="Expected Outcome"
          value={`${detail.expected_match_status}${detail.expected_reason_code ? " / " + detail.expected_reason_code : ""}`}
          mono
        />
        <Field
          label="Source Exception"
          value={detail.source_exception_id || "none"}
          mono
        />
        <Field
          label="Created"
          value={new Date(detail.created_at).toLocaleString()}
        />
      </div>

      {/* Replay result */}
      {result && (
        <Section title="Replay Result">
          <div className="space-y-2">
            <div className="flex items-center gap-2 mb-2">
              <ResultBadge result={result} />
              <span className="text-sm text-zinc-200">{result.summary}</span>
            </div>
            <div className="grid grid-cols-2 gap-4 text-xs">
              <div>
                <span className="text-zinc-500">Expected: </span>
                <span className="font-mono text-zinc-300">
                  {result.expected_match_status}
                  {result.expected_reason_code
                    ? ` / ${result.expected_reason_code}`
                    : ""}
                </span>
              </div>
              <div>
                <span className="text-zinc-500">Actual: </span>
                <span className="font-mono text-zinc-300">
                  {result.actual_match_status}
                  {result.actual_reason_code
                    ? ` / ${result.actual_reason_code}`
                    : ""}
                </span>
              </div>
            </div>
          </div>
        </Section>
      )}

      {/* Snapshots */}
      <div className="grid grid-cols-2 gap-4 mt-4">
        <Section title="Internal Snapshot">
          <SnapshotView data={detail.internal_snapshot} />
        </Section>
        <Section title="Custodian Snapshot">
          <SnapshotView data={detail.custodian_snapshot} />
        </Section>
      </div>

      {detail.config_snapshot && (
        <div className="mt-4">
          <Section title="Config">
            <SnapshotView data={detail.config_snapshot} />
          </Section>
        </div>
      )}
    </div>
  );
}

function Field({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2">
      <div className="text-xs text-zinc-500 mb-0.5">{label}</div>
      <div className={`text-sm text-zinc-200 ${mono ? "font-mono" : ""}`}>
        {value}
      </div>
    </div>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
      <h3 className="text-sm font-medium text-zinc-400 mb-2">{title}</h3>
      {children}
    </div>
  );
}

function SnapshotView({
  data,
}: {
  data: Record<string, unknown> | null;
}) {
  if (!data) {
    return (
      <p className="text-xs text-zinc-600">No snapshot data available.</p>
    );
  }
  return (
    <div className="space-y-1">
      {Object.entries(data).map(([key, val]) => (
        <div key={key} className="flex justify-between text-xs">
          <span className="text-zinc-500">{key}</span>
          <span className="font-mono text-zinc-300">
            {typeof val === "object" ? JSON.stringify(val) : String(val ?? "")}
          </span>
        </div>
      ))}
    </div>
  );
}

function ResultBadge({ result }: { result: ReplayRunResult }) {
  if (!result.changed) {
    return (
      <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-zinc-700 text-zinc-300">
        UNCHANGED
      </span>
    );
  }
  if (result.improvement) {
    return (
      <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-green-900/50 text-green-400">
        IMPROVED
      </span>
    );
  }
  if (result.regression) {
    return (
      <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-900/50 text-red-400">
        REGRESSED
      </span>
    );
  }
  return (
    <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-900/50 text-amber-400">
      CHANGED
    </span>
  );
}
