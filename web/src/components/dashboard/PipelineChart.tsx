import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

import { GlowCard } from '@/components/ui/GlowCard';
import type { PipelineStage } from '@/types/stats';

const colors = ['#7C3AED', '#8B5CF6', '#0EA5E9', '#06B6D4', '#14B8A6'];

export function PipelineChart({ stages }: { stages: PipelineStage[] }) {
  return (
    <GlowCard className="h-[360px]">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <p className="text-sm uppercase tracking-[0.24em] text-muted">Pipeline</p>
          <h2 className="text-xl font-semibold text-foreground">Воронка откликов</h2>
        </div>
      </div>
      <ResponsiveContainer width="100%" height="85%">
        <BarChart data={stages} margin={{ left: -24, right: 0, top: 10, bottom: 0 }}>
          <CartesianGrid vertical={false} stroke="rgba(255,255,255,0.06)" />
          <XAxis dataKey="name" tick={{ fill: '#A1A1AA', fontSize: 12 }} axisLine={false} tickLine={false} />
          <YAxis tick={{ fill: '#A1A1AA', fontSize: 12 }} axisLine={false} tickLine={false} />
          <Tooltip cursor={{ fill: 'rgba(255,255,255,0.03)' }} contentStyle={{ background: '#111113', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 16 }} />
          <Bar dataKey="value" radius={[14, 14, 0, 0]}>
            {stages.map((stage, index) => (
              <Cell key={stage.name} fill={colors[index % colors.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </GlowCard>
  );
}
