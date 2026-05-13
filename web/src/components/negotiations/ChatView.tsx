import { useEffect, useState } from 'react';
import { toast } from 'sonner';

import { EmptyState } from '@/components/ui/EmptyState';
import { LoadingScreen } from '@/components/ui/LoadingScreen';
import { MessageBubble } from '@/components/negotiations/MessageBubble';
import { ReplyBox } from '@/components/negotiations/ReplyBox';
import { sleep } from '@/lib/utils';
import { useGenerateReply, useSendReply } from '@/hooks/useNegotiations';
import type { Negotiation, NegotiationMessage } from '@/types/negotiation';

export function ChatView({ negotiation, messages, isLoading, refetch }: { negotiation: Negotiation | null; messages: NegotiationMessage[] | undefined; isLoading: boolean; refetch: () => void }) {
  const [draft, setDraft] = useState('');
  const [quickReplyOptionId, setQuickReplyOptionId] = useState<string | null>(null);
  const [isStreaming, setIsStreaming] = useState(false);
  const sendReply = useSendReply(negotiation?.id ?? null);
  const generateReply = useGenerateReply(negotiation?.id ?? null);

  useEffect(() => {
    setDraft('');
    setQuickReplyOptionId(null);
  }, [negotiation?.id]);

  useEffect(() => {
    const listener = (event: KeyboardEvent) => {
      if (!negotiation) {
        return;
      }
      if (event.key.toLowerCase() === 'r') {
        const field = document.querySelector<HTMLTextAreaElement>('textarea');
        field?.focus();
      }
      if (event.code === 'Space' && document.activeElement?.tagName !== 'TEXTAREA' && !generateReply.isPending) {
        event.preventDefault();
        void handleGenerate();
      }
    };

    window.addEventListener('keydown', listener);
    return () => window.removeEventListener('keydown', listener);
  });

  if (!negotiation) {
    return <EmptyState title="Выберите переговоры" description="Слева доступен список активных диалогов. Навигируйся клавишами J/K." />;
  }

  const handleSend = async () => {
    try {
      await sendReply.mutateAsync(
        quickReplyOptionId ? { quickReplyOptionId } : { text: draft },
      );
      setDraft('');
      setQuickReplyOptionId(null);
      refetch();
    } catch (error) {
      toast.error('Не удалось отправить ответ', { description: error instanceof Error ? error.message : 'Unknown error' });
    }
  };

  async function handleGenerate() {
    try {
      const generated = await generateReply.mutateAsync();
      setIsStreaming(true);
      setDraft('');
      setQuickReplyOptionId(generated.quickReplyOptionId ?? null);
      for (const symbol of generated.text) {
        setDraft((current) => current + symbol);
        await sleep(8);
      }
      setIsStreaming(false);
      if (generated.quickReplyOptionId) {
        toast.info('Бот-сценарий распознан', { description: 'Можно отправить быстрый ответ без ручного ввода.' });
      }
    } catch (error) {
      setIsStreaming(false);
      toast.error('Не удалось сгенерировать AI reply', { description: error instanceof Error ? error.message : 'Unknown error' });
    }
  }

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="glass-panel flex-1 rounded-3xl p-4">
        <div className="mb-4 flex items-center justify-between gap-3 border-b border-white/6 pb-4">
          <div>
            <h2 className="text-lg font-semibold text-foreground">Negotiation #{negotiation.id}</h2>
            <p className="text-sm text-muted">Status: {negotiation.status}</p>
          </div>
        </div>
        <div className="space-y-4">
          {isLoading ? <LoadingScreen /> : messages?.map((message) => <MessageBubble key={message.id} message={message} />)}
        </div>
      </div>
      <ReplyBox
        value={draft}
        generating={generateReply.isPending || isStreaming}
        sending={sendReply.isPending}
        onChange={(value) => {
          setDraft(value);
          setQuickReplyOptionId(null);
        }}
        onGenerate={() => void handleGenerate()}
        onSend={() => void handleSend()}
      />
    </div>
  );
}
