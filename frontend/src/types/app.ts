/**
 * アプリケーション関連の型定義
 */

import { CreateFieldRequest, Field } from "./field";

/**
 * ページネーション情報
 */
export interface Pagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

/**
 * アプリケーション
 */
export interface App {
  id: number;
  name: string;
  description: string;
  table_name: string;
  icon: string;
  created_by: number;
  created_at: string;
  updated_at: string;
  fields?: Field[];
  field_count: number;
}

/**
 * アプリ作成リクエスト
 */
export interface CreateAppRequest {
  name: string;
  description?: string;
  icon?: string;
  fields: CreateFieldRequest[];
}

/**
 * アプリ更新リクエスト
 */
export interface UpdateAppRequest {
  name?: string;
  description?: string;
  icon?: string;
}

/**
 * アプリ一覧レスポンス
 */
export interface AppListResponse {
  apps: App[];
  pagination: Pagination;
}
