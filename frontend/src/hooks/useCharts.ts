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
  // queryKeyの安定化のため、requestの主要フィールドのみを使用
  const queryKeyRequest = request
    ? {
        chart_type: request.chart_type,
        x_field: request.x_axis?.field,
        y_field: request.y_axis?.field,
        y_agg: request.y_axis?.aggregation,
      }
    : null;

  return useQuery({
    queryKey: ["chartData", appId, queryKeyRequest],
    queryFn: () => chartsApi.getData(appId, request!),
    // x_axis.fieldが空の場合はリクエストを送信しない
    enabled: !!appId && !!request && !!request.x_axis?.field,
    // 不要な再フェッチ・リトライを防ぐ
    staleTime: 30000, // 30秒間はキャッシュを使用
    refetchOnWindowFocus: false,
    retry: false, // エラー時のリトライを無効化
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
