import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { formatEventLabel } from '@/lib/format';
import { realtimeClient } from '@/lib/websocket';
import { useNotificationsStore } from '@/store/notifications';
import { useWorkerStore } from '@/store/worker';
import type { RealtimeEvent } from '@/types/api';

function describePayload(payload: unknown): string {
  if (typeof payload === 'string') {
    return payload;
  }
  if (payload && typeof payload === 'object') {
    return JSON.stringify(payload);
  }
  return 'Новое событие системы';
}

export function useRealtime(): void {
  const queryClient = useQueryClient();
  const addNotification = useNotificationsStore((state) => state.addNotification);
  const setWebsocketState = useWorkerStore((state) => state.setWebsocketState);
  const touchEvent = useWorkerStore((state) => state.touchEvent);

  useEffect(() => {
    const unsubscribeState = realtimeClient.onStateChange(setWebsocketState);
    const unsubscribeEvents = realtimeClient.subscribe((incoming) => {
      if ('type' in incoming) {
        return;
      }

      const event = incoming as RealtimeEvent;
      const title = formatEventLabel(event.event);
      const description = describePayload(event.payload);
      addNotification({ event: event.event, title, description });
      touchEvent();
      void queryClient.invalidateQueries({ queryKey: ['stats'] });
      void queryClient.invalidateQueries({ queryKey: ['negotiations'] });
      void queryClient.invalidateQueries({ queryKey: ['accounts'] });
      toast(title, { description });
    });

    realtimeClient.connect();

    return () => {
      unsubscribeState();
      unsubscribeEvents();
      realtimeClient.disconnect();
    };
  }, [addNotification, queryClient, setWebsocketState, touchEvent]);
}
