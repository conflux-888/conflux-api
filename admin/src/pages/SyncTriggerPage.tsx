import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { Card } from "@/components/ui/Card";

export function SyncTriggerPage() {
  const queryClient = useQueryClient();
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: status } = useQuery({
    queryKey: ["sync-status"],
    queryFn: api.syncStatus,
    refetchInterval: 5000,
  });

  async function trigger() {
    setRunning(true);
    setError(null);
    try {
      await api.triggerSync();
      queryClient.invalidateQueries({ queryKey: ["sync-status"] });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Sync failed");
    } finally {
      setRunning(false);
    }
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">GDELT Sync</h1>
        <p className="mt-1 text-sm text-ink-muted">
          Fetches the latest GDELT export and upserts conflict events. New critical events
          trigger nearby notifications.
        </p>
      </div>

      <Card title="Current state" subtitle={`Status: ${status?.status ?? "—"}`}>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <Metric label="Events synced" value={status?.events_synced ?? 0} />
          <Metric
            label="Last sync timestamp"
            value={status?.last_sync_timestamp || "—"}
            mono
          />
          <Metric
            label="Last sync at"
            value={status?.last_sync_at ? new Date(status.last_sync_at).toLocaleString() : "—"}
          />
          <Metric label="Error" value={status?.error_message || "—"} />
        </div>
      </Card>

      <Card title="Trigger manually">
        <p className="mb-4 text-xs text-ink-muted">
          Runs the same code path as the background scheduler. Blocks until the fetch completes.
        </p>
        <button onClick={trigger} disabled={running} className="btn-primary">
          <RefreshCw className={running ? "mr-2 h-4 w-4 animate-spin" : "mr-2 h-4 w-4"} />
          {running ? "Running…" : "Run sync now"}
        </button>
        {error && (
          <div className="mt-3 rounded-btn border border-critical/40 bg-critical/10 px-3 py-2 text-xs text-critical">
            {error}
          </div>
        )}
      </Card>
    </div>
  );
}

function Metric({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
  return (
    <div>
      <div className="text-xs text-ink-muted">{label}</div>
      <div className={mono ? "mt-1 font-mono text-sm" : "mt-1 text-sm"}>{value}</div>
    </div>
  );
}
