import { ActivityFeed } from '@/components/dashboard/ActivityFeed';
import { ApplyControls } from '@/components/dashboard/ApplyControls';
import { PipelineChart } from '@/components/dashboard/PipelineChart';
import { RealtimeEvents } from '@/components/dashboard/RealtimeEvents';
import { StatsCards } from '@/components/dashboard/StatsCards';
import { LoadingScreen } from '@/components/ui/LoadingScreen';
import { useNegotiations } from '@/hooks/useNegotiations';
import { useDerivedStats, useStats } from '@/hooks/useStats';

export default function DashboardPage() {
  const statsQuery = useStats();
  const negotiationsQuery = useNegotiations();
  const { cards, pipeline } = useDerivedStats(negotiationsQuery.data, statsQuery.data);

  if (statsQuery.isLoading || negotiationsQuery.isLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="space-y-4">
      <StatsCards cards={cards} />
      <div className="grid gap-4 xl:grid-cols-[1.6fr_1fr]">
        <PipelineChart stages={pipeline} />
        <RealtimeEvents />
      </div>
      <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <ActivityFeed />
        <ApplyControls />
      </div>
    </div>
  );
}
