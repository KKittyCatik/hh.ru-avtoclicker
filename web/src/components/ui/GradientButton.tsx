import type { ButtonHTMLAttributes, PropsWithChildren } from 'react';

import { cn } from '@/lib/utils';

export function GradientButton({ children, className, disabled, ...props }: PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement>>) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-2xl border border-white/10 bg-gradient-to-r from-violet-600 to-cyan-500 px-4 py-2 text-sm font-medium text-white transition hover:scale-[1.01] hover:shadow-glow disabled:cursor-not-allowed disabled:opacity-60',
        className,
      )}
      disabled={disabled}
      {...props}
    >
      {children}
    </button>
  );
}
