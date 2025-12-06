import {
  ChartConfig,
  ChartDataRequest,
  ChartDataResponse,
  SaveChartConfigRequest,
} from "@/types";
import client from "./client";

/**
 * チャートAPI
 */
export const chartsApi = {
  // チャートデータを取得
  getData: async (
    appId: number,
    data: ChartDataRequest
  ): Promise<ChartDataResponse> => {
    const response = await client.post<ChartDataResponse>(
      `/apps/${appId}/charts/data`,
      data
    );
    return response.data;
  },

  // チャート設定一覧を取得
  getConfigs: async (appId: number): Promise<{ configs: ChartConfig[] }> => {
    const response = await client.get<{ configs: ChartConfig[] }>(
      `/apps/${appId}/charts/config`
    );
    return response.data;
  },

  // チャート設定を保存
  saveConfig: async (
    appId: number,
    data: SaveChartConfigRequest
  ): Promise<ChartConfig> => {
    const response = await client.post<ChartConfig>(
      `/apps/${appId}/charts/config`,
      data
    );
    return response.data;
  },

  // チャート設定を削除
  deleteConfig: async (appId: number, configId: number): Promise<void> => {
    await client.delete(`/apps/${appId}/charts/config/${configId}`);
  },
};
