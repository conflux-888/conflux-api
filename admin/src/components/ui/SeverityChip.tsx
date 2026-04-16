import { Severity } from "@/lib/api";

const COLORS: Record<Severity, string> = {
  critical: "#FF0000",
  high: "#FF8C00",
  medium: "#E6C000",
  low: "#00FF00",
};

export function SeverityChip({ severity }: { severity: Severity }) {
  const color = COLORS[severity];
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider"
      style={{
        color,
        backgroundColor: `${color}26`,
        borderColor: `${color}66`,
      }}
    >
      <span
        className="h-1.5 w-1.5 rounded-full"
        style={{ backgroundColor: color }}
      />
      {severity}
    </span>
  );
}
