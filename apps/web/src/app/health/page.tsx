import { apiFetch } from "@/lib/api";

type HealthResponse = {
  status: string;
  database: string;
};

async function getHealth(): Promise<HealthResponse | null> {
  try {
    return await apiFetch<HealthResponse>("/healthz");
  } catch {
    return null;
  }
}

export default async function HealthPage() {
  const health = await getHealth();

  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        System Health
      </h2>
      <div className="space-y-3">
        <HealthRow
          service="API Server"
          endpoint="/healthz"
          status={health ? "ok" : "unreachable"}
        />
        <HealthRow
          service="PostgreSQL"
          endpoint="localhost:5432"
          status={health?.database ?? "unknown"}
        />
        <HealthRow
          service="Redis"
          endpoint="localhost:6379"
          status="not checked"
        />
        <HealthRow
          service="Worker"
          endpoint="--"
          status="not checked"
        />
      </div>
    </div>
  );
}

function HealthRow({
  service,
  endpoint,
  status,
}: {
  service: string;
  endpoint: string;
  status: string;
}) {
  const color =
    status === "ok" || status === "connected"
      ? "text-green-400"
      : status === "error" || status === "unreachable"
        ? "text-red-400"
        : "text-zinc-500";

  return (
    <div className="flex items-center justify-between bg-zinc-800 border border-zinc-700 rounded px-4 py-3">
      <div>
        <div className="text-sm text-zinc-200">{service}</div>
        <div className="text-xs text-zinc-500 font-mono">{endpoint}</div>
      </div>
      <div className={`text-xs uppercase font-medium ${color}`}>{status}</div>
    </div>
  );
}
