import { Suspense, lazy } from 'react';
import { createBrowserRouter, isRouteErrorResponse, useRouteError } from 'react-router-dom';

import { AppLayout } from '@/components/layout/AppLayout';
import { LoadingScreen } from '@/components/ui/LoadingScreen';

const DashboardPage = lazy(() => import('@/pages/Dashboard'));
const NegotiationsPage = lazy(() => import('@/pages/Negotiations'));
const AccountsPage = lazy(() => import('@/pages/Accounts'));
const SettingsPage = lazy(() => import('@/pages/Settings'));

function withSuspense(node: React.ReactNode) {
  return <Suspense fallback={<LoadingScreen />}>{node}</Suspense>;
}

function RouteErrorBoundary() {
  const error = useRouteError();
  const message = isRouteErrorResponse(error)
    ? `${error.status} ${error.statusText}`
    : error instanceof Error
      ? error.message
      : 'Unknown route error';

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6 text-center">
      <div className="glass-panel max-w-lg rounded-3xl p-10">
        <p className="mb-3 text-sm uppercase tracking-[0.24em] text-muted">Route error</p>
        <h1 className="mb-4 text-3xl font-semibold text-foreground">Что-то пошло не так</h1>
        <p className="text-sm text-muted">{message}</p>
      </div>
    </div>
  );
}

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    errorElement: <RouteErrorBoundary />,
    children: [
      { index: true, element: withSuspense(<DashboardPage />) },
      { path: 'negotiations', element: withSuspense(<NegotiationsPage />) },
      { path: 'accounts', element: withSuspense(<AccountsPage />) },
      { path: 'settings', element: withSuspense(<SettingsPage />) },
    ],
  },
]);
