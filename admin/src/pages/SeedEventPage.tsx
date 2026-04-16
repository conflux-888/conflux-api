import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Rocket, Trash2, Trash } from "lucide-react";
import { api, ApiError, EventRecord, SeedEventRequest, Severity } from "@/lib/api";
import { Card } from "@/components/ui/Card";
import { SeverityChip } from "@/components/ui/SeverityChip";

const ROOT_CODES = [
  { code: "18", label: "Violent clash" },
  { code: "19", label: "Use of force" },
  { code: "20", label: "Military force" },
];

export function SeedEventPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<SeedEventRequest>({
    title: "Admin test event",
    latitude: 13.7563,
    longitude: 100.5018,
    severity: "critical",
    country: "TH",
    location_name: "Bangkok, Thailand",
    event_type: "Military force",
    event_root_code: "20",
    description: "Synthetic event from admin console",
    num_articles: 20,
  });
  const [result, setResult] = useState<
    | { kind: "idle" }
    | { kind: "loading" }
    | { kind: "success"; eventId: string; externalId: string }
    | { kind: "error"; message: string }
  >({ kind: "idle" });

  const { data: seededEvents, isLoading: listLoading } = useQuery({
    queryKey: ["seeded-events"],
    queryFn: () => api.listSeededEvents({ limit: 20 }),
  });

  const [deleteMsg, setDeleteMsg] = useState<string | null>(null);

  const deleteMutation = useMutation({
    mutationFn: api.deleteSeededEvent,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["seeded-events"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["unread-count"] });
      setDeleteMsg(
        `Deleted 1 event and ${data.notifications_deleted} related notifications`
      );
      setTimeout(() => setDeleteMsg(null), 4000);
    },
  });

  const deleteAllMutation = useMutation({
    mutationFn: api.deleteAllSeededEvents,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["seeded-events"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["unread-count"] });
      setDeleteMsg(
        `Deleted ${data.events_deleted} events and ${data.notifications_deleted} notifications`
      );
      setTimeout(() => setDeleteMsg(null), 4000);
    },
  });

  function update<K extends keyof SeedEventRequest>(key: K, value: SeedEventRequest[K]) {
    setForm((f) => ({ ...f, [key]: value }));
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setResult({ kind: "loading" });
    try {
      const data = await api.seedEvent(form);
      setResult({
        kind: "success",
        eventId: data.event.id,
        externalId: data.event.external_id ?? "",
      });
      queryClient.invalidateQueries({ queryKey: ["seeded-events"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["unread-count"] });
    } catch (err) {
      setResult({
        kind: "error",
        message: err instanceof ApiError ? err.message : "Failed to seed event",
      });
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Seed Test Event</h1>
        <p className="mt-1 text-sm text-ink-muted">
          Creates a synthetic GDELT-style event and dispatches{" "}
          <code className="rounded bg-white/10 px-1.5 py-0.5 text-xs">NotifyNearbyCritical</code>{" "}
          — same pipeline as production sync.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card title="Event parameters" className="lg:col-span-2">
          <form onSubmit={onSubmit} className="space-y-4">
            <div>
              <label className="field-label">Title</label>
              <input
                required
                className="field-input"
                value={form.title}
                onChange={(e) => update("title", e.target.value)}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="field-label">Latitude</label>
                <input
                  type="number"
                  step="0.0001"
                  required
                  className="field-input font-mono"
                  value={form.latitude}
                  onChange={(e) => update("latitude", parseFloat(e.target.value))}
                />
              </div>
              <div>
                <label className="field-label">Longitude</label>
                <input
                  type="number"
                  step="0.0001"
                  required
                  className="field-input font-mono"
                  value={form.longitude}
                  onChange={(e) => update("longitude", parseFloat(e.target.value))}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="field-label">Country (FIPS 2-char)</label>
                <input
                  required
                  maxLength={2}
                  className="field-input uppercase"
                  value={form.country}
                  onChange={(e) => update("country", e.target.value.toUpperCase())}
                />
              </div>
              <div>
                <label className="field-label">Location name</label>
                <input
                  className="field-input"
                  value={form.location_name}
                  onChange={(e) => update("location_name", e.target.value)}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="field-label">Severity</label>
                <div className="flex gap-2">
                  {(["critical", "high", "medium", "low"] as Severity[]).map((s) => (
                    <button
                      key={s}
                      type="button"
                      onClick={() => update("severity", s)}
                      className={
                        form.severity === s
                          ? "rounded-full border border-critical/60 bg-critical/20 px-3 py-1 text-xs font-bold uppercase text-critical"
                          : "rounded-full border border-surface-border px-3 py-1 text-xs font-semibold uppercase text-ink-muted hover:text-ink"
                      }
                    >
                      {s}
                    </button>
                  ))}
                </div>
              </div>
              <div>
                <label className="field-label">CAMEO root code</label>
                <select
                  className="field-input"
                  value={form.event_root_code}
                  onChange={(e) => update("event_root_code", e.target.value)}
                >
                  {ROOT_CODES.map((rc) => (
                    <option key={rc.code} value={rc.code}>
                      {rc.code} — {rc.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className="field-label">Description</label>
              <textarea
                rows={2}
                className="field-input"
                value={form.description}
                onChange={(e) => update("description", e.target.value)}
              />
            </div>

            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                type="submit"
                disabled={result.kind === "loading"}
                className="btn-primary"
              >
                <Rocket className="mr-2 h-4 w-4" />
                {result.kind === "loading" ? "Seeding…" : "Seed Event"}
              </button>
            </div>
          </form>
        </Card>

        <Card title="Last submission">
          {result.kind === "idle" && (
            <div className="text-xs text-ink-muted">
              Submit the form to inject an event and trigger the notification pipeline.
            </div>
          )}
          {result.kind === "loading" && (
            <div className="text-xs text-ink-muted">Dispatching…</div>
          )}
          {result.kind === "error" && (
            <div className="rounded-btn border border-critical/40 bg-critical/10 px-3 py-2 text-xs text-critical">
              {result.message}
            </div>
          )}
          {result.kind === "success" && (
            <div className="space-y-2 text-xs">
              <div className="flex items-center gap-2">
                <SeverityChip severity={form.severity} />
                <span className="text-ink-muted">queued</span>
              </div>
              <div>
                <div className="text-ink-muted">Event ID</div>
                <div className="font-mono">{result.eventId}</div>
              </div>
              <div>
                <div className="text-ink-muted">External ID</div>
                <div className="font-mono break-all">{result.externalId}</div>
              </div>
              <div className="pt-2 text-ink-muted">
                Check the Notifications tab — a new critical_nearby entry should appear within seconds.
              </div>
            </div>
          )}
        </Card>
      </div>

      {deleteMsg && (
        <div className="rounded-card border border-low/40 bg-low/10 px-4 py-3 text-sm text-low">
          {deleteMsg}
        </div>
      )}

      <Card
        title="Seeded events"
        subtitle={
          seededEvents
            ? `${seededEvents.length} shown · deleting also removes related notifications`
            : undefined
        }
        action={
          seededEvents && seededEvents.length > 0 ? (
            <button
              onClick={() => {
                if (confirm(`Delete all ${seededEvents.length} seeded events and their notifications?`)) {
                  deleteAllMutation.mutate();
                }
              }}
              disabled={deleteAllMutation.isPending}
              className="inline-flex items-center gap-1.5 rounded-btn border border-critical/40 bg-critical/10 px-3 py-1.5 text-xs font-semibold text-critical hover:bg-critical/20 disabled:opacity-50"
            >
              <Trash className="h-3.5 w-3.5" />
              {deleteAllMutation.isPending ? "Deleting…" : "Delete all"}
            </button>
          ) : undefined
        }
      >
        {listLoading && <div className="text-xs text-ink-muted">Loading…</div>}
        {!listLoading && (!seededEvents || seededEvents.length === 0) && (
          <div className="py-6 text-center text-xs text-ink-muted">
            No seeded events yet.
          </div>
        )}
        <div className="space-y-2">
          {seededEvents?.map((e) => (
            <SeededEventRow
              key={e.id}
              event={e}
              onDelete={() => deleteMutation.mutate(e.id)}
              deleting={deleteMutation.isPending && deleteMutation.variables === e.id}
            />
          ))}
        </div>
      </Card>
    </div>
  );
}

function SeededEventRow({
  event,
  onDelete,
  deleting,
}: {
  event: EventRecord;
  onDelete: () => void;
  deleting: boolean;
}) {
  const [lng, lat] = event.location?.coordinates ?? [0, 0];
  return (
    <div className="flex items-start gap-4 rounded-btn border border-surface-border bg-white/5 p-3">
      <div className="flex-1 min-w-0">
        <div className="mb-1 flex items-center gap-2">
          <SeverityChip severity={event.severity} />
          <span className="truncate text-sm font-semibold">{event.title}</span>
        </div>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-ink-muted">
          <span>{event.country}</span>
          <span>·</span>
          <span className="font-mono">
            {lat.toFixed(4)}, {lng.toFixed(4)}
          </span>
          <span>·</span>
          <span>{formatDistanceToNow(new Date(event.created_at), { addSuffix: true })}</span>
        </div>
        <div className="mt-1 truncate font-mono text-[10px] text-ink-faint">
          {event.external_id}
        </div>
      </div>
      <button
        onClick={() => {
          if (confirm(`Delete "${event.title}" and its notifications?`)) onDelete();
        }}
        disabled={deleting}
        className="rounded-btn border border-critical/30 bg-critical/10 p-2 text-critical transition hover:bg-critical/20 disabled:opacity-50"
        title="Delete seeded event"
      >
        <Trash2 className="h-4 w-4" />
      </button>
    </div>
  );
}
