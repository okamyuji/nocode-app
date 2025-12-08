import type {
  CreateDashboardWidgetRequest,
  DashboardWidget,
  DashboardWidgetListResponse,
  ReorderWidgetsRequest,
  UpdateDashboardWidgetRequest,
} from "@/types";
import client from "./client";
import type { IDashboardWidgetsApi } from "./interfaces";

/**
 * ダッシュボードウィジェットAPI実装
 */
export const dashboardWidgetsApi: IDashboardWidgetsApi = {
  // ダッシュボードウィジェット一覧を取得
  getAll: async (
    visibleOnly?: boolean
  ): Promise<DashboardWidgetListResponse> => {
    const params = visibleOnly ? { visible: "true" } : {};
    const response = await client.get<DashboardWidgetListResponse>(
      "/dashboard/widgets",
      { params }
    );
    return response.data;
  },

  // ダッシュボードウィジェットを作成
  create: async (
    data: CreateDashboardWidgetRequest
  ): Promise<DashboardWidget> => {
    const response = await client.post<DashboardWidget>(
      "/dashboard/widgets",
      data
    );
    return response.data;
  },

  // ダッシュボードウィジェットを更新
  update: async (
    widgetId: number,
    data: UpdateDashboardWidgetRequest
  ): Promise<DashboardWidget> => {
    const response = await client.put<DashboardWidget>(
      `/dashboard/widgets/${widgetId}`,
      data
    );
    return response.data;
  },

  // ダッシュボードウィジェットを削除
  delete: async (widgetId: number): Promise<void> => {
    await client.delete(`/dashboard/widgets/${widgetId}`);
  },

  // ウィジェットの並び替え
  reorder: async (
    data: ReorderWidgetsRequest
  ): Promise<{ message: string }> => {
    const response = await client.put<{ message: string }>(
      "/dashboard/widgets/reorder",
      data
    );
    return response.data;
  },

  // 表示/非表示の切り替え
  toggleVisibility: async (widgetId: number): Promise<DashboardWidget> => {
    const response = await client.post<DashboardWidget>(
      `/dashboard/widgets/${widgetId}/toggle`
    );
    return response.data;
  },
};
