/**
 * ダッシュボード関連の型定義
 */

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
