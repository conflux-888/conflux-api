import { useEffect } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { isAuthenticated } from "@/lib/auth";

export function ProtectedRoute({ children }: { children: JSX.Element }) {
  const navigate = useNavigate();

  useEffect(() => {
    const onExpired = () => navigate("/login", { replace: true });
    window.addEventListener("auth:expired", onExpired);
    return () => window.removeEventListener("auth:expired", onExpired);
  }, [navigate]);

  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }
  return children;
}
