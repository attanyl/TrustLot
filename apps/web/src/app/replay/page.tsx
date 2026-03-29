import Link from "next/link";
import { apiFetch } from "@/lib/api";

type ReplayCase = {
  id: string;
  description: string;
  status: string;
  source_exception_id?: string;
  entity_type: string;
  expected_match_status: string;
  expected_reason_code: string;
  created_at: string;
};

type ReplayCasesResponse = {
  cases: ReplayCase[];
  total: number;
};

async function getReplayCases(): Promise<ReplayCasesResponse> {
  try {
    return await apiFetch<ReplayCasesResponse>("/api/v1/replay-cases");
  } catch {
    return { cases: [], total: 0 };
  }
}

export default async function ReplayPage() {
  const data = await getReplayCases();
  const cases = data.cases ?? [];

  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        Replay Cases
      </h2>
      <p className="text-xs text-zinc-500 mb-4">
        Replay cases are captured scenarios from past reconciliation failures.
        Each case preserves the original data and expected outcome, enabling
        regression testing against new rule versions.
      </p>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-700 text-left text-xs text-zinc-500 uppercase tracking-wide">
            <th className="pb-2 pr-4">Case ID</th>
            <th className="pb-2 pr-4">Entity Type</th>
            <th className="pb-2 pr-4">Expected Reason</th>
            <th className="pb-2 pr-4">Status</th>
            <th className="pb-2">Created</th>
          </tr>
        </thead>
        <tbody>
          {cases.length === 0 ? (
            <tr>
              <td
                colSpan={5}
                className="py-8 text-center text-zinc-600 text-xs"
              >
                No replay cases captured yet.
              </td>
            </tr>
          ) : (
            cases.map((rc) => (
              <tr
                key={rc.id}
                className="border-b border-zinc-800 hover:bg-zinc-800/50"
              >
                <td className="py-2 pr-4 font-mono text-xs">
                  <Link
                    href={`/replay/${rc.id}`}
                    className="text-blue-400 hover:text-blue-300"
                  >
                    {rc.id.slice(0, 12)}...
                  </Link>
                </td>
                <td className="py-2 pr-4 text-zinc-300">{rc.entity_type}</td>
                <td className="py-2 pr-4 font-mono text-xs text-zinc-300">
                  {rc.expected_reason_code || "-"}
                </td>
                <td className="py-2 pr-4">
                  <StatusBadge status={rc.status} />
                </td>
                <td className="py-2 text-zinc-400 text-xs">
                  {new Date(rc.created_at).toLocaleDateString()}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending: "bg-zinc-700 text-zinc-300",
    passing: "bg-green-900/50 text-green-400",
    failing: "bg-red-900/50 text-red-400",
  };
  return (
    <span
      className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${colors[status] || "bg-zinc-700 text-zinc-400"}`}
    >
      {status}
    </span>
  );
}
