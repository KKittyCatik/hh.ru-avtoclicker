import { motion, type HTMLMotionProps } from 'framer-motion';
import type { PropsWithChildren } from 'react';

import { cn } from '@/lib/utils';

export function GlowCard({ children, className, ...props }: PropsWithChildren<HTMLMotionProps<'div'>>) {
  return (
    <motion.div
      whileHover={{ y: -4, scale: 1.01 }}
      transition={{ type: 'spring', stiffness: 240, damping: 20 }}
      className={cn(
        'glass-panel rounded-3xl border border-white/10 bg-white/[0.04] p-5 shadow-soft',
        className,
      )}
      {...props}
    >
      {children}
    </motion.div>
  );
}
