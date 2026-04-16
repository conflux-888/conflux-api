import { clearAuth, getToken } from "./auth";

const API_BASE = import.meta.env.DEV
  ? (import.meta.env.VITE_API_BASE ?? "http://localhost:8080/api/v1")
  : "/api/v1";

export class ApiError extends Error {
  code: string;
  status: number;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

type Options = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  body?: unknown;
  query?: Record<string, string | number | boolean | undefined>;
  auth?: boolean;
};

async function request<T>(path: string, opts: Options = {}): Promise<T> {
  const { method = "GET", body, query, auth = true } = opts;

  let url = API_BASE + path;
  if (query) {
    const qs = new URLSearchParams();
    for (const [k, v] of Object.entries(query)) {
      if (v !== undefined && v !== null && v !== "") qs.set(k, String(v));
    }
    const q = qs.toString();
    if (q) url += `?${q}`;
  }

  const headers: Record<string, string> = {};
  if (body) headers["Content-Type"] = "application/json";
  if (auth) {
    const token = getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    clearAuth();
    window.dispatchEvent(new Event("auth:expired"));
  }

  const text = await res.text();
  const json = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const code = json?.error?.code ?? "HTTP_ERROR";
    const message = json?.error?.message ?? res.statusText;
    throw new ApiError(res.status, code, message);
  }

  return json?.data as T;
}

export const api = {
  // auth
  login: (email: string, password: string) =>
    request<{ access_token: string; expires_in: number }>("/auth/login", {
      method: "POST",
      body: { email, password },
      auth: false,
    }),
  me: () =>
    request<{ id: string; email: string; display_name: string }>("/users/me"),

  // admin: events
  seedEvent: (body: SeedEventRequest) =>
    request<{ event: EventRecord; notification_dispatch: string }>(
      "/admin/events/seed",
      { method: "POST", body }
    ),
  listSeededEvents: (opts?: { page?: number; limit?: number }) =>
    request<EventRecord[]>("/admin/events/seeded", {
      query: { page: opts?.page, limit: opts?.limit },
    }),
  deleteSeededEvent: (id: string) =>
    request<{ message: string; notifications_deleted: number }>(
      `/admin/events/${id}`,
      { method: "DELETE" }
    ),
  deleteAllSeededEvents: () =>
    request<{ events_deleted: number; notifications_deleted: number }>(
      "/admin/events/seeded",
      { method: "DELETE" }
    ),

  // admin: sync
  syncStatus: () => request<SyncState>("/admin/sync/status"),
  triggerSync: () =>
    request<SyncState>("/admin/sync/trigger", { method: "POST" }),

  // admin: summary
  triggerSummary: (date: string) =>
    request<DailySummary>("/admin/summaries/trigger", {
      method: "POST",
      body: { date },
    }),
  getSummary: (date: string) => request<DailySummary>(`/summaries/${date}`),

  // notifications
  listNotifications: (opts?: { unreadOnly?: boolean; page?: number; limit?: number }) =>
    request<Notification[]>("/notifications/me", {
      query: {
        unread_only: opts?.unreadOnly,
        page: opts?.page,
        limit: opts?.limit,
      },
    }),
  unreadCount: () =>
    request<{ unread_count: number }>("/notifications/me/unread-count"),
  markRead: (id: string) =>
    request<{ message: string }>(`/notifications/${id}/read`, { method: "POST" }),
  markAllRead: () =>
    request<{ modified_count: number }>("/notifications/read-all", { method: "POST" }),

  // preferences
  getPreferences: () => request<UserPreferences>("/preferences"),
  updatePreferences: (body: Partial<UserPreferences>) =>
    request<UserPreferences>("/preferences", { method: "PUT", body }),
  updateLocation: (latitude: number, longitude: number) =>
    request<{ message: string }>("/preferences/location", {
      method: "PUT",
      body: { latitude, longitude },
    }),
};

// Types
export type Severity = "critical" | "high" | "medium" | "low";

export type SeedEventRequest = {
  title: string;
  latitude: number;
  longitude: number;
  severity: Severity;
  country: string;
  location_name?: string;
  event_type?: string;
  event_root_code?: string;
  description?: string;
  num_articles?: number;
};

export type EventRecord = {
  id: string;
  source: string;
  external_id?: string;
  event_type: string;
  severity: Severity;
  title: string;
  country: string;
  location_name: string;
  location: { type: string; coordinates: [number, number] };
  event_date: string;
  created_at: string;
};

export type SyncState = {
  id: string;
  last_sync_timestamp: string;
  last_sync_at: string;
  status: string;
  events_synced: number;
  error_message?: string;
};

export type DailySummary = {
  id: string;
  summary_date: string;
  status: string;
  event_count: number;
  title: string;
  content: string;
  severity_breakdown: { critical: number; high: number; medium: number; low: number };
  generated_at: string;
};

export type Notification = {
  id: string;
  type: "critical_nearby" | "daily_briefing";
  title: string;
  body: string;
  event_id?: string;
  summary_date?: string;
  distance_km?: number;
  read_at?: string | null;
  created_at: string;
};

export type UserPreferences = {
  notifications_enabled: boolean;
  min_severity: Severity;
  radius_km: number;
  last_location?: { type: string; coordinates: [number, number] } | null;
  last_location_at?: string | null;
};
