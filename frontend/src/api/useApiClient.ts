/**
 * 依存性注入用APIクライアントフック
 */

import { useContext } from "react";
import { ApiClientContext } from "./ApiClientContext";
import type { IApiClient } from "./interfaces";

/**
 * コンテキストからAPIクライアントにアクセスするフック
 * @returns APIクライアントインスタンス
 * @throws ApiClientProvider外で使用された場合にエラー
 */
export function useApiClient(): IApiClient {
  const context = useContext(ApiClientContext);
  if (!context) {
    throw new Error("useApiClient must be used within an ApiClientProvider");
  }
  return context;
}

/**
 * 便利な個別APIフック
 */

// 認証API
export function useAuthApi() {
  return useApiClient().auth;
}

// アプリAPI
export function useAppsApi() {
  return useApiClient().apps;
}

// フィールドAPI
export function useFieldsApi() {
  return useApiClient().fields;
}

// レコードAPI
export function useRecordsApi() {
  return useApiClient().records;
}

// ビューAPI
export function useViewsApi() {
  return useApiClient().views;
}

// チャートAPI
export function useChartsApi() {
  return useApiClient().charts;
}

// ユーザーAPI
export function useUsersApi() {
  return useApiClient().users;
}

// プロフィールAPI
export function useProfileApi() {
  return useApiClient().profile;
}

// ダッシュボードAPI
export function useDashboardApi() {
  return useApiClient().dashboard;
}

// ダッシュボードウィジェットAPI
export function useDashboardWidgetsApi() {
  return useApiClient().dashboardWidgets;
}

// データソースAPI
export function useDataSourcesApi() {
  return useApiClient().dataSources;
}
