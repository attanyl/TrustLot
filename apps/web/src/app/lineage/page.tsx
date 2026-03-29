export default function LineagePage() {
  return (
    <div>
      <h2 className="text-lg font-semibold text-zinc-200 mb-4">
        Data Lineage
      </h2>
      <div className="bg-zinc-800 border border-zinc-700 rounded p-6">
        <p className="text-sm text-zinc-400 mb-2">
          Trace the path of any record from raw custodian feed through
          normalization, reconciliation, and trust scoring.
        </p>
        <p className="text-xs text-zinc-600">
          Enter an entity type and ID to view the lineage graph. This view will
          be available once ingestion and reconciliation are operational.
        </p>
      </div>
    </div>
  );
}
