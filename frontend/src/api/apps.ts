import {
  App,
  AppListResponse,
  CreateAppRequest,
  UpdateAppRequest,
} from "@/types";
import client from "./client";

/**
 * アプリAPI
 */
export const appsApi = {
  // アプリ一覧を取得
  getAll: async (page = 1, limit = 20): Promise<AppListResponse> => {
    const response = await client.get<AppListResponse>("/apps", {
      params: { page, limit },
    });
    return response.data;
  },

  // IDでアプリを取得
  getById: async (id: number): Promise<App> => {
    const response = await client.get<App>(`/apps/${id}`);
    return response.data;
  },

  // アプリを作成
  create: async (data: CreateAppRequest): Promise<App> => {
    const response = await client.post<App>("/apps", data);
    return response.data;
  },

  // アプリを更新
  update: async (id: number, data: UpdateAppRequest): Promise<App> => {
    const response = await client.put<App>(`/apps/${id}`, data);
    return response.data;
  },

  // アプリを削除
  delete: async (id: number): Promise<void> => {
    await client.delete(`/apps/${id}`);
  },
};
