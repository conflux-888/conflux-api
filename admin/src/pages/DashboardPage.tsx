import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Newspaper, RefreshCw, Rocket } from "lucide-react";
import { api } from "@/lib/api";
import { Card } from "@/components/ui/Card";

export function DashboardPage() {
  const { data: sync } = useQuery({
    queryKey: ["sync-status"],
    queryFn: api.syncStatus,
    refetchInterval: 10000,
  });

  const { data: seeded } = useQuery({
    queryKey: ["seeded-events", { limit: 1 }],
    queryFn: () => api.listSeededEvents({ limit: 1 }),
  });

  const lastSyncAt = sync?.last_sync_at
    ? new Date(sync.last_sync_at).toLocaleString()
    : "—";

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="mt-1 text-sm text-ink-muted">
          Operator console for testing notification delivery end-to-end.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <Card title="GDELT Sync" subtitle={sync?.status ?? "—"}>
          <div className="text-2xl font-bold">{sync?.events_synced ?? 0}</div>
          <div className="text-xs text-ink-muted">events in last cycle</div>
        </Card>

        <Card title="Last Sync" subtitle="Most recent run">
          <div className="font-mono text-sm">{lastSyncAt}</div>
        </Card>

        <Card title="Seeded Events" subtitle="Admin test events">
          <div className="text-2xl font-bold">{seeded?.length ?? 0}+</div>
          <div className="text-xs text-ink-muted">active in DB</div>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <ActionCard
          to="/seed"
          icon={<Rocket className="h-5 w-5" />}
          title="Seed Test Event"
          desc="Inject a synthetic critical event to test the critical_nearby notification flow."
        />
        <ActionCard
          to="/summary"
          icon={<Newspaper className="h-5 w-5" />}
          title="Generate Daily Summary"
          desc="Trigger LLM summary generation for a date to test the daily_briefing broadcast."
        />
        <ActionCard
          to="/sync"
          icon={<RefreshCw className="h-5 w-5" />}
          title="Run GDELT Sync"
          desc="Manually run the GDELT fetcher. Useful to seed real conflict events."
        />
      </div>
    </div>
  );
}

function ActionCard({
  to,
  icon,
  title,
  desc,
}: {
  to: string;
  icon: React.ReactNode;
  title: string;
  desc: string;
}) {
  return (
    <Link to={to} className="card block p-5 transition hover:bg-surface-raised">
      <div className="flex items-start gap-3">
        <div className="rounded-btn bg-critical/15 p-2 text-critical">{icon}</div>
        <div className="flex-1">
          <div className="text-sm font-semibold">{title}</div>
          <div className="mt-1 text-xs text-ink-muted">{desc}</div>
        </div>
      </div>
    </Link>
  );
}
