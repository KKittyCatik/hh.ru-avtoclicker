import { BellRing } from 'lucide-react';

import { formatRelativeTime } from '@/lib/format';
import type { NotificationItem } from '@/store/notifications';
import { StatusBadge } from '@/components/ui/StatusBadge';

export function NotificationToast({ item }: { item: NotificationItem }) {
  return (
    <div className="glass-panel rounded-3xl p-4">
      <div className="mb-2 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 text-sm font-medium text-foreground">
          <BellRing className="h-4 w-4 text-violet-300" />
          {item.title}
        </div>
        {!item.read ? <StatusBadge label="new" variant="accent" /> : null}
      </div>
      <p className="mb-2 text-sm text-muted">{item.description}</p>
      <p className="text-xs text-muted">{formatRelativeTime(item.createdAt)}</p>
    </div>
  );
}
