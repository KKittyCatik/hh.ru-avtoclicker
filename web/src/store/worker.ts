import { create } from 'zustand';
import { persist } from 'zustand/middleware';

import type { SocketState } from '@/lib/websocket';

type WorkerState = {
  isRunning: boolean;
  isPublishing: boolean;
  websocketState: SocketState;
  sidebarCollapsed: boolean;
  lastEventAt: string | null;
  setRunning: (value: boolean) => void;
  setPublishing: (value: boolean) => void;
  setWebsocketState: (value: SocketState) => void;
  touchEvent: () => void;
  toggleSidebar: () => void;
};

export const useWorkerStore = create<WorkerState>()(
  persist(
    (set) => ({
      isRunning: false,
      isPublishing: false,
      websocketState: 'idle',
      sidebarCollapsed: false,
      lastEventAt: null,
      setRunning: (value) => set({ isRunning: value }),
      setPublishing: (value) => set({ isPublishing: value }),
      setWebsocketState: (value) => set({ websocketState: value }),
      touchEvent: () => set({ lastEventAt: new Date().toISOString() }),
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
    }),
    {
      name: 'hh-worker-store',
      partialize: (state) => ({ sidebarCollapsed: state.sidebarCollapsed }),
    },
  ),
);
