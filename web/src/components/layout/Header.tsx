import { Bell, Command, Search } from 'lucide-react';

export function Header({ title, onOpenPalette, onOpenNotifications }: { title: string; onOpenPalette: () => void; onOpenNotifications: () => void }) {
  return (
    <header className="sticky top-0 z-20 border-b border-white/6 bg-background/80 backdrop-blur-xl">
      <div className="mx-auto flex max-w-[1440px] items-center justify-between gap-4 px-4 py-4 md:px-6">
        <div>
          <p className="text-xs uppercase tracking-[0.24em] text-muted">AI job search copilot</p>
          <h1 className="text-2xl font-semibold text-foreground">{title}</h1>
        </div>
        <div className="flex items-center gap-3">
          <button onClick={onOpenPalette} type="button" className="glass-panel hidden items-center gap-2 rounded-2xl px-4 py-2 text-sm text-muted md:flex">
            <Search className="h-4 w-4" />
            Search or run command
            <span className="rounded-lg border border-white/10 px-2 py-0.5 text-xs text-foreground"><Command className="inline h-3 w-3" />K</span>
          </button>
          <button onClick={onOpenNotifications} type="button" className="glass-panel rounded-2xl p-3 text-muted transition hover:text-foreground">
            <Bell className="h-4 w-4" />
          </button>
        </div>
      </div>
    </header>
  );
}
