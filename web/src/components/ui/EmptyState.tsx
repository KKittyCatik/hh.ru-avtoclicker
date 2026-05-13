import type { ReactNode } from 'react';

export function EmptyState({ title, description, icon }: { title: string; description: string; icon?: ReactNode }) {
  return (
    <div className="glass-panel flex min-h-48 flex-col items-center justify-center rounded-3xl border border-dashed border-white/10 p-8 text-center">
      {icon ? <div className="mb-4 text-violet-300">{icon}</div> : null}
      <h3 className="mb-2 text-lg font-semibold text-foreground">{title}</h3>
      <p className="max-w-md text-sm text-muted">{description}</p>
    </div>
  );
}
