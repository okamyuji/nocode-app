/**
 * 認証フック
 */

import { useAuthStore } from "@/stores";
import { useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";

/**
 * 認証関連の状態とアクションを提供するフック
 */
export function useAuth() {
  const {
    user,
    token,
    isLoading,
    isAuthenticated,
    isAdmin,
    login,
    register,
    logout,
    fetchUser,
  } = useAuthStore();

  const navigate = useNavigate();

  // トークンがあるがユーザー情報がない場合、ユーザー情報を取得
  useEffect(() => {
    if (token && !user) {
      fetchUser();
    }
  }, [token, user, fetchUser]);

  // ログイン処理
  const handleLogin = useCallback(
    async (email: string, password: string) => {
      await login(email, password);
      navigate("/");
    },
    [login, navigate]
  );

  // 登録処理
  const handleRegister = useCallback(
    async (email: string, password: string, name: string) => {
      await register(email, password, name);
      navigate("/");
    },
    [register, navigate]
  );

  // ログアウト処理
  const handleLogout = useCallback(() => {
    logout();
    navigate("/login");
  }, [logout, navigate]);

  return {
    user,
    isLoading,
    isAuthenticated,
    isAdmin,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
  };
}
