import { ExternalLink, Wifi } from 'lucide-react';

import { ResumePublishButton } from '@/components/accounts/ResumePublishButton';
import { GlowCard } from '@/components/ui/GlowCard';
import { StatusBadge } from '@/components/ui/StatusBadge';
import { useWorkerStore } from '@/store/worker';
import type { Account } from '@/types/account';

export function AccountCard({ account }: { account: Account }) {
  const websocketState = useWorkerStore((state) => state.websocketState);
  const tokenExpired = account.token.expires_at ? new Date(account.token.expires_at).getTime() < Date.now() : true;

  return (
    <GlowCard>
      <div className="mb-5 flex items-start justify-between gap-3">
        <div>
          <h3 className="text-lg font-semibold text-foreground">{account.name}</h3>
          <p className="text-sm text-muted">{account.resume_ids.length} resumes · {account.search_urls.length} search URLs</p>
        </div>
        <StatusBadge label={account.needs_reauth ? 'reauth required' : tokenExpired ? 'token expired' : 'active'} variant={account.needs_reauth ? 'danger' : tokenExpired ? 'warning' : 'success'} />
      </div>
      <div className="grid gap-3 text-sm text-muted md:grid-cols-2">
        <div className="rounded-2xl border border-white/6 bg-white/[0.03] p-3">
          <p className="mb-1 text-xs uppercase tracking-[0.24em]">Apply limit</p>
          <p className="text-foreground">{account.preferences.min_salary || 0} ₽ min salary</p>
        </div>
        <div className="rounded-2xl border border-white/6 bg-white/[0.03] p-3">
          <p className="mb-1 text-xs uppercase tracking-[0.24em]">WebSocket</p>
          <p className="flex items-center gap-2 text-foreground"><Wifi className="h-4 w-4 text-cyan-300" /> {websocketState}</p>
        </div>
      </div>
      <div className="mt-5 flex flex-wrap gap-3">
        <ResumePublishButton />
        {account.needs_reauth ? (
          <a className="inline-flex items-center gap-2 rounded-2xl border border-rose-500/20 bg-rose-500/10 px-4 py-2 text-sm font-medium text-rose-200 transition hover:bg-rose-500/15" href="/api/auth/login">
            <ExternalLink className="h-4 w-4" />
            Reconnect
          </a>
        ) : null}
      </div>
    </GlowCard>
  );
}
