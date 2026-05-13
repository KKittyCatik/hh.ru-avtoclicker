import { useQuery } from '@tanstack/react-query';

import { apiClient } from '@/lib/api';

export function useAccounts() {
  return useQuery({
    queryKey: ['accounts'],
    queryFn: apiClient.getAccounts,
    staleTime: 30_000,
  });
}
