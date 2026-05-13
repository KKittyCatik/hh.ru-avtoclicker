import { BarChart3, BriefcaseBusiness, ChevronLeft, MessagesSquare, Settings } from 'lucide-react';
import { NavLink } from 'react-router-dom';

import { StatusBadge } from '@/components/ui/StatusBadge';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/store/auth';
import { useWorkerStore } from '@/store/worker';

const items = [
  { to: '/', label: 'Dashboard', icon: BarChart3 },
  { to: '/negotiations', label: 'Negotiations', icon: MessagesSquare },
  { to: '/accounts', label: 'Accounts', icon: BriefcaseBusiness },
  { to: '/settings', label: 'Settings', icon: Settings },
];

export function Sidebar() {
  const collapsed = useWorkerStore((state) => state.sidebarCollapsed);
  const toggleSidebar = useWorkerStore((state) => state.toggleSidebar);
  const websocketState = useWorkerStore((state) => state.websocketState);
  const isRunning = useWorkerStore((state) => state.isRunning);
  const settings = useAuthStore((state) => state.settings);

  return (
    <aside className={cn('sticky top-0 hidden h-screen shrink-0 border-r border-white/6 bg-black/20 px-3 py-4 backdrop-blur-xl lg:block', collapsed ? 'w-[92px]' : 'w-[280px]')}>
      <div className="mb-8 flex items-center justify-between gap-2 px-2">
        <div className={cn('overflow-hidden transition-all', collapsed && 'w-0 opacity-0')}>
          <p className="text-xs uppercase tracking-[0.24em] text-muted">hh.ru</p>
          <p className="text-lg font-semibold text-gradient">Copilot</p>
        </div>
        <button type="button" onClick={toggleSidebar} className="glass-panel rounded-2xl p-2 text-muted transition hover:text-foreground">
          <ChevronLeft className={cn('h-4 w-4 transition', collapsed && 'rotate-180')} />
        </button>
      </div>
      <nav className="space-y-2">
        {items.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              cn(
                'group flex items-center gap-3 rounded-2xl px-3 py-3 text-sm text-muted transition hover:bg-white/5 hover:text-foreground',
                isActive && 'bg-gradient-to-r from-violet-600/20 to-cyan-500/10 text-foreground shadow-glow',
              )
            }
          >
            <Icon className="h-4 w-4 shrink-0" />
            {!collapsed ? <span>{label}</span> : null}
          </NavLink>
        ))}
      </nav>
      <div className="mt-auto space-y-3 px-2 pb-2 pt-10">
        <div className="glass-panel rounded-3xl p-4">
          <div className="mb-3 flex items-center justify-between">
            <span className="text-xs uppercase tracking-[0.24em] text-muted">Realtime</span>
            <StatusBadge
              label={websocketState}
              variant={websocketState === 'connected' ? 'success' : websocketState === 'reconnecting' ? 'warning' : 'danger'}
            />
          </div>
          {!collapsed ? (
            <div className="space-y-2 text-sm text-muted">
              <p>Worker: <span className="text-foreground">{isRunning ? 'running' : 'idle'}</span></p>
              <p>Daily limit: <span className="text-foreground">{settings.dailyLimit}</span></p>
            </div>
          ) : null}
        </div>
      </div>
    </aside>
  );
}
