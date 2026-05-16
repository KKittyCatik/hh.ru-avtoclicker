import { EmptyState } from '@/components/ui/EmptyState';
import type { Account } from '@/types/account';
import { AccountCard } from '@/components/accounts/AccountCard';

export function AccountList({ accounts }: { accounts: Account[] }) {
  if (!accounts.length) {
    return (
      <div className="space-y-4">
        <EmptyState title="Нет аккаунтов" description="Авторизуйтесь через hh.ru, чтобы добавить аккаунт автоматически." />
        <div className="flex justify-center">
          <a
            href="/api/auth/login"
            className="inline-flex items-center rounded-2xl border border-cyan-500/30 bg-cyan-500/15 px-4 py-2 text-sm font-medium text-cyan-100 transition hover:bg-cyan-500/25"
          >
            Войти через hh.ru
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {accounts.map((account) => (
        <AccountCard key={account.id} account={account} />
      ))}
    </div>
  );
}
