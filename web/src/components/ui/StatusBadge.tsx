import { cva } from 'class-variance-authority';

import { cn } from '@/lib/utils';

const badgeStyles = cva('inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium', {
  variants: {
    variant: {
      neutral: 'border-white/10 bg-white/5 text-foreground',
      success: 'border-emerald-500/30 bg-emerald-500/12 text-emerald-300',
      warning: 'border-amber-500/30 bg-amber-500/12 text-amber-300',
      danger: 'border-rose-500/30 bg-rose-500/12 text-rose-300',
      info: 'border-cyan-500/30 bg-cyan-500/12 text-cyan-300',
      accent: 'border-violet-500/30 bg-violet-500/12 text-violet-200',
    },
  },
  defaultVariants: {
    variant: 'neutral',
  },
});

export function StatusBadge({ label, variant = 'neutral', className }: { label: string; variant?: 'neutral' | 'success' | 'warning' | 'danger' | 'info' | 'accent'; className?: string }) {
  return <span className={cn(badgeStyles({ variant }), className)}>{label}</span>;
}
