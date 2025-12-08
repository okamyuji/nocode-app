/**
 * API エクスポート
 */

// API実装のエクスポート
export { appsApi } from "./apps";
export { authApi } from "./auth";
export { chartsApi } from "./charts";
export { default as client } from "./client";
export { dashboardApi } from "./dashboard";
export { dashboardWidgetsApi } from "./dashboardWidgets";
export { createDataSourceApi } from "./datasources";
export type { DataSourceApi } from "./datasources";
export { fieldsApi } from "./fields";
export { recordsApi } from "./records";
export { profileApi, usersApi } from "./users";
export { viewsApi } from "./views";

// 依存性注入・テスト用インターフェースのエクスポート
export type {
  IApiClient,
  IAppsApi,
  IAuthApi,
  IChartsApi,
  IDashboardApi,
  IDashboardWidgetsApi,
  IDataSourcesApi,
  IFieldsApi,
  IProfileApi,
  IRecordsApi,
  IUsersApi,
  IViewsApi,
} from "./interfaces";

// コンテキストのエクスポート
export { ApiClientContext } from "./ApiClientContext";

// コンテキストプロバイダーのエクスポート
export { ApiClientProvider } from "./ApiClientProvider";

// 依存性注入用フックのエクスポート
export {
  useApiClient,
  useAppsApi,
  useAuthApi,
  useChartsApi,
  useDashboardApi,
  useDashboardWidgetsApi,
  useDataSourcesApi,
  useFieldsApi,
  useProfileApi,
  useRecordsApi,
  useUsersApi,
  useViewsApi,
} from "./useApiClient";

// 統合APIクライアントインスタンスのエクスポート
export { apiClient } from "./apiClientInstance";
