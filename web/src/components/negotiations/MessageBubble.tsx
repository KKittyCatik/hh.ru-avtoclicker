import { Bot, UserRound } from 'lucide-react';

import { StatusBadge } from '@/components/ui/StatusBadge';
import { cn } from '@/lib/utils';
import { formatDateTime } from '@/lib/format';
import type { NegotiationMessage } from '@/types/negotiation';

export function MessageBubble({ message }: { message: NegotiationMessage }) {
  const fromHR = message.from.toLowerCase() === 'hr';

  return (
    <div className={cn('flex', fromHR ? 'justify-start' : 'justify-end')}>
      <div className={cn('max-w-[80%] rounded-3xl border px-4 py-3', fromHR ? 'border-white/10 bg-white/[0.04]' : 'border-violet-500/30 bg-violet-500/12')}>
        <div className="mb-2 flex items-center gap-2 text-xs text-muted">
          {fromHR ? <Bot className="h-3.5 w-3.5" /> : <UserRound className="h-3.5 w-3.5" />}
          <span>{fromHR ? 'HR' : 'You'}</span>
          <span>{formatDateTime(message.created_at)}</span>
          {message.potential_bot ? <StatusBadge label="bot detected" variant="info" /> : null}
          {message.needs_human_input ? <StatusBadge label="needs input" variant="warning" /> : null}
        </div>
        <p className="whitespace-pre-wrap text-sm leading-6 text-foreground">{message.text || 'Quick reply option'}</p>
      </div>
    </div>
  );
}
