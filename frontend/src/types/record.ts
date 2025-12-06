/**
 * レコード関連の型定義
 */

import { Pagination } from "./app";

/**
 * レコードデータ
 */
export type RecordData = Record<string, unknown>;

/**
 * レコードアイテム
 */
export interface RecordItem {
  id: number;
  data: RecordData;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/**
 * レコード作成リクエスト
 */
export interface CreateRecordRequest {
  data: RecordData;
}

/**
 * レコード更新リクエスト
 */
export interface UpdateRecordRequest {
  data: RecordData;
}

/**
 * レコード一括作成リクエスト
 */
export interface BulkCreateRecordRequest {
  records: RecordData[];
}

/**
 * レコード一括削除リクエスト
 */
export interface BulkDeleteRecordRequest {
  ids: number[];
}

/**
 * レコード一覧レスポンス
 */
export interface RecordListResponse {
  records: RecordItem[];
  pagination: Pagination;
}

/**
 * レコードクエリオプション
 */
export interface RecordQueryOptions {
  page?: number;
  limit?: number;
  sort?: string;
  order?: "asc" | "desc";
  filters?: FilterItem[];
}

/**
 * フィルター条件
 */
export interface FilterItem {
  field: string;
  operator: "eq" | "ne" | "gt" | "gte" | "lt" | "lte" | "like" | "in";
  value: string;
}
