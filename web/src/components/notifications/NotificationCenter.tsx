import * as Dialog from '@radix-ui/react-dialog';
import * as ScrollArea from '@radix-ui/react-scroll-area';

import { NotificationToast } from '@/components/notifications/NotificationToast';
import { EmptyState } from '@/components/ui/EmptyState';
import { useNotificationsStore } from '@/store/notifications';

export function NotificationCenter() {
  const items = useNotificationsStore((state) => state.items);
  const isOpen = useNotificationsStore((state) => state.isOpen);
  const setOpen = useNotificationsStore((state) => state.setOpen);
  const markAllRead = useNotificationsStore((state) => state.markAllRead);

  return (
    <Dialog.Root open={isOpen} onOpenChange={setOpen}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed right-4 top-4 z-50 h-[calc(100vh-2rem)] w-[min(420px,calc(100vw-2rem))] rounded-3xl border border-white/10 bg-[#0d0d11]/95 p-5 shadow-soft">
          <div className="mb-4 flex items-center justify-between">
            <div>
              <Dialog.Title className="text-lg font-semibold text-foreground">Notification center</Dialog.Title>
              <Dialog.Description className="text-sm text-muted">История realtime событий и действий пользователя.</Dialog.Description>
            </div>
            <button className="text-sm text-violet-300" onClick={markAllRead} type="button">
              Mark all read
            </button>
          </div>
          <ScrollArea.Root className="h-[calc(100%-4rem)] overflow-hidden">
            <ScrollArea.Viewport className="h-full">
              <div className="space-y-3 pr-3">
                {items.length ? items.map((item) => <NotificationToast key={item.id} item={item} />) : <EmptyState title="Пока пусто" description="История уведомлений появится после первых realtime событий." />}
              </div>
            </ScrollArea.Viewport>
            <ScrollArea.Scrollbar orientation="vertical" className="flex w-2.5 touch-none p-0.5">
              <ScrollArea.Thumb className="relative flex-1 rounded-full bg-white/10" />
            </ScrollArea.Scrollbar>
          </ScrollArea.Root>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
