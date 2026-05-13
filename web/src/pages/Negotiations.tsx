import { useEffect, useMemo, useState } from 'react';

import { ChatView } from '@/components/negotiations/ChatView';
import { NegotiationList } from '@/components/negotiations/NegotiationList';
import { LoadingScreen } from '@/components/ui/LoadingScreen';
import { useNegotiationMessages, useNegotiations } from '@/hooks/useNegotiations';

export default function NegotiationsPage() {
  const negotiationsQuery = useNegotiations();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const selected = useMemo(() => negotiationsQuery.data?.find((item) => item.id === selectedId) ?? negotiationsQuery.data?.[0] ?? null, [negotiationsQuery.data, selectedId]);
  const messagesQuery = useNegotiationMessages(selected?.id ?? null);

  useEffect(() => {
    if (!selectedId && negotiationsQuery.data?.length) {
      setSelectedId(negotiationsQuery.data[0].id);
    }
  }, [negotiationsQuery.data, selectedId]);

  useEffect(() => {
    const listener = (event: KeyboardEvent) => {
      if (!negotiationsQuery.data?.length || event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) {
        return;
      }
      const currentIndex = negotiationsQuery.data.findIndex((item) => item.id === (selected?.id ?? ''));
      if (event.key.toLowerCase() === 'j') {
        event.preventDefault();
        const next = negotiationsQuery.data[Math.min(currentIndex + 1, negotiationsQuery.data.length - 1)];
        if (next) setSelectedId(next.id);
      }
      if (event.key.toLowerCase() === 'k') {
        event.preventDefault();
        const previous = negotiationsQuery.data[Math.max(currentIndex - 1, 0)];
        if (previous) setSelectedId(previous.id);
      }
    };

    window.addEventListener('keydown', listener);
    return () => window.removeEventListener('keydown', listener);
  }, [negotiationsQuery.data, selected]);

  if (negotiationsQuery.isLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="grid gap-4 xl:grid-cols-[420px_1fr]">
      <NegotiationList items={negotiationsQuery.data ?? []} selectedId={selected?.id ?? null} onSelect={setSelectedId} />
      <ChatView negotiation={selected} messages={messagesQuery.data} isLoading={messagesQuery.isLoading} refetch={() => void messagesQuery.refetch()} />
    </div>
  );
}
