/**
 * データソースAPI関数
 */

import type { AxiosInstance } from "axios";
import type {
  ColumnListResponse,
  CreateDataSourceRequest,
  DataSource,
  DataSourceListResponse,
  TableListResponse,
  TestConnectionRequest,
  TestConnectionResponse,
  UpdateDataSourceRequest,
} from "../types/datasource";

/**
 * データソースAPI関数を作成する
 */
export function createDataSourceApi(client: AxiosInstance) {
  return {
    /**
     * データソース一覧を取得
     */
    async getDataSources(
      page: number = 1,
      limit: number = 20
    ): Promise<DataSourceListResponse> {
      const response = await client.get<DataSourceListResponse>(
        "/datasources",
        {
          params: { page, limit },
        }
      );
      return response.data;
    },

    /**
     * データソースを取得
     */
    async getDataSource(id: number): Promise<DataSource> {
      const response = await client.get<DataSource>(`/datasources/${id}`);
      return response.data;
    },

    /**
     * データソースを作成
     */
    async createDataSource(data: CreateDataSourceRequest): Promise<DataSource> {
      const response = await client.post<DataSource>("/datasources", data);
      return response.data;
    },

    /**
     * データソースを更新
     */
    async updateDataSource(
      id: number,
      data: UpdateDataSourceRequest
    ): Promise<DataSource> {
      const response = await client.put<DataSource>(`/datasources/${id}`, data);
      return response.data;
    },

    /**
     * データソースを削除
     */
    async deleteDataSource(id: number): Promise<void> {
      await client.delete(`/datasources/${id}`);
    },

    /**
     * テスト接続を実行
     */
    async testConnection(
      data: TestConnectionRequest
    ): Promise<TestConnectionResponse> {
      const response = await client.post<TestConnectionResponse>(
        "/datasources/test",
        data
      );
      return response.data;
    },

    /**
     * テーブル一覧を取得
     */
    async getTables(dataSourceId: number): Promise<TableListResponse> {
      const response = await client.get<TableListResponse>(
        `/datasources/${dataSourceId}/tables`
      );
      return response.data;
    },

    /**
     * カラム一覧を取得
     */
    async getColumns(
      dataSourceId: number,
      tableName: string
    ): Promise<ColumnListResponse> {
      const response = await client.get<ColumnListResponse>(
        `/datasources/${dataSourceId}/tables/${tableName}/columns`
      );
      return response.data;
    },
  };
}

export type DataSourceApi = ReturnType<typeof createDataSourceApi>;
