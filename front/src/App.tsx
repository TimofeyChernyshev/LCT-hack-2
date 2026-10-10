import { Navigate, Route, Routes } from "react-router-dom";
import { Shell } from "./components/Shell";
import { ConfirmPage, Home, LoginPage, RegisterPage, RequireRole } from "./pages/AuthPages";
import { CandidateHome } from "./pages/CandidateHome";
import { OnboardingPage } from "./pages/OnboardingPage";
import { TestPage } from "./pages/TestPage";
import { CandidateInvitations } from "./pages/CandidateInvitations";
import { VacanciesPage } from "./pages/VacanciesPage";
import { EmployerHome } from "./pages/EmployerHome";
import { CatalogPage } from "./pages/CatalogPage";
import { EmployerInvitations } from "./pages/EmployerInvitations";
import { EmployerTasksPage } from "./pages/EmployerTasksPage";
import { SettingsPage } from "./pages/SettingsPage";

export function App() {
  return (
    <Routes>
      <Route element={<Shell />}>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/confirm-email" element={<ConfirmPage />} />
        <Route path="/candidate" element={<RequireRole role="candidate"><CandidateHome /></RequireRole>} />
        <Route path="/candidate/onboarding" element={<RequireRole role="candidate"><OnboardingPage /></RequireRole>} />
        <Route path="/candidate/test" element={<RequireRole role="candidate"><TestPage /></RequireRole>} />
        <Route path="/candidate/invitations" element={<RequireRole role="candidate"><CandidateInvitations /></RequireRole>} />
        <Route path="/candidate/vacancies" element={<RequireRole role="candidate"><VacanciesPage /></RequireRole>} />
        <Route path="/candidate/settings" element={<RequireRole role="candidate"><SettingsPage /></RequireRole>} />
        <Route path="/employer" element={<RequireRole role="employer"><EmployerHome /></RequireRole>} />
        <Route path="/employer/candidates" element={<RequireRole role="employer"><CatalogPage /></RequireRole>} />
        <Route path="/employer/invitations" element={<RequireRole role="employer"><EmployerInvitations /></RequireRole>} />
        <Route path="/employer/tasks" element={<RequireRole role="employer"><EmployerTasksPage /></RequireRole>} />
        <Route path="/employer/settings" element={<RequireRole role="employer"><SettingsPage /></RequireRole>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
