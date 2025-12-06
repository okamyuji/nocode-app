/**
 * ビュー関連の型定義
 */

/**
 * ビュー種類
 */
export type ViewType = "table" | "list" | "calendar" | "chart";

/**
 * ビュー設定
 */
export interface ViewConfig {
  columns?: string[];
  sort?: {
    field: string;
    order: "asc" | "desc";
  };
  filters?: {
    field: string;
    operator: string;
    value: string;
  }[];
  [key: string]: unknown;
}

/**
 * アプリビュー
 */
export interface AppView {
  id: number;
  app_id: number;
  name: string;
  view_type: ViewType;
  config?: ViewConfig;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * ビュー作成リクエスト
 */
export interface CreateViewRequest {
  name: string;
  view_type: ViewType;
  config?: ViewConfig;
  is_default?: boolean;
}

/**
 * ビュー更新リクエスト
 */
export interface UpdateViewRequest {
  name?: string;
  config?: ViewConfig;
  is_default?: boolean;
}
