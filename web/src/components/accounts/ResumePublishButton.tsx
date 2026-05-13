import { Rocket } from 'lucide-react';

import { GradientButton } from '@/components/ui/GradientButton';
import { useApply } from '@/hooks/useApply';

export function ResumePublishButton() {
  const { publishMutation } = useApply();

  return (
    <GradientButton disabled={publishMutation.isPending} onClick={() => publishMutation.mutate()} type="button">
      <Rocket className="h-4 w-4" />
      {publishMutation.isPending ? 'Publishing...' : 'Publish resume'}
    </GradientButton>
  );
}
