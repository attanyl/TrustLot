import { apiFetch } from "@/lib/api";

type ExceptionsResponse = {
  exceptions: Array<{ id: string; status: string }>;
  total: number;
};

type ReconRunsResponse = {
  runs: Array<{ id: string }>;
  total: number;
};

async function getStats() {
  try {
    const [exceptions, runs] = await Promise.all([
      apiFetch<ExceptionsResponse>("/api/v1/exceptions"),
      apiFetch<ReconRunsResponse>("/api/v1/recon-runs"),
    ]);

    const openCount = exceptions.exceptions.filter(
      (e) => e.status === "open"
    ).length;

    return {
      openExceptions: openCount,
      totalExceptions: exceptions.total,
      reconRuns: runs.total,
      connected: true,
    };
  } catch {
    return {
      openExceptions: 0,
      totalExceptions: 0,
      reconRuns: 0,
      connected: false,
    };
  }
}

export default async function DashboardPage() {
  const stats = await getStats();

  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        System Overview
      </h2>
      <div className="grid grid-cols-3 gap-4 mb-8">
        <MetricCard
          label="Open Exceptions"
          value={stats.connected ? String(stats.openExceptions) : "--"}
        />
        <MetricCard
          label="Total Exceptions"
          value={stats.connected ? String(stats.totalExceptions) : "--"}
        />
        <MetricCard
          label="Recon Runs"
          value={stats.connected ? String(stats.reconRuns) : "--"}
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <StatusCard
          title="API Connection"
          status={stats.connected ? "Connected" : "Disconnected — start the API server"}
        />
        <StatusCard
          title="Data"
          status={stats.connected && stats.totalExceptions > 0 ? "Seeded" : "Run make seed to populate"}
        />
      </div>
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-zinc-800 border border-zinc-700 rounded px-4 py-3">
      <div className="text-xs text-zinc-500 uppercase tracking-wide">
        {label}
      </div>
      <div className="text-2xl font-bold text-zinc-200 mt-1">{value}</div>
    </div>
  );
}

function StatusCard({ title, status }: { title: string; status: string }) {
  return (
    <div className="bg-zinc-800 border border-zinc-700 rounded px-4 py-3">
      <div className="text-sm font-medium text-zinc-300">{title}</div>
      <div className="text-xs text-zinc-500 mt-1">{status}</div>
    </div>
  );
}
