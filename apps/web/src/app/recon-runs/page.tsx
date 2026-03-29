import { apiFetch } from "@/lib/api";

type ReconRun = {
  ID: string;
  EntityType: string;
  AsOfDate: string;
  Status: string;
  StartedAt: string;
  CompletedAt: string | null;
  MatchCount: number;
  BreakCount: number;
};

type ReconRunsResponse = {
  runs: ReconRun[];
  total: number;
};

async function getReconRuns(): Promise<ReconRunsResponse> {
  try {
    return await apiFetch<ReconRunsResponse>("/api/v1/recon-runs");
  } catch {
    return { runs: [], total: 0 };
  }
}

export default async function ReconRunsPage() {
  const data = await getReconRuns();

  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        Reconciliation Runs
      </h2>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-700 text-left text-xs text-zinc-500 uppercase tracking-wide">
            <th className="pb-2 pr-4">Run ID</th>
            <th className="pb-2 pr-4">Date</th>
            <th className="pb-2 pr-4">Entity Type</th>
            <th className="pb-2 pr-4">Matches</th>
            <th className="pb-2 pr-4">Breaks</th>
            <th className="pb-2">Status</th>
          </tr>
        </thead>
        <tbody>
          {data.runs.length === 0 ? (
            <tr>
              <td
                colSpan={6}
                className="py-8 text-center text-zinc-600 text-xs"
              >
                No reconciliation runs yet. Run the API server and seed the
                database.
              </td>
            </tr>
          ) : (
            data.runs.map((run) => (
              <tr
                key={run.ID}
                className="border-b border-zinc-800 hover:bg-zinc-800/50"
              >
                <td className="py-2 pr-4 font-mono text-xs text-zinc-400">
                  {run.ID.substring(0, 8)}...
                </td>
                <td className="py-2 pr-4 text-zinc-300">
                  {run.AsOfDate.substring(0, 10)}
                </td>
                <td className="py-2 pr-4 text-zinc-400">{run.EntityType}</td>
                <td className="py-2 pr-4 text-green-400">{run.MatchCount}</td>
                <td className="py-2 pr-4 text-red-400">{run.BreakCount}</td>
                <td className="py-2">
                  <span
                    className={`px-1.5 py-0.5 rounded text-xs ${
                      run.Status === "completed"
                        ? "bg-green-900/50 text-green-300"
                        : run.Status === "running"
                          ? "bg-blue-900/50 text-blue-300"
                          : "bg-red-900/50 text-red-300"
                    }`}
                  >
                    {run.Status}
                  </span>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
