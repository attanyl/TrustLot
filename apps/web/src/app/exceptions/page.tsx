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

type ExceptionsResponse = {
  exceptions: Exception[];
  total: number;
};

async function getExceptions(): Promise<ExceptionsResponse> {
  try {
    return await apiFetch<ExceptionsResponse>("/api/v1/exceptions");
  } catch {
    return { exceptions: [], total: 0 };
  }
}

export default async function ExceptionsPage() {
  const data = await getExceptions();
  const openCount = data.exceptions.filter((e) => e.Status === "open").length;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-zinc-200">
          Exception Inbox
        </h2>
        <div className="text-xs text-zinc-500">{openCount} open</div>
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-700 text-left text-xs text-zinc-500 uppercase tracking-wide">
            <th className="pb-2 pr-4">ID</th>
            <th className="pb-2 pr-4">Reason Code</th>
            <th className="pb-2 pr-4">Entity Type</th>
            <th className="pb-2 pr-4">Status</th>
            <th className="pb-2">Created</th>
          </tr>
        </thead>
        <tbody>
          {data.exceptions.length === 0 ? (
            <tr>
              <td
                colSpan={5}
                className="py-8 text-center text-zinc-600 text-xs"
              >
                No exceptions. Run the API server and seed the database.
              </td>
            </tr>
          ) : (
            data.exceptions.map((exc) => (
              <tr
                key={exc.ID}
                className="border-b border-zinc-800 hover:bg-zinc-800/50"
              >
                <td className="py-2 pr-4 font-mono text-xs text-zinc-400">
                  <Link
                    href={`/exceptions/${exc.ID}`}
                    className="hover:text-zinc-200"
                  >
                    {exc.ID.substring(0, 8)}...
                  </Link>
                </td>
                <td className="py-2 pr-4">
                  <span className="px-1.5 py-0.5 bg-zinc-700 rounded text-xs font-mono">
                    {exc.ReasonCode}
                  </span>
                </td>
                <td className="py-2 pr-4 text-zinc-400">{exc.EntityType}</td>
                <td className="py-2 pr-4">
                  <StatusBadge status={exc.Status} />
                </td>
                <td className="py-2 text-zinc-500 text-xs">
                  {new Date(exc.CreatedAt).toLocaleDateString()}
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
    open: "bg-red-900/50 text-red-300",
    acknowledged: "bg-yellow-900/50 text-yellow-300",
    resolved: "bg-green-900/50 text-green-300",
    suppressed: "bg-zinc-700 text-zinc-400",
  };

  return (
    <span
      className={`px-1.5 py-0.5 rounded text-xs ${colors[status] ?? "bg-zinc-700 text-zinc-400"}`}
    >
      {status}
    </span>
  );
}
