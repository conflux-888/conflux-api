import { NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  Boxes,
  Gauge,
  LogOut,
  RefreshCw,
  Rocket,
  Newspaper,
} from "lucide-react";
import { clearAuth, getUsername } from "@/lib/auth";
import clsx from "clsx";

const NAV = [
  { to: "/", label: "Dashboard", icon: Gauge, end: true },
  { to: "/seed", label: "Seed Event", icon: Rocket },
  { to: "/sync", label: "GDELT Sync", icon: RefreshCw },
  { to: "/summary", label: "Daily Summary", icon: Newspaper },
];

export function Shell() {
  const navigate = useNavigate();
  const username = getUsername();

  function logout() {
    clearAuth();
    navigate("/login");
  }

  return (
    <div className="flex h-full min-h-screen">
      <aside className="flex w-60 flex-shrink-0 flex-col border-r border-surface-border bg-black/20 p-4">
        <div className="mb-8 flex items-center gap-2 px-2">
          <Boxes className="h-5 w-5 text-critical" />
          <div>
            <div className="text-sm font-bold tracking-tight">CONFLUX</div>
            <div className="text-[10px] uppercase tracking-widest text-ink-muted">
              Admin
            </div>
          </div>
        </div>

        <nav className="flex-1 space-y-1">
          {NAV.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                clsx(
                  "flex items-center gap-2.5 rounded-btn px-3 py-2 text-sm transition",
                  isActive
                    ? "bg-critical/15 text-critical"
                    : "text-ink-muted hover:bg-surface-raised hover:text-ink"
                )
              }
            >
              <Icon className="h-4 w-4" />
              <span className="flex-1">{label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="mt-4 border-t border-surface-border pt-4">
          <div className="mb-2 px-2 text-xs text-ink-muted">{username}</div>
          <button
            onClick={logout}
            className="flex w-full items-center gap-2 rounded-btn px-3 py-2 text-sm text-ink-muted hover:bg-surface-raised hover:text-ink"
          >
            <LogOut className="h-4 w-4" />
            Logout
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto p-6 lg:p-10">
        <Outlet />
      </main>
    </div>
  );
}
