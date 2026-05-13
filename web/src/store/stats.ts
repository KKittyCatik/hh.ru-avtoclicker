import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type StatsUIState = {
  highlightedMetric: 'applies' | 'invitations' | 'views' | 'replyRate';
  setHighlightedMetric: (metric: StatsUIState['highlightedMetric']) => void;
};

export const useStatsStore = create<StatsUIState>()(
  persist(
    (set) => ({
      highlightedMetric: 'applies',
      setHighlightedMetric: (metric) => set({ highlightedMetric: metric }),
    }),
    { name: 'hh-stats-ui-store' },
  ),
);
