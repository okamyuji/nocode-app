/**
 * フィールド関連の型定義
 */

/**
 * フィールド種類
 */
export type FieldType =
  | "text"
  | "textarea"
  | "number"
  | "date"
  | "datetime"
  | "select"
  | "multiselect"
  | "checkbox"
  | "radio"
  | "link"
  | "attachment";

/**
 * フィールドオプション
 */
export interface FieldOptions {
  choices?: string[];
  link_type?: "url" | "email";
  [key: string]: unknown;
}

/**
 * フィールド
 */
export interface Field {
  id: number;
  app_id: number;
  field_code: string;
  field_name: string;
  field_type: FieldType;
  options?: FieldOptions;
  required: boolean;
  display_order: number;
  created_at: string;
  updated_at: string;
}

/**
 * フィールド作成リクエスト
 */
export interface CreateFieldRequest {
  field_code: string;
  field_name: string;
  field_type: FieldType;
  options?: FieldOptions;
  required?: boolean;
  display_order?: number;
}

/**
 * フィールド更新リクエスト
 */
export interface UpdateFieldRequest {
  field_name?: string;
  options?: FieldOptions;
  required?: boolean;
  display_order?: number;
}

/**
 * フィールド順序アイテム
 */
export interface FieldOrderItem {
  id: number;
  display_order: number;
}

/**
 * フィールド順序更新リクエスト
 */
export interface UpdateFieldOrderRequest {
  fields: FieldOrderItem[];
}

/**
 * フィールド種類の表示ラベル
 */
export const FIELD_TYPE_LABELS: Record<FieldType, string> = {
  text: "テキスト",
  textarea: "複数行テキスト",
  number: "数値",
  date: "日付",
  datetime: "日時",
  select: "選択（単一）",
  multiselect: "選択（複数）",
  checkbox: "チェックボックス",
  radio: "ラジオボタン",
  link: "リンク",
  attachment: "添付ファイル",
};
