import type { PropsWithChildren } from 'react';

import { cn } from '@/lib/utils';

export function PageContainer({ children, className }: PropsWithChildren<{ className?: string }>) {
  return <div className={cn('mx-auto w-full max-w-[1440px] px-4 pb-10 md:px-6', className)}>{children}</div>;
}
