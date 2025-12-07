/**
 * チャート操作フック
 */

import { chartsApi } from "@/api";
import { ChartDataRequest, SaveChartConfigRequest } from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

/**
 * チャートデータを取得するフック
 */
export function useChartData(appId: number, request: ChartDataRequest | null) {
  return useQuery({
    queryKey: ["chartData", appId, request],
    queryFn: () => chartsApi.getData(appId, request!),
    // x_axis.fieldが空の場合はリクエストを送信しない
    enabled: !!appId && !!request && !!request.x_axis?.field,
  });
}

/**
 * チャート設定一覧を取得するフック
 */
export function useChartConfigs(appId: number) {
  return useQuery({
    queryKey: ["chartConfigs", appId],
    queryFn: () => chartsApi.getConfigs(appId),
    enabled: !!appId,
  });
}

/**
 * チャート設定を保存するフック
 */
export function useSaveChartConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: number;
      data: SaveChartConfigRequest;
    }) => chartsApi.saveConfig(appId, data),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["chartConfigs", appId] });
    },
  });
}

/**
 * チャート設定を削除するフック
 */
export function useDeleteChartConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, configId }: { appId: number; configId: number }) =>
      chartsApi.deleteConfig(appId, configId),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["chartConfigs", appId] });
    },
  });
}
