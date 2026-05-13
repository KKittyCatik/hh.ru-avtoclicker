import * as Dialog from '@radix-ui/react-dialog';
import { AnimatePresence, motion } from 'framer-motion';
import { Command, Play, Rocket, Square, UserSquare2 } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';

import { Header } from '@/components/layout/Header';
import { PageContainer } from '@/components/layout/PageContainer';
import { Sidebar } from '@/components/layout/Sidebar';
import { NotificationCenter } from '@/components/notifications/NotificationCenter';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { useApply } from '@/hooks/useApply';
import { useRealtime } from '@/hooks/useRealtime';
import { cn } from '@/lib/utils';
import { useNotificationsStore } from '@/store/notifications';

const titles: Record<string, string> = {
  '/': 'Dashboard',
  '/negotiations': 'Negotiations',
  '/accounts': 'Accounts',
  '/settings': 'Settings',
};

export function AppLayout() {
  useRealtime();
  const location = useLocation();
  const navigate = useNavigate();
  const setNotificationsOpen = useNotificationsStore((state) => state.setOpen);
  const unread = useNotificationsStore((state) => state.items.filter((item) => !item.read).length);
  const { startMutation, stopMutation, publishMutation } = useApply();
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [query, setQuery] = useState('');

  useEffect(() => {
    const listener = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setPaletteOpen((current) => !current);
      }
    };

    window.addEventListener('keydown', listener);
    return () => window.removeEventListener('keydown', listener);
  }, []);

  const actions = useMemo(
    () => [
      { label: 'Open dashboard', icon: Rocket, run: () => navigate('/') },
      { label: 'Open negotiations', icon: Command, run: () => navigate('/negotiations') },
      { label: 'Open accounts', icon: UserSquare2, run: () => navigate('/accounts') },
      { label: 'Start apply worker', icon: Play, run: () => void startMutation.mutateAsync() },
      { label: 'Stop worker', icon: Square, run: () => void stopMutation.mutateAsync() },
      { label: 'Publish resume', icon: Rocket, run: () => void publishMutation.mutateAsync() },
    ],
    [navigate, publishMutation, startMutation, stopMutation],
  );

  const filteredActions = actions.filter((action) => action.label.toLowerCase().includes(query.toLowerCase()));

  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <Sidebar />
      <div className="min-w-0 flex-1">
        <Header
          title={titles[location.pathname] ?? 'hh.ru Copilot'}
          onOpenPalette={() => setPaletteOpen(true)}
          onOpenNotifications={() => setNotificationsOpen(true)}
        />
        <PageContainer className="pt-6">
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, y: 18 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -12 }}
              transition={{ duration: 0.22, ease: 'easeOut' }}
            >
              <Outlet />
            </motion.div>
          </AnimatePresence>
        </PageContainer>
      </div>
      <NotificationCenter />
      <Dialog.Root open={paletteOpen} onOpenChange={setPaletteOpen}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm" />
          <Dialog.Content className="fixed left-1/2 top-[18%] z-50 w-[min(640px,calc(100vw-2rem))] -translate-x-1/2 rounded-3xl border border-white/10 bg-[#0d0d11]/95 p-4 shadow-soft">
            <div className="mb-4 flex items-center justify-between px-2">
              <div>
                <Dialog.Title className="text-lg font-semibold text-foreground">Command palette</Dialog.Title>
                <Dialog.Description className="text-sm text-muted">Глобальный поиск и быстрые действия</Dialog.Description>
              </div>
              <StatusBadge label={`${unread} unread`} variant="accent" />
            </div>
            <input
              autoFocus
              className="mb-3 w-full rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-3 text-sm text-foreground placeholder:text-muted"
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search commands..."
              value={query}
            />
            <div className="space-y-2">
              {filteredActions.map((action) => (
                <button
                  key={action.label}
                  className={cn('flex w-full items-center gap-3 rounded-2xl px-4 py-3 text-left text-sm text-muted transition hover:bg-white/5 hover:text-foreground')}
                  onClick={() => {
                    action.run();
                    setPaletteOpen(false);
                  }}
                  type="button"
                >
                  <action.icon className="h-4 w-4" />
                  {action.label}
                </button>
              ))}
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
}
