import { ReactNode } from "react";
import clsx from "clsx";

export function Card({
  children,
  className,
  title,
  subtitle,
  action,
}: {
  children: ReactNode;
  className?: string;
  title?: string;
  subtitle?: string;
  action?: ReactNode;
}) {
  return (
    <div className={clsx("card p-5", className)}>
      {(title || action) && (
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            {title && <h3 className="text-sm font-semibold tracking-tight">{title}</h3>}
            {subtitle && <p className="mt-0.5 text-xs text-ink-muted">{subtitle}</p>}
          </div>
          {action}
        </div>
      )}
      {children}
    </div>
  );
}
