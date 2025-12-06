import { AppView, CreateViewRequest, UpdateViewRequest } from "@/types";
import client from "./client";

/**
 * ビューAPI
 */
export const viewsApi = {
  // アプリIDでビュー一覧を取得
  getByAppId: async (appId: number): Promise<{ views: AppView[] }> => {
    const response = await client.get<{ views: AppView[] }>(
      `/apps/${appId}/views`
    );
    return response.data;
  },

  // ビューを作成
  create: async (appId: number, data: CreateViewRequest): Promise<AppView> => {
    const response = await client.post<AppView>(`/apps/${appId}/views`, data);
    return response.data;
  },

  // ビューを更新
  update: async (
    appId: number,
    viewId: number,
    data: UpdateViewRequest
  ): Promise<AppView> => {
    const response = await client.put<AppView>(
      `/apps/${appId}/views/${viewId}`,
      data
    );
    return response.data;
  },

  // ビューを削除
  delete: async (appId: number, viewId: number): Promise<void> => {
    await client.delete(`/apps/${appId}/views/${viewId}`);
  },
};
