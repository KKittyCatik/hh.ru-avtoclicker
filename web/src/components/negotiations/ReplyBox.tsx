import { Sparkles, SendHorizonal } from 'lucide-react';
import { useEffect, useRef } from 'react';

import { GradientButton } from '@/components/ui/GradientButton';

export function ReplyBox({ value, generating, sending, onChange, onSend, onGenerate }: { value: string; generating: boolean; sending: boolean; onChange: (value: string) => void; onSend: () => void; onGenerate: () => void }) {
  const ref = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) {
      return;
    }
    element.style.height = '0px';
    element.style.height = `${element.scrollHeight}px`;
  }, [value]);

  return (
    <div className="glass-panel rounded-3xl p-4">
      <textarea
        ref={ref}
        className="min-h-[96px] w-full resize-none rounded-2xl border border-white/10 bg-white/[0.03] px-4 py-3 text-sm text-foreground placeholder:text-muted"
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && event.ctrlKey) {
            event.preventDefault();
            onSend();
          }
        }}
        placeholder="Напишите ответ HR или сгенерируйте AI draft..."
        value={value}
      />
      <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
        <p className="text-xs text-muted">Ctrl+Enter — отправить, Space — AI draft, R — фокус на ответ</p>
        <div className="flex gap-3">
          <GradientButton className="from-zinc-700 to-zinc-500" disabled={generating} onClick={onGenerate} type="button">
            <Sparkles className="h-4 w-4" />
            {generating ? 'Generating...' : 'Generate AI Reply'}
          </GradientButton>
          <GradientButton disabled={sending || value.trim().length === 0} onClick={onSend} type="button">
            <SendHorizonal className="h-4 w-4" />
            {sending ? 'Sending...' : 'Send'}
          </GradientButton>
        </div>
      </div>
    </div>
  );
}
