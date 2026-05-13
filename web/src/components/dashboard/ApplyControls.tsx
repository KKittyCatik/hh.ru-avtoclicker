import { Play, Rocket, Square } from 'lucide-react';

import { GradientButton } from '@/components/ui/GradientButton';
import { GlowCard } from '@/components/ui/GlowCard';
import { useApply } from '@/hooks/useApply';
import { useWorkerStore } from '@/store/worker';

export function ApplyControls() {
  const isRunning = useWorkerStore((state) => state.isRunning);
  const isPublishing = useWorkerStore((state) => state.isPublishing);
  const { startMutation, stopMutation, publishMutation } = useApply();

  return (
    <GlowCard>
      <div className="mb-5">
        <p className="text-sm uppercase tracking-[0.24em] text-muted">Actions</p>
        <h2 className="text-xl font-semibold text-foreground">Apply Controls</h2>
      </div>
      <div className="grid gap-3 md:grid-cols-3">
        <GradientButton disabled={startMutation.isPending || isRunning} onClick={() => startMutation.mutate()}>
          <Play className="h-4 w-4" />
          {startMutation.isPending ? 'Starting...' : 'Start Apply Worker'}
        </GradientButton>
        <GradientButton className="from-zinc-700 to-zinc-500" disabled={stopMutation.isPending || !isRunning} onClick={() => stopMutation.mutate()}>
          <Square className="h-4 w-4" />
          {stopMutation.isPending ? 'Stopping...' : 'Stop Worker'}
        </GradientButton>
        <GradientButton className="from-violet-600 to-fuchsia-500" disabled={publishMutation.isPending || isPublishing} onClick={() => publishMutation.mutate()}>
          <Rocket className="h-4 w-4" />
          {publishMutation.isPending ? 'Publishing...' : 'Publish Resume'}
        </GradientButton>
      </div>
    </GlowCard>
  );
}
