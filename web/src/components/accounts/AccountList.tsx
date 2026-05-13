import { EmptyState } from '@/components/ui/EmptyState';
import type { Account } from '@/types/account';
import { AccountCard } from '@/components/accounts/AccountCard';

export function AccountList({ accounts }: { accounts: Account[] }) {
  if (!accounts.length) {
    return <EmptyState title="Нет аккаунтов" description="Подключи hh.ru аккаунты в backend-конфигурации, чтобы увидеть их здесь." />;
  }

  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {accounts.map((account) => (
        <AccountCard key={account.id} account={account} />
      ))}
    </div>
  );
}
