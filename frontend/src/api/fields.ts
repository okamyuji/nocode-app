import {
  CreateFieldRequest,
  Field,
  UpdateFieldOrderRequest,
  UpdateFieldRequest,
} from "@/types";
import client from "./client";

/**
 * フィールドAPI
 */
export const fieldsApi = {
  // アプリIDでフィールド一覧を取得
  getByAppId: async (appId: number): Promise<{ fields: Field[] }> => {
    const response = await client.get<{ fields: Field[] }>(
      `/apps/${appId}/fields`
    );
    return response.data;
  },

  // フィールドを作成
  create: async (appId: number, data: CreateFieldRequest): Promise<Field> => {
    const response = await client.post<Field>(`/apps/${appId}/fields`, data);
    return response.data;
  },

  // フィールドを更新
  update: async (
    appId: number,
    fieldId: number,
    data: UpdateFieldRequest
  ): Promise<Field> => {
    const response = await client.put<Field>(
      `/apps/${appId}/fields/${fieldId}`,
      data
    );
    return response.data;
  },

  // フィールドを削除
  delete: async (appId: number, fieldId: number): Promise<void> => {
    await client.delete(`/apps/${appId}/fields/${fieldId}`);
  },

  // フィールドの表示順序を更新
  updateOrder: async (
    appId: number,
    data: UpdateFieldOrderRequest
  ): Promise<void> => {
    await client.put(`/apps/${appId}/fields/order`, data);
  },
};
