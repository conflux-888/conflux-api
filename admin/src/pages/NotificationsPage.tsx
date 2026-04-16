import { useQuery, useQueryClient } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { BellRing, Check } from "lucide-react";
import { api, Notification } from "@/lib/api";
import { Card } from "@/components/ui/Card";

export function NotificationsPage() {
  const queryClient = useQueryClient();

  const { data: notifs, isLoading } = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.listNotifications({ limit: 50 }),
    refetchInterval: 3000,
  });

  async function markRead(id: string) {
    await api.markRead(id);
    queryClient.invalidateQueries({ queryKey: ["notifications"] });
    queryClient.invalidateQueries({ queryKey: ["unread-count"] });
  }

  async function markAllRead() {
    await api.markAllRead();
    queryClient.invalidateQueries({ queryKey: ["notifications"] });
    queryClient.invalidateQueries({ queryKey: ["unread-count"] });
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Notifications</h1>
          <p className="mt-1 text-sm text-ink-muted">
            Live stream for this admin account. Polls every 3s.
          </p>
        </div>
        <button onClick={markAllRead} className="btn-ghost">
          <Check className="mr-2 h-4 w-4" /> Mark all read
        </button>
      </div>

      {isLoading && (
        <Card>
          <div className="text-xs text-ink-muted">Loading…</div>
        </Card>
      )}

      {!isLoading && (!notifs || notifs.length === 0) && (
        <Card>
          <div className="flex flex-col items-center gap-3 py-10 text-ink-muted">
            <BellRing className="h-8 w-8 opacity-40" />
            <div className="text-sm">No notifications yet.</div>
            <div className="text-xs">
              Use the Seed Event page to inject an event and trigger one.
            </div>
          </div>
        </Card>
      )}

      <div className="space-y-2">
        {notifs?.map((n) => (
          <NotificationRow key={n.id} n={n} onRead={() => markRead(n.id)} />
        ))}
      </div>
    </div>
  );
}

function NotificationRow({ n, onRead }: { n: Notification; onRead: () => void }) {
  const unread = !n.read_at;
  const typeColor =
    n.type === "critical_nearby" ? "text-critical" : "text-blue-400";

  return (
    <div
      className={
        unread
          ? "card flex items-start gap-4 p-4 ring-1 ring-critical/40"
          : "card flex items-start gap-4 p-4 opacity-70"
      }
    >
      <div className={`mt-0.5 ${typeColor}`}>
        <BellRing className="h-4 w-4" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="mb-0.5 flex items-center gap-2">
          <span className="text-[10px] uppercase tracking-widest text-ink-muted">
            {n.type}
          </span>
          {n.distance_km != null && (
            <span className="text-[10px] text-ink-faint">· {n.distance_km.toFixed(1)} km</span>
          )}
          {unread && <span className="h-1.5 w-1.5 rounded-full bg-critical" />}
        </div>
        <div className="text-sm font-semibold">{n.title}</div>
        <div className="mt-0.5 truncate text-xs text-ink-muted">{n.body}</div>
        <div className="mt-1 text-[10px] text-ink-faint">
          {formatDistanceToNow(new Date(n.created_at), { addSuffix: true })}
        </div>
      </div>
      {unread && (
        <button
          onClick={onRead}
          className="rounded-btn border border-surface-border px-2 py-1 text-[10px] uppercase tracking-wider text-ink-muted hover:text-ink"
        >
          Mark read
        </button>
      )}
    </div>
  );
}
