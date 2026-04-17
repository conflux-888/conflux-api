import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Boxes } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { setAuth } from "@/lib/auth";

export function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const { access_token } = await api.login(username, password);
      setAuth(access_token, username);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-app-gradient p-6">
      <div className="card w-full max-w-sm p-8">
        <div className="mb-6 flex items-center gap-3">
          <Boxes className="h-7 w-7 text-critical" />
          <div>
            <div className="text-lg font-bold tracking-tight">Conflux Admin</div>
            <div className="text-xs uppercase tracking-widest text-ink-muted">
              Operator Console
            </div>
          </div>
        </div>

        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="field-label">Username</label>
            <input
              type="text"
              required
              autoFocus
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="field-input"
              placeholder="admin"
            />
          </div>
          <div>
            <label className="field-label">Password</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="field-input"
              placeholder="••••••••"
            />
          </div>

          {error && (
            <div className="rounded-btn border border-critical/40 bg-critical/10 px-3 py-2 text-xs text-critical">
              {error}
            </div>
          )}

          <button type="submit" disabled={loading} className="btn-primary w-full">
            {loading ? "Signing in…" : "Sign in"}
          </button>
        </form>
      </div>
    </div>
  );
}
