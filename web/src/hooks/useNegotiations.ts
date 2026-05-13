import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { apiClient } from '@/lib/api';

export function useNegotiations() {
  return useQuery({
    queryKey: ['negotiations'],
    queryFn: apiClient.getNegotiations,
    staleTime: 10_000,
  });
}

export function useNegotiationMessages(negotiationId: string | null) {
  return useQuery({
    queryKey: ['negotiations', negotiationId, 'messages'],
    queryFn: () => apiClient.getNegotiationMessages(negotiationId ?? ''),
    enabled: Boolean(negotiationId),
  });
}

export function useSendReply(negotiationId: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: { text?: string; quickReplyOptionId?: string }) => apiClient.sendReply(negotiationId ?? '', payload),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['negotiations'] }),
        queryClient.invalidateQueries({ queryKey: ['negotiations', negotiationId, 'messages'] }),
      ]);
    },
  });
}

export function useGenerateReply(negotiationId: string | null) {
  return useMutation({
    mutationFn: () => apiClient.generateReply(negotiationId ?? ''),
  });
}
