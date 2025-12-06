import type { DashboardStatsResponse } from "@/types";
import client from "./client";
import type { IDashboardApi } from "./interfaces";

/**
 * ダッシュボードAPI実装
 */
export const dashboardApi: IDashboardApi = {
  // ダッシュボード統計を取得
  getStats: async (): Promise<DashboardStatsResponse> => {
    const response =
      await client.get<DashboardStatsResponse>("/dashboard/stats");
    return response.data;
  },
};
