import { Navigate, Route, Routes } from "react-router-dom";
import { Shell } from "./components/layout/Shell";
import { ProtectedRoute } from "./routes/ProtectedRoute";
import { LoginPage } from "./pages/LoginPage";
import { DashboardPage } from "./pages/DashboardPage";
import { SeedEventPage } from "./pages/SeedEventPage";
import { SyncTriggerPage } from "./pages/SyncTriggerPage";
import { SummaryTriggerPage } from "./pages/SummaryTriggerPage";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        element={
          <ProtectedRoute>
            <Shell />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<DashboardPage />} />
        <Route path="/seed" element={<SeedEventPage />} />
        <Route path="/sync" element={<SyncTriggerPage />} />
        <Route path="/summary" element={<SummaryTriggerPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
