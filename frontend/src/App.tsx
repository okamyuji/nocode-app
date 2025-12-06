/**
 * アプリケーションルートコンポーネント
 */

import { ErrorBoundary, Layout, ProtectedRoute } from "@/components/common";
import {
  AppChartPage,
  AppCreatePage,
  AppListPage,
  DashboardPage,
  DataManagementPage,
  LoginPage,
  RecordListPage,
  RegisterPage,
  SettingsPage,
} from "@/pages";
import { useAuthStore } from "@/stores";
import { useEffect } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

function App() {
  const { fetchUser, token } = useAuthStore();

  // 初期化時にトークンが存在すればユーザー情報を取得
  useEffect(() => {
    if (token) {
      fetchUser();
    }
  }, [token, fetchUser]);

  return (
    <ErrorBoundary>
      <Routes>
        {/* 公開ルート */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        {/* 認証必須ルート */}
        <Route
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<DashboardPage />} />
          <Route path="/apps" element={<AppListPage />} />
          <Route path="/apps/new" element={<AppCreatePage />} />
          <Route path="/apps/:appId/records" element={<RecordListPage />} />
          <Route path="/apps/:appId/charts" element={<AppChartPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/data" element={<DataManagementPage />} />
        </Route>

        {/* フォールバック */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </ErrorBoundary>
  );
}

export default App;
