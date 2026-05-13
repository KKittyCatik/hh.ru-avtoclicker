import { Search } from 'lucide-react';
import { useMemo, useState } from 'react';

import { EmptyState } from '@/components/ui/EmptyState';
import type { Negotiation } from '@/types/negotiation';
import { NegotiationCard } from '@/components/negotiations/NegotiationCard';

export function NegotiationList({ items, selectedId, onSelect }: { items: Negotiation[]; selectedId: string | null; onSelect: (id: string) => void }) {
  const [query, setQuery] = useState('');

  const filtered = useMemo(() => items.filter((item) => `${item.id} ${item.status} ${item.last_message.text}`.toLowerCase().includes(query.toLowerCase())), [items, query]);

  return (
    <div className="glass-panel rounded-3xl p-4">
      <div className="mb-4 flex items-center gap-3 rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3">
        <Search className="h-4 w-4 text-muted" />
        <input className="w-full bg-transparent text-sm text-foreground placeholder:text-muted" onChange={(event) => setQuery(event.target.value)} placeholder="Search negotiations..." value={query} />
      </div>
      <div className="space-y-3">
        {filtered.length ? filtered.map((item) => <NegotiationCard key={item.id} item={item} onClick={() => onSelect(item.id)} selected={item.id === selectedId} />) : <EmptyState title="Ничего не найдено" description="Измени поисковый запрос или дождись новых переговоров." />}
      </div>
    </div>
  );
}
