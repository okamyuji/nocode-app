/**
 * UI状態管理ストア
 */

import { create } from "zustand";
import { persist } from "zustand/middleware";

/**
 * UI状態のインターフェース
 */
interface UIState {
  sidebarWidth: number;
  sidebarCollapsed: boolean;
  setSidebarWidth: (width: number) => void;
  toggleSidebarCollapsed: () => void;
}

// サイドバー幅の定数
const MIN_SIDEBAR_WIDTH = 60;
const MAX_SIDEBAR_WIDTH = 400;
const DEFAULT_SIDEBAR_WIDTH = 240;

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarWidth: DEFAULT_SIDEBAR_WIDTH,
      sidebarCollapsed: false,

      // サイドバー幅を設定（最小/最大値で制限）
      setSidebarWidth: (width: number) =>
        set({
          sidebarWidth: Math.min(
            MAX_SIDEBAR_WIDTH,
            Math.max(MIN_SIDEBAR_WIDTH, width)
          ),
        }),

      // サイドバーの折りたたみを切り替え
      toggleSidebarCollapsed: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
    }),
    {
      name: "ui-settings",
    }
  )
);

export { DEFAULT_SIDEBAR_WIDTH, MAX_SIDEBAR_WIDTH, MIN_SIDEBAR_WIDTH };
