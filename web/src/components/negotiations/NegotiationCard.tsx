import { motion } from 'framer-motion';
import * as Avatar from '@radix-ui/react-avatar';

import { StatusBadge } from '@/components/ui/StatusBadge';
import { cn } from '@/lib/utils';
import { formatRelativeTime } from '@/lib/format';
import type { Negotiation } from '@/types/negotiation';

export function NegotiationCard({ item, selected, onClick }: { item: Negotiation; selected: boolean; onClick: () => void }) {
  const initial = item.vacancy_id.slice(0, 2).toUpperCase();

  return (
    <motion.button
      layout
      type="button"
      onClick={onClick}
      className={cn(
        'w-full rounded-3xl border p-4 text-left transition',
        selected ? 'border-violet-500/30 bg-violet-500/10 shadow-glow' : 'border-white/6 bg-white/[0.03] hover:border-white/10 hover:bg-white/[0.05]',
      )}
    >
      <div className="mb-3 flex items-start gap-3">
        <Avatar.Root className="inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-600/80 to-cyan-500/80 text-sm font-semibold text-white">
          <Avatar.Fallback>{initial}</Avatar.Fallback>
        </Avatar.Root>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <p className="truncate text-sm font-medium text-foreground">Vacancy #{item.vacancy_id}</p>
            {item.needs_reply ? <StatusBadge label="needs input" variant="warning" /> : null}
          </div>
          <p className="mt-1 line-clamp-2 text-sm text-muted">{item.last_message.text || 'Нет текста сообщения'}</p>
        </div>
      </div>
      <div className="flex items-center justify-between gap-2 text-xs text-muted">
        <StatusBadge label={item.status} variant={item.is_bot ? 'info' : 'neutral'} />
        <span>{formatRelativeTime(item.updated_at)}</span>
      </div>
    </motion.button>
  );
}
