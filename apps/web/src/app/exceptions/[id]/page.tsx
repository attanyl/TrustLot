import Link from "next/link";
import { apiFetch } from "@/lib/api";

type Exception = {
  ID: string;
  ReconRunID: string;
  EntityType: string;
  EntityID: string;
  ReasonCode: string;
  Status: string;
  AssignedTo: string;
  CreatedAt: string;
  UpdatedAt: string;
};

type EvidenceItem = {
  label: string;
  value: string;
};

type FieldDiff = {
  field: string;
  internal_value: string;
  custodian_value: string;
  delta?: string;
};

type SuggestedAction = {
  code: string;
  label: string;
};

type Explanation = {
  exception_id: string;
  recon_result_id: string;
  entity_type: string;
  match_status: string;
  reason_code: string;
  summary: string;
  evidence: EvidenceItem[];
  field_diffs: FieldDiff[];
  suggested_actions: SuggestedAction[];
};

type TrustScoreResult = {
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

async function getException(id: string): Promise<Exception | null> {
  try {
    return await apiFetch<Exception>(`/api/v1/exceptions/${id}`);
  } catch {
    return null;
  }
}

async function getExplanation(id: string): Promise<Explanation | null> {
  try {
    return await apiFetch<Explanation>(`/api/v1/exceptions/${id}/explanation`);
  } catch {
    return null;
  }
}

async function getTrustScore(id: string): Promise<TrustScoreResult | null> {
  try {
    return await apiFetch<TrustScoreResult>(
      `/api/v1/trust-scores/exception/${id}`
    );
  } catch {
    return null;
  }
}

export default async function ExceptionDetailPage({
  params,
}: {
  params: { id: string };
}) {
  const exc = await getException(params.id);
  const explanation = exc ? await getExplanation(params.id) : null;
  const trustScore = exc ? await getTrustScore(params.id) : null;

  return (
    <div>
      <div className="mb-4">
        <Link
          href="/exceptions"
          className="text-xs text-zinc-500 hover:text-zinc-300"
        >
          &larr; Back to exceptions
        </Link>
      </div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        Exception Detail
      </h2>

      {exc ? (
        <>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <Field label="ID" value={exc.ID} mono />
            <Field label="Status" value={exc.Status} />
            <Field label="Reason Code" value={exc.ReasonCode} mono />
            <Field label="Entity Type" value={exc.EntityType} />
            <Field label="Entity ID" value={exc.EntityID} mono />
            <Field label="Recon Run" value={exc.ReconRunID} mono />
            <Field label="Assigned To" value={exc.AssignedTo || "unassigned"} />
            <Field
              label="Created"
              value={new Date(exc.CreatedAt).toLocaleString()}
            />
          </div>

          {trustScore && <TrustScoreCard score={trustScore} />}

          {explanation ? (
            <div className="space-y-4">
              {/* Summary */}
              <Section title="Summary">
                <p className="text-sm text-zinc-200">{explanation.summary}</p>
              </Section>

              {/* Field Diffs */}
              {explanation.field_diffs.length > 0 && (
                <Section title="Field Comparison">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="text-zinc-500 border-b border-zinc-700">
                        <th className="text-left py-1.5 pr-4 font-medium">
                          Field
                        </th>
                        <th className="text-right py-1.5 px-4 font-medium">
                          Internal
                        </th>
                        <th className="text-right py-1.5 px-4 font-medium">
                          Custodian
                        </th>
                        <th className="text-right py-1.5 pl-4 font-medium">
                          Delta
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {explanation.field_diffs.map((diff) => (
                        <tr
                          key={diff.field}
                          className="border-b border-zinc-800"
                        >
                          <td className="py-1.5 pr-4 text-zinc-400 font-mono">
                            {diff.field}
                          </td>
                          <td className="py-1.5 px-4 text-right text-zinc-200 font-mono">
                            {diff.internal_value}
                          </td>
                          <td className="py-1.5 px-4 text-right text-zinc-200 font-mono">
                            {diff.custodian_value}
                          </td>
                          <td className="py-1.5 pl-4 text-right font-mono text-amber-400">
                            {diff.delta || "-"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </Section>
              )}

              {/* Evidence */}
              {explanation.evidence.length > 0 && (
                <Section title="Evidence">
                  <div className="grid grid-cols-2 gap-x-6 gap-y-1.5">
                    {explanation.evidence.map((item) => (
                      <div key={item.label} className="flex justify-between">
                        <span className="text-xs text-zinc-500">
                          {item.label}
                        </span>
                        <span className="text-xs text-zinc-300 font-mono">
                          {item.value}
                        </span>
                      </div>
                    ))}
                  </div>
                </Section>
              )}

              {/* Suggested Actions */}
              {explanation.suggested_actions.length > 0 && (
                <Section title="Suggested Actions">
                  <ul className="space-y-1.5">
                    {explanation.suggested_actions.map((action) => (
                      <li key={action.code} className="flex items-start gap-2">
                        <span className="text-[10px] font-mono bg-zinc-700 text-zinc-400 px-1.5 py-0.5 rounded shrink-0">
                          {action.code}
                        </span>
                        <span className="text-xs text-zinc-300">
                          {action.label}
                        </span>
                      </li>
                    ))}
                  </ul>
                </Section>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4">
              <Section title="Evidence">
                <p className="text-xs text-zinc-600">
                  Explanation unavailable. Ensure the API is running.
                </p>
              </Section>
              <Section title="Lineage">
                <p className="text-xs text-zinc-600">
                  Data lineage trace will appear here in a future milestone.
                </p>
              </Section>
            </div>
          )}
        </>
      ) : (
        <div className="bg-zinc-800 border border-zinc-700 rounded p-4">
          <p className="text-sm text-zinc-400">
            Exception not found. Ensure the API is running and the database is
            seeded.
          </p>
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

function TrustScoreCard({ score }: { score: TrustScoreResult }) {
  const level =
    score.score >= 75 ? "High" : score.score >= 40 ? "Medium" : "Low";
  const color =
    score.score >= 75
      ? "text-emerald-400 border-emerald-500/30"
      : score.score >= 40
        ? "text-amber-400 border-amber-500/30"
        : "text-red-400 border-red-500/30";
  const bgBar =
    score.score >= 75
      ? "bg-emerald-500"
      : score.score >= 40
        ? "bg-amber-500"
        : "bg-red-500";

  const components = [
    { label: "Source Reliability", value: score.components.source_reliability },
    { label: "Reconciliation", value: score.components.reconciliation },
    { label: "Freshness", value: score.components.freshness },
    { label: "Completeness", value: score.components.completeness },
    { label: "Human Review", value: score.components.human_review },
    {
      label: "Anomaly Penalty",
      value: score.components.anomaly_penalty,
      isPenalty: true,
    },
  ];

  return (
    <div className={`bg-zinc-800 border ${color.split(" ")[1]} rounded p-4 mb-4`}>
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Trust Score</h3>
      <div className="flex items-center gap-4 mb-3">
        <span className={`text-2xl font-mono font-bold ${color.split(" ")[0]}`}>
          {score.score.toFixed(1)}
        </span>
        <span
          className={`text-xs font-medium px-2 py-0.5 rounded ${color.split(" ")[0]} bg-zinc-700`}
        >
          {level}
        </span>
        <div className="flex-1 h-2 bg-zinc-700 rounded-full overflow-hidden">
          <div
            className={`h-full ${bgBar} rounded-full`}
            style={{ width: `${Math.min(100, score.score)}%` }}
          />
        </div>
      </div>
      <p className="text-xs text-zinc-400 mb-3">{score.rationale_summary}</p>
      <div className="grid grid-cols-3 gap-2">
        {components.map((c) => (
          <div key={c.label} className="flex justify-between items-center">
            <span className="text-[10px] text-zinc-500">{c.label}</span>
            <span
              className={`text-xs font-mono ${c.isPenalty && c.value > 0 ? "text-red-400" : "text-zinc-300"}`}
            >
              {c.isPenalty && c.value > 0 ? "-" : ""}
              {c.value.toFixed(2)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
