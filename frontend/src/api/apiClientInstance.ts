/**
 * 統合APIクライアント インスタンス
 * テスト時にモック実装に置き換え可能
 */

import { appsApi } from "./apps";
import { authApi } from "./auth";
import { chartsApi } from "./charts";
import client from "./client";
import { dashboardApi } from "./dashboard";
import { dashboardWidgetsApi } from "./dashboardWidgets";
import { createDataSourceApi } from "./datasources";
import { fieldsApi } from "./fields";
import type { IApiClient } from "./interfaces";
import { recordsApi } from "./records";
import { profileApi, usersApi } from "./users";
import { viewsApi } from "./views";

export const apiClient: IApiClient = {
  auth: authApi,
  apps: appsApi,
  fields: fieldsApi,
  records: recordsApi,
  views: viewsApi,
  charts: chartsApi,
  users: usersApi,
  profile: profileApi,
  dashboard: dashboardApi,
  dashboardWidgets: dashboardWidgetsApi,
  dataSources: createDataSourceApi(client),
};
