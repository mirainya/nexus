import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ToastProvider } from './components/Toast';
import MainLayout from './layouts/MainLayout';
import LoginPage from './pages/login';
import DashboardPage from './pages/dashboard';
import PipelinesPage from './pages/pipelines';
import PipelineDetailPage from './pages/pipelines/detail';
import PromptsPage from './pages/prompts';
import JobsPage from './pages/jobs';
import ReviewsPage from './pages/reviews';
import EntitiesPage from './pages/entities';
import SettingsPage from './pages/settings';
import PlaygroundPage from './pages/playground';
import SearchPage from './pages/search';
import GraphPage from './pages/graph';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
});

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000 < Date.now();
  } catch {
    return true;
  }
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token');
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem('token');
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<ProtectedRoute><MainLayout /></ProtectedRoute>}>
              <Route index element={<DashboardPage />} />
              <Route path="pipelines" element={<PipelinesPage />} />
              <Route path="pipelines/:id" element={<PipelineDetailPage />} />
              <Route path="prompts" element={<PromptsPage />} />
              <Route path="jobs" element={<JobsPage />} />
              <Route path="reviews" element={<ReviewsPage />} />
              <Route path="entities" element={<EntitiesPage />} />
              <Route path="graph" element={<GraphPage />} />
              <Route path="settings" element={<SettingsPage />} />
              <Route path="playground" element={<PlaygroundPage />} />
              <Route path="search" element={<SearchPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ToastProvider>
    </QueryClientProvider>
  );
}
