import { nanoid } from 'nanoid';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

import type { RealtimeEventName } from '@/types/api';

export type NotificationItem = {
  id: string;
  event: RealtimeEventName | 'system';
  title: string;
  description: string;
  createdAt: string;
  read: boolean;
};

type NotificationsState = {
  items: NotificationItem[];
  isOpen: boolean;
  addNotification: (notification: Omit<NotificationItem, 'id' | 'createdAt' | 'read'>) => void;
  markAllRead: () => void;
  setOpen: (value: boolean) => void;
};

export const useNotificationsStore = create<NotificationsState>()(
  persist(
    (set) => ({
      items: [],
      isOpen: false,
      addNotification: (notification) =>
        set((state) => ({
          items: [
            {
              ...notification,
              id: nanoid(),
              createdAt: new Date().toISOString(),
              read: false,
            },
            ...state.items,
          ].slice(0, 60),
        })),
      markAllRead: () => set((state) => ({ items: state.items.map((item) => ({ ...item, read: true })) })),
      setOpen: (value) => set({ isOpen: value }),
    }),
    { name: 'hh-notifications-store' },
  ),
);
