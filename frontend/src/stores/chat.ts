import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AppStore {
  // UI 状态
  sidebarOpen: boolean;
  darkMode: boolean;
  
  // 操作方法
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  toggleDarkMode: () => void;
  setDarkMode: (dark: boolean) => void;
}

export const useChatStore = create<AppStore>()(
  persist(
    (set) => ({
      // 初始状态
      sidebarOpen: true,
      darkMode: false,

      // 操作方法
      toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      toggleDarkMode: () => set((state) => ({ darkMode: !state.darkMode })),
      setDarkMode: (dark) => set({ darkMode: dark }),
    }),
    {
      name: 'app-store',
      partialize: (state) => ({
        darkMode: state.darkMode,
      }),
    }
  )
);