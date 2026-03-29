export default function TrustScoresPage() {
  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        Trust Scores
      </h2>
      <p className="text-xs text-zinc-500 mb-4">
        Trust Score measures automation readiness. Each score is composed of
        source reliability, mapping confidence, reconciliation status,
        freshness, completeness, human review state, and anomaly penalty.
      </p>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-700 text-left text-xs text-zinc-500 uppercase tracking-wide">
            <th className="pb-2 pr-4">Entity</th>
            <th className="pb-2 pr-4">Total</th>
            <th className="pb-2 pr-4">Source</th>
            <th className="pb-2 pr-4">Mapping</th>
            <th className="pb-2 pr-4">Recon</th>
            <th className="pb-2 pr-4">Fresh</th>
            <th className="pb-2 pr-4">Complete</th>
            <th className="pb-2 pr-4">Review</th>
            <th className="pb-2">Anomaly</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td
              colSpan={9}
              className="py-8 text-center text-zinc-600 text-xs"
            >
              No trust scores computed yet.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}
