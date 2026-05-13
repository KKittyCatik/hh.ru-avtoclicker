import { BellRing } from 'lucide-react';

import { GlowCard } from '@/components/ui/GlowCard';
import { useNotificationsStore } from '@/store/notifications';
import { formatRelativeTime } from '@/lib/format';

export function ActivityFeed() {
  const items = useNotificationsStore((state) => state.items.slice(0, 8));

  return (
    <GlowCard className="h-full">
      <div className="mb-4 flex items-center gap-2">
        <BellRing className="h-4 w-4 text-violet-300" />
        <h2 className="text-lg font-semibold text-foreground">Activity feed</h2>
      </div>
      <div className="space-y-3">
        {items.length ? (
          items.map((item) => (
            <div key={item.id} className="rounded-2xl border border-white/6 bg-white/[0.02] p-3">
              <p className="text-sm font-medium text-foreground">{item.title}</p>
              <p className="mt-1 text-sm text-muted">{item.description}</p>
              <p className="mt-2 text-xs text-muted">{formatRelativeTime(item.createdAt)}</p>
            </div>
          ))
        ) : (
          <p className="text-sm text-muted">Realtime события появятся здесь сразу после первых откликов и ответов HR.</p>
        )}
      </div>
    </GlowCard>
  );
}
