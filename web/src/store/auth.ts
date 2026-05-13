import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type AppSettings = {
  llmProvider: 'openai' | 'deepseek';
  dailyLimit: number;
  scheduleFilter: 'remote' | 'hybrid' | 'office' | 'any';
  minimumSalary: number;
  excludeAgencies: boolean;
};

type AuthState = {
  workspaceName: string;
  activeAccountId: string | null;
  settings: AppSettings;
  setWorkspaceName: (value: string) => void;
  setActiveAccountId: (value: string | null) => void;
  updateSettings: (value: Partial<AppSettings>) => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      workspaceName: 'hh.ru Copilot',
      activeAccountId: null,
      settings: {
        llmProvider: 'openai',
        dailyLimit: 50,
        scheduleFilter: 'remote',
        minimumSalary: 180000,
        excludeAgencies: true,
      },
      setWorkspaceName: (value) => set({ workspaceName: value }),
      setActiveAccountId: (value) => set({ activeAccountId: value }),
      updateSettings: (value) => set((state) => ({ settings: { ...state.settings, ...value } })),
    }),
    { name: 'hh-auth-store' },
  ),
);
