/**
 * 認証必須ルートコンポーネント
 * 未認証ユーザーをログインページにリダイレクトする
 */

import { useAuthStore } from "@/stores";
import { Navigate, useLocation } from "react-router-dom";
import { Loading } from "./Loading";

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, token } = useAuthStore();
  const location = useLocation();

  // 認証情報の読み込み中
  if (isLoading) {
    return <Loading fullScreen message="認証情報を確認中..." />;
  }

  // 未認証の場合はログインページへリダイレクト
  if (!isAuthenticated && !token) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
}
