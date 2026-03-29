import { apiFetch } from "@/lib/api";

type TrustScore = {
  id: string;
  entity_type: string;
  entity_id: string;
  score: number;
  components: {
    source_reliability: number;
    reconciliation: number;
    freshness: number;
    completeness: number;
    human_review: number;
    anomaly_penalty: number;
  };
  rationale_summary: string;
  computed_at: string;
};

type TrustScoresResponse = {
  scores: TrustScore[];
  total: number;
};

async function getTrustScores(): Promise<TrustScore[]> {
  try {
    const data = await apiFetch<TrustScoresResponse>(
      "/api/v1/trust-scores?limit=100"
    );
    return data.scores ?? [];
  } catch {
    return [];
  }
}

function scoreBadge(score: number) {
  if (score >= 75) return { label: "High", color: "text-emerald-400" };
  if (score >= 40) return { label: "Medium", color: "text-amber-400" };
  return { label: "Low", color: "text-red-400" };
}

function scoreBar(score: number) {
  const pct = Math.min(100, Math.max(0, score));
  let bg = "bg-red-500";
  if (score >= 75) bg = "bg-emerald-500";
  else if (score >= 40) bg = "bg-amber-500";
  return (
    <div className="w-20 h-1.5 bg-zinc-700 rounded-full overflow-hidden">
      <div className={`h-full ${bg} rounded-full`} style={{ width: `${pct}%` }} />
    </div>
  );
}

function fmt(v: number) {
  return v.toFixed(2);
}

export default async function TrustScoresPage() {
  const scores = await getTrustScores();

  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-1">
        Trust Scores
      </h2>
      <p className="text-xs text-zinc-500 mb-4">
        Automation readiness scores computed from reconciliation state, data
        freshness, completeness, and review status. Each score is deterministic
        and explainable.
      </p>

      {scores.length === 0 ? (
        <div className="bg-zinc-800 border border-zinc-700 rounded p-6 text-center">
          <p className="text-sm text-zinc-500">
            No trust scores computed yet. Trigger a score computation via the
            API or exception detail page.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-700 text-left text-xs text-zinc-500 uppercase tracking-wide">
                <th className="pb-2 pr-4">Entity</th>
                <th className="pb-2 pr-4">Score</th>
                <th className="pb-2 pr-4">Level</th>
                <th className="pb-2 pr-4">Source</th>
                <th className="pb-2 pr-4">Recon</th>
                <th className="pb-2 pr-4">Fresh</th>
                <th className="pb-2 pr-4">Complete</th>
                <th className="pb-2 pr-4">Review</th>
                <th className="pb-2 pr-4">Penalty</th>
                <th className="pb-2">Rationale</th>
              </tr>
            </thead>
            <tbody>
              {scores.map((s) => {
                const badge = scoreBadge(s.score);
                return (
                  <tr
                    key={s.id}
                    className="border-b border-zinc-800 hover:bg-zinc-800/50"
                  >
                    <td className="py-2 pr-4">
                      <span className="text-[10px] font-mono bg-zinc-700 text-zinc-400 px-1 py-0.5 rounded mr-1.5">
                        {s.entity_type}
                      </span>
                      <span className="text-xs font-mono text-zinc-300">
                        {s.entity_id.slice(0, 8)}
                      </span>
                    </td>
                    <td className="py-2 pr-4">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-mono text-zinc-200 w-10 text-right">
                          {fmt(s.score)}
                        </span>
                        {scoreBar(s.score)}
                      </div>
                    </td>
                    <td className={`py-2 pr-4 text-xs font-medium ${badge.color}`}>
                      {badge.label}
                    </td>
                    <td className="py-2 pr-4 text-xs font-mono text-zinc-400">
                      {fmt(s.components.source_reliability)}
                    </td>
                    <td className="py-2 pr-4 text-xs font-mono text-zinc-400">
                      {fmt(s.components.reconciliation)}
                    </td>
                    <td className="py-2 pr-4 text-xs font-mono text-zinc-400">
                      {fmt(s.components.freshness)}
                    </td>
                    <td className="py-2 pr-4 text-xs font-mono text-zinc-400">
                      {fmt(s.components.completeness)}
                    </td>
                    <td className="py-2 pr-4 text-xs font-mono text-zinc-400">
                      {fmt(s.components.human_review)}
                    </td>
                    <td className="py-2 pr-4 text-xs font-mono text-zinc-400">
                      {fmt(s.components.anomaly_penalty)}
                    </td>
                    <td className="py-2 text-xs text-zinc-400 max-w-xs truncate">
                      {s.rationale_summary}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
