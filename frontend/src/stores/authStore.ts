/**
 * 認証状態管理ストア
 */

import { authApi } from "@/api";
import { User } from "@/types";
import { create } from "zustand";
import { persist } from "zustand/middleware";

/**
 * 認証状態のインターフェース
 */
interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
      isAdmin: false,

      // ユーザーを設定
      setUser: (user) =>
        set({ user, isAuthenticated: !!user, isAdmin: user?.role === "admin" }),

      // トークンを設定
      setToken: (token) => {
        if (token) {
          localStorage.setItem("token", token);
        } else {
          localStorage.removeItem("token");
        }
        set({ token });
      },

      // ログイン
      login: async (email, password) => {
        set({ isLoading: true });
        try {
          const response = await authApi.login({ email, password });
          get().setToken(response.token);
          set({
            user: response.user,
            isAuthenticated: true,
            isAdmin: response.user.role === "admin",
          });
        } finally {
          set({ isLoading: false });
        }
      },

      // 登録
      register: async (email, password, name) => {
        set({ isLoading: true });
        try {
          const response = await authApi.register({ email, password, name });
          get().setToken(response.token);
          set({
            user: response.user,
            isAuthenticated: true,
            isAdmin: response.user.role === "admin",
          });
        } finally {
          set({ isLoading: false });
        }
      },

      // ログアウト
      logout: () => {
        get().setToken(null);
        set({ user: null, isAuthenticated: false, isAdmin: false });
      },

      // ユーザー情報を取得
      fetchUser: async () => {
        const token = get().token || localStorage.getItem("token");
        if (!token) return;

        set({ isLoading: true });
        try {
          const user = await authApi.me();
          set({
            user,
            isAuthenticated: true,
            token,
            isAdmin: user.role === "admin",
          });
        } catch {
          get().logout();
        } finally {
          set({ isLoading: false });
        }
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({ token: state.token }),
    }
  )
);
