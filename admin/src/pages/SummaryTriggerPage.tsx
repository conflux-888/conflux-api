import { useState } from "react";
import { Newspaper } from "lucide-react";
import { api, ApiError, DailySummary } from "@/lib/api";
import { Card } from "@/components/ui/Card";

function yesterdayUTC(): string {
  const d = new Date();
  d.setUTCDate(d.getUTCDate() - 1);
  return d.toISOString().slice(0, 10);
}

export function SummaryTriggerPage() {
  const [date, setDate] = useState(yesterdayUTC());
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<DailySummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function trigger() {
    setRunning(true);
    setError(null);
    try {
      const data = await api.triggerSummary(date);
      setResult(data);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to generate summary");
    } finally {
      setRunning(false);
    }
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Daily Summary</h1>
        <p className="mt-1 text-sm text-ink-muted">
          Generates an LLM briefing for the chosen UTC day. First generation broadcasts a{" "}
          <code className="rounded bg-white/10 px-1.5 py-0.5 text-xs">daily_briefing</code>{" "}
          notification to all enabled users.
        </p>
      </div>

      <Card title="Trigger">
        <div className="flex flex-wrap items-end gap-4">
          <div>
            <label className="field-label">Date (UTC)</label>
            <input
              type="date"
              className="field-input font-mono"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          </div>
          <button onClick={trigger} disabled={running} className="btn-primary">
            <Newspaper className="mr-2 h-4 w-4" />
            {running ? "Generating…" : "Generate summary"}
          </button>
        </div>

        {error && (
          <div className="mt-4 rounded-btn border border-critical/40 bg-critical/10 px-3 py-2 text-xs text-critical">
            {error}
          </div>
        )}
      </Card>

      {result && (
        <Card title={result.title} subtitle={`${result.summary_date} · ${result.status}`}>
          <div className="mb-4 grid grid-cols-4 gap-2 text-center">
            {(["critical", "high", "medium", "low"] as const).map((s) => (
              <div key={s} className="rounded-btn bg-white/5 p-3">
                <div className="text-[10px] uppercase tracking-wide text-ink-muted">{s}</div>
                <div className="mt-1 text-lg font-bold">
                  {result.severity_breakdown[s]}
                </div>
              </div>
            ))}
          </div>
          <div className="text-xs text-ink-muted">
            {result.event_count} events · generated{" "}
            {new Date(result.generated_at).toLocaleString()}
          </div>
          <div className="mt-4 whitespace-pre-wrap text-sm leading-relaxed text-ink">
            {result.content}
          </div>
        </Card>
      )}
    </div>
  );
}
