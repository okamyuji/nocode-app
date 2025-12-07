/**
 * APIクライアント インターフェース
 * 依存性注入とテスト用のインターフェース定義
 * 各APIモジュールのコントラクトを定義し、ユニットテストでモック化を可能にする
 */

import {
  App,
  AppListResponse,
  AppView,
  AuthResponse,
  BulkCreateRecordRequest,
  BulkDeleteRecordRequest,
  ChangePasswordRequest,
  ChartConfig,
  ChartDataRequest,
  ChartDataResponse,
  ColumnListResponse,
  CreateAppRequest,
  CreateDataSourceRequest,
  CreateFieldRequest,
  CreateRecordRequest,
  CreateUserRequest,
  CreateViewRequest,
  DashboardStatsResponse,
  DataSource,
  DataSourceListResponse,
  Field,
  LoginRequest,
  RecordItem,
  RecordListResponse,
  RecordQueryOptions,
  RegisterRequest,
  SaveChartConfigRequest,
  TableListResponse,
  TestConnectionRequest,
  TestConnectionResponse,
  UpdateAppRequest,
  UpdateDataSourceRequest,
  UpdateFieldOrderRequest,
  UpdateFieldRequest,
  UpdateProfileRequest,
  UpdateRecordRequest,
  UpdateUserRequest,
  UpdateViewRequest,
  User,
  UserListResponse,
} from "@/types";

/**
 * 認証APIインターフェース
 */
export interface IAuthApi {
  register(data: RegisterRequest): Promise<AuthResponse>;
  login(data: LoginRequest): Promise<AuthResponse>;
  me(): Promise<User>;
  refresh(): Promise<{ token: string }>;
}

/**
 * アプリAPIインターフェース
 */
export interface IAppsApi {
  getAll(page?: number, limit?: number): Promise<AppListResponse>;
  getById(id: number): Promise<App>;
  create(data: CreateAppRequest): Promise<App>;
  update(id: number, data: UpdateAppRequest): Promise<App>;
  delete(id: number): Promise<void>;
}

/**
 * フィールドAPIインターフェース
 */
export interface IFieldsApi {
  getByAppId(appId: number): Promise<{ fields: Field[] }>;
  create(appId: number, data: CreateFieldRequest): Promise<Field>;
  update(
    appId: number,
    fieldId: number,
    data: UpdateFieldRequest
  ): Promise<Field>;
  delete(appId: number, fieldId: number): Promise<void>;
  updateOrder(appId: number, data: UpdateFieldOrderRequest): Promise<void>;
}

/**
 * レコードAPIインターフェース
 */
export interface IRecordsApi {
  getAll(
    appId: number,
    options?: RecordQueryOptions
  ): Promise<RecordListResponse>;
  getById(appId: number, recordId: number): Promise<RecordItem>;
  create(appId: number, data: CreateRecordRequest): Promise<RecordItem>;
  update(
    appId: number,
    recordId: number,
    data: UpdateRecordRequest
  ): Promise<RecordItem>;
  delete(appId: number, recordId: number): Promise<void>;
  bulkCreate(
    appId: number,
    data: BulkCreateRecordRequest
  ): Promise<{ records: RecordItem[] }>;
  bulkDelete(appId: number, data: BulkDeleteRecordRequest): Promise<void>;
}

/**
 * ビューAPIインターフェース
 */
export interface IViewsApi {
  getByAppId(appId: number): Promise<{ views: AppView[] }>;
  create(appId: number, data: CreateViewRequest): Promise<AppView>;
  update(
    appId: number,
    viewId: number,
    data: UpdateViewRequest
  ): Promise<AppView>;
  delete(appId: number, viewId: number): Promise<void>;
}

/**
 * チャートAPIインターフェース
 */
export interface IChartsApi {
  getData(appId: number, data: ChartDataRequest): Promise<ChartDataResponse>;
  getConfigs(appId: number): Promise<{ configs: ChartConfig[] }>;
  saveConfig(appId: number, data: SaveChartConfigRequest): Promise<ChartConfig>;
  deleteConfig(appId: number, configId: number): Promise<void>;
}

/**
 * ユーザーAPIインターフェース（管理者専用）
 */
export interface IUsersApi {
  getAll(page?: number, limit?: number): Promise<UserListResponse>;
  getById(id: number): Promise<User>;
  create(data: CreateUserRequest): Promise<User>;
  update(id: number, data: UpdateUserRequest): Promise<User>;
  delete(id: number): Promise<void>;
}

/**
 * プロフィールAPIインターフェース
 */
export interface IProfileApi {
  updateProfile(data: UpdateProfileRequest): Promise<User>;
  changePassword(data: ChangePasswordRequest): Promise<{ message: string }>;
}

/**
 * ダッシュボードAPIインターフェース
 */
export interface IDashboardApi {
  getStats(): Promise<DashboardStatsResponse>;
}

/**
 * データソースAPIインターフェース
 */
export interface IDataSourcesApi {
  getDataSources(
    page?: number,
    limit?: number
  ): Promise<DataSourceListResponse>;
  getDataSource(id: number): Promise<DataSource>;
  createDataSource(data: CreateDataSourceRequest): Promise<DataSource>;
  updateDataSource(
    id: number,
    data: UpdateDataSourceRequest
  ): Promise<DataSource>;
  deleteDataSource(id: number): Promise<void>;
  testConnection(data: TestConnectionRequest): Promise<TestConnectionResponse>;
  getTables(dataSourceId: number): Promise<TableListResponse>;
  getColumns(
    dataSourceId: number,
    tableName: string
  ): Promise<ColumnListResponse>;
}

/**
 * 統合APIクライアント インターフェース
 */
export interface IApiClient {
  auth: IAuthApi;
  apps: IAppsApi;
  fields: IFieldsApi;
  records: IRecordsApi;
  views: IViewsApi;
  charts: IChartsApi;
  users: IUsersApi;
  profile: IProfileApi;
  dashboard: IDashboardApi;
  dataSources: IDataSourcesApi;
}
