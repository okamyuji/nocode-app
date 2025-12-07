/**
 * データソース関連の型定義
 */

/** データベースタイプ */
export type DBType = "postgresql" | "mysql" | "oracle" | "sqlserver";

/** データベースタイプの表示名 */
export const DB_TYPE_LABELS: Record<DBType, string> = {
  postgresql: "PostgreSQL",
  mysql: "MySQL",
  oracle: "Oracle",
  sqlserver: "SQL Server",
};

/** デフォルトポート番号 */
export const DEFAULT_PORTS: Record<DBType, number> = {
  postgresql: 5432,
  mysql: 3306,
  oracle: 1521,
  sqlserver: 1433,
};

/** データソース */
export interface DataSource {
  id: number;
  name: string;
  db_type: DBType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** データソース作成リクエスト */
export interface CreateDataSourceRequest {
  name: string;
  db_type: DBType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
}

/** データソース更新リクエスト */
export interface UpdateDataSourceRequest {
  name?: string;
  host?: string;
  port?: number;
  database_name?: string;
  username?: string;
  password?: string;
}

/** テスト接続リクエスト */
export interface TestConnectionRequest {
  db_type: DBType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
}

/** テスト接続レスポンス */
export interface TestConnectionResponse {
  success: boolean;
  message: string;
}

/** データソース一覧レスポンス */
export interface DataSourceListResponse {
  data_sources: DataSource[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

/** テーブル情報 */
export interface TableInfo {
  name: string;
  schema?: string;
}

/** テーブル一覧レスポンス */
export interface TableListResponse {
  tables: TableInfo[];
}

/** カラム情報 */
export interface ColumnInfo {
  name: string;
  data_type: string;
  is_nullable: boolean;
  is_primary_key: boolean;
  default_value?: string;
}

/** カラム一覧レスポンス */
export interface ColumnListResponse {
  columns: ColumnInfo[];
}

/** 外部アプリ作成リクエストのフィールド */
export interface CreateExternalFieldRequest {
  source_column_name: string;
  field_code: string;
  field_name: string;
  field_type: string;
  options?: Record<string, unknown>;
  required: boolean;
  display_order: number;
}

/** 外部アプリ作成リクエスト */
export interface CreateExternalAppRequest {
  name: string;
  description?: string;
  icon?: string;
  data_source_id: number;
  source_table_name: string;
  fields: CreateExternalFieldRequest[];
}

/** データ型マッピング（外部DBカラム型 → アプリフィールド型） */
export function mapDataTypeToFieldType(dataType: string): string {
  const normalizedType = dataType.toLowerCase();

  // 数値型
  if (
    /^(int|bigint|smallint|tinyint|decimal|numeric|float|double|real|number)/.test(
      normalizedType
    )
  ) {
    return "number";
  }

  // 日付時刻型
  if (/^(datetime|timestamp)/.test(normalizedType)) {
    return "datetime";
  }
  if (/^date$/.test(normalizedType)) {
    return "date";
  }

  // テキスト型（長い）
  if (/^(text|clob|longtext|mediumtext|ntext)/.test(normalizedType)) {
    return "textarea";
  }

  // ブール型
  if (/^(bool|boolean|bit)/.test(normalizedType)) {
    return "checkbox";
  }

  // JSON型
  if (/^json/.test(normalizedType)) {
    return "textarea";
  }

  // デフォルトはテキスト
  return "text";
}
