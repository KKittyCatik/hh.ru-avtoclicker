import { AccountList } from '@/components/accounts/AccountList';
import { LoadingScreen } from '@/components/ui/LoadingScreen';
import { useAccounts } from '@/hooks/useAccounts';

export default function AccountsPage() {
  const accountsQuery = useAccounts();

  if (accountsQuery.isLoading) {
    return <LoadingScreen />;
  }

  return <AccountList accounts={accountsQuery.data ?? []} />;
}
