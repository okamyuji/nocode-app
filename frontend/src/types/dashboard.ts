/**
 * ダッシュボード関連の型定義
 */

import type { App } from "./app";

/**
 * ダッシュボード統計
 */
export interface DashboardStats {
  app_count: number;
  total_records: number;
  user_count: number;
  todays_updates: number;
}

/**
 * GET /api/v1/dashboard/stats のレスポンス
 */
export interface DashboardStatsResponse {
  stats: DashboardStats;
}

/**
 * ウィジェットの表示形式
 */
export type WidgetViewType = "table" | "list" | "chart";

/**
 * ウィジェットのサイズ
 */
export type WidgetSize = "small" | "medium" | "large";

/**
 * ウィジェット設定
 */
export interface WidgetConfig {
  [key: string]: unknown;
}

/**
 * ダッシュボードウィジェット
 */
export interface DashboardWidget {
  id: number;
  user_id: number;
  app_id: number;
  display_order: number;
  view_type: WidgetViewType;
  is_visible: boolean;
  widget_size: WidgetSize;
  config?: WidgetConfig;
  created_at: string;
  updated_at: string;
  app?: App;
}

/**
 * ダッシュボードウィジェット作成リクエスト
 */
export interface CreateDashboardWidgetRequest {
  app_id: number;
  display_order?: number;
  view_type?: WidgetViewType;
  is_visible?: boolean;
  widget_size?: WidgetSize;
  config?: WidgetConfig;
}

/**
 * ダッシュボードウィジェット更新リクエスト
 */
export interface UpdateDashboardWidgetRequest {
  display_order?: number;
  view_type?: WidgetViewType;
  is_visible?: boolean;
  widget_size?: WidgetSize;
  config?: WidgetConfig;
}

/**
 * ウィジェット並び替えリクエスト
 */
export interface ReorderWidgetsRequest {
  widget_ids: number[];
}

/**
 * ダッシュボードウィジェット一覧レスポンス
 */
export interface DashboardWidgetListResponse {
  widgets: DashboardWidget[];
}
