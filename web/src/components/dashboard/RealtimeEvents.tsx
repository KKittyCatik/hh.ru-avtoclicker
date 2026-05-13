import { Activity, Wifi } from 'lucide-react';

import { GlowCard } from '@/components/ui/GlowCard';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { useWorkerStore } from '@/store/worker';
import { formatRelativeTime } from '@/lib/format';

export function RealtimeEvents() {
  const websocketState = useWorkerStore((state) => state.websocketState);
  const lastEventAt = useWorkerStore((state) => state.lastEventAt);

  return (
    <GlowCard>
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity className="h-4 w-4 text-cyan-300" />
          <h2 className="text-lg font-semibold text-foreground">Realtime</h2>
        </div>
        <StatusBadge label={websocketState} variant={websocketState === 'connected' ? 'success' : 'warning'} />
      </div>
      <div className="space-y-3 text-sm text-muted">
        <div className="flex items-center justify-between rounded-2xl border border-white/6 bg-white/[0.02] p-3">
          <span className="flex items-center gap-2"><Wifi className="h-4 w-4 text-violet-300" /> WebSocket</span>
          <span className="text-foreground">{websocketState}</span>
        </div>
        <div className="rounded-2xl border border-white/6 bg-white/[0.02] p-3">
          <p className="mb-1 text-xs uppercase tracking-[0.24em] text-muted">Last signal</p>
          <p className="text-foreground">{lastEventAt ? formatRelativeTime(lastEventAt) : 'Ожидаем первое событие'}</p>
        </div>
      </div>
    </GlowCard>
  );
}
