import { BarChart3, Eye, MailPlus, Sparkles } from 'lucide-react';
import { motion } from 'framer-motion';

import { AnimatedCounter } from '@/components/ui/AnimatedCounter';
import { GlowCard } from '@/components/ui/GlowCard';
import { useStatsStore } from '@/store/stats';
import type { StatCard } from '@/types/stats';

const icons = {
  applies: BarChart3,
  invitations: MailPlus,
  views: Eye,
  replyRate: Sparkles,
};

export function StatsCards({ cards }: { cards: StatCard[] }) {
  const highlightedMetric = useStatsStore((state) => state.highlightedMetric);
  const setHighlightedMetric = useStatsStore((state) => state.setHighlightedMetric);

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      {cards.map((card, index) => {
        const Icon = icons[card.id];
        return (
          <motion.div key={card.id} initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: index * 0.06 }}>
            <GlowCard
              className={highlightedMetric === card.id ? 'border-violet-500/30 shadow-glow' : ''}
              onClick={() => setHighlightedMetric(card.id)}
            >
              <div className="mb-5 flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted">{card.label}</p>
                  <p className="mt-2 text-3xl font-semibold text-foreground"><AnimatedCounter value={card.value} suffix={card.suffix} /></p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/[0.04] p-3 text-violet-300">
                  <Icon className="h-5 w-5" />
                </div>
              </div>
              <p className="text-sm text-muted">{card.hint}</p>
            </GlowCard>
          </motion.div>
        );
      })}
    </div>
  );
}
