import {
  BulkCreateRecordRequest,
  BulkDeleteRecordRequest,
  CreateRecordRequest,
  RecordItem,
  RecordListResponse,
  RecordQueryOptions,
  UpdateRecordRequest,
} from "@/types";
import client from "./client";

/**
 * レコードAPI
 */
export const recordsApi = {
  // レコード一覧を取得
  getAll: async (
    appId: number,
    options: RecordQueryOptions = {}
  ): Promise<RecordListResponse> => {
    const params: Record<string, string | number> = {
      page: options.page || 1,
      limit: options.limit || 20,
    };

    if (options.sort) {
      params.sort = options.sort;
    }
    if (options.order) {
      params.order = options.order;
    }

    // フィルターをクエリ文字列形式に変換: filter=field:op:value
    const filterParams: string[] = [];
    if (options.filters) {
      options.filters.forEach((f) => {
        filterParams.push(`${f.field}:${f.operator}:${f.value}`);
      });
    }

    const queryString = new URLSearchParams(
      params as Record<string, string>
    ).toString();
    const filterString = filterParams.map((f) => `filter=${f}`).join("&");
    const url = `/apps/${appId}/records?${queryString}${filterString ? "&" + filterString : ""}`;

    const response = await client.get<RecordListResponse>(url);
    return response.data;
  },

  // IDでレコードを取得
  getById: async (appId: number, recordId: number): Promise<RecordItem> => {
    const response = await client.get<RecordItem>(
      `/apps/${appId}/records/${recordId}`
    );
    return response.data;
  },

  // レコードを作成
  create: async (
    appId: number,
    data: CreateRecordRequest
  ): Promise<RecordItem> => {
    const response = await client.post<RecordItem>(
      `/apps/${appId}/records`,
      data
    );
    return response.data;
  },

  // レコードを更新
  update: async (
    appId: number,
    recordId: number,
    data: UpdateRecordRequest
  ): Promise<RecordItem> => {
    const response = await client.put<RecordItem>(
      `/apps/${appId}/records/${recordId}`,
      data
    );
    return response.data;
  },

  // レコードを削除
  delete: async (appId: number, recordId: number): Promise<void> => {
    await client.delete(`/apps/${appId}/records/${recordId}`);
  },

  // レコードを一括作成
  bulkCreate: async (
    appId: number,
    data: BulkCreateRecordRequest
  ): Promise<{ records: RecordItem[] }> => {
    const response = await client.post<{ records: RecordItem[] }>(
      `/apps/${appId}/records/bulk`,
      data
    );
    return response.data;
  },

  // レコードを一括削除
  bulkDelete: async (
    appId: number,
    data: BulkDeleteRecordRequest
  ): Promise<void> => {
    await client.delete(`/apps/${appId}/records/bulk`, { data });
  },
};
