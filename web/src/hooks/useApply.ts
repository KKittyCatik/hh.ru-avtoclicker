import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { apiClient } from '@/lib/api';
import { useWorkerStore } from '@/store/worker';

export function useApply() {
  const queryClient = useQueryClient();
  const setRunning = useWorkerStore((state) => state.setRunning);
  const setPublishing = useWorkerStore((state) => state.setPublishing);

  const startMutation = useMutation({
    mutationFn: apiClient.startApply,
    onMutate: () => setRunning(true),
    onError: (error) => {
      setRunning(false);
      toast.error('Не удалось запустить воркер', { description: error.message });
    },
    onSuccess: async () => {
      toast.success('Воркер откликов запущен');
      await queryClient.invalidateQueries({ queryKey: ['stats'] });
    },
  });

  const stopMutation = useMutation({
    mutationFn: apiClient.stopApply,
    onMutate: () => setRunning(false),
    onError: (error) => {
      setRunning(true);
      toast.error('Не удалось остановить воркер', { description: error.message });
    },
    onSuccess: () => {
      toast.success('Воркер остановлен');
    },
  });

  const publishMutation = useMutation({
    mutationFn: apiClient.publishResume,
    onMutate: () => setPublishing(true),
    onError: (error) => {
      setPublishing(false);
      toast.error('Не удалось опубликовать резюме', { description: error.message });
    },
    onSuccess: async () => {
      toast.success('Резюме опубликовано');
      await queryClient.invalidateQueries({ queryKey: ['accounts'] });
    },
    onSettled: () => setPublishing(false),
  });

  return { startMutation, stopMutation, publishMutation };
}
