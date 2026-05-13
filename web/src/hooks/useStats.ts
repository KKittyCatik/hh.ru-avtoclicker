import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { apiClient } from '@/lib/api';
import type { Negotiation } from '@/types/negotiation';
import type { PipelineStage, StatCard } from '@/types/stats';

export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: apiClient.getStats,
    staleTime: 15_000,
  });
}

export function useDerivedStats(negotiations: Negotiation[] | undefined, stats: { applies_today: number; invitations: number; views: number } | undefined) {
  return useMemo(() => {
    const items = negotiations ?? [];
    const replies = items.filter((item) => item.needs_reply).length;
    const interviews = items.filter((item) => item.status.toLowerCase().includes('interview')).length;
    const offers = items.filter((item) => item.status.toLowerCase().includes('offer')).length;
    const replyRate = stats?.applies_today ? Math.round((replies / Math.max(stats.applies_today, 1)) * 100) : 0;

    const cards: StatCard[] = [
      { id: 'applies', label: 'Откликов сегодня', value: stats?.applies_today ?? 0, hint: 'Активность текущего цикла' },
      { id: 'invitations', label: 'Приглашения', value: stats?.invitations ?? 0, hint: 'Входящие сигналы от HR' },
      { id: 'views', label: 'Просмотры', value: stats?.views ?? 0, hint: 'Просмотры резюме работодателями' },
      { id: 'replyRate', label: 'AI reply rate', value: replyRate, suffix: '%', hint: 'Доля переписок, требующих ответа' },
    ];

    const pipeline: PipelineStage[] = [
      { name: 'Applied', value: stats?.applies_today ?? items.length },
      { name: 'Viewed', value: stats?.views ?? 0 },
      { name: 'HR replied', value: replies },
      { name: 'Interview', value: interviews },
      { name: 'Offer', value: offers },
    ];

    return { cards, pipeline };
  }, [negotiations, stats]);
}
