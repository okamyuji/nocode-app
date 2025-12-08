/**
 * ダッシュボードウィジェット用フック
 */

import { useDashboardWidgetsApi } from "@/api";
import type {
  CreateDashboardWidgetRequest,
  DashboardWidgetListResponse,
  ReorderWidgetsRequest,
  UpdateDashboardWidgetRequest,
} from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

/**
 * ダッシュボードウィジェット一覧を取得するフック
 */
export function useDashboardWidgets(visibleOnly: boolean = false) {
  const dashboardWidgetsApi = useDashboardWidgetsApi();

  return useQuery<DashboardWidgetListResponse>({
    queryKey: ["dashboard", "widgets", visibleOnly ? "visible" : "all"],
    queryFn: () => dashboardWidgetsApi.getAll(visibleOnly),
    staleTime: 0, // 常に新鮮なデータを取得
    refetchOnMount: true,
  });
}

/**
 * ダッシュボードウィジェットを作成するフック
 */
export function useCreateDashboardWidget() {
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateDashboardWidgetRequest) =>
      dashboardWidgetsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
    },
  });
}

/**
 * ダッシュボードウィジェットを更新するフック
 */
export function useUpdateDashboardWidget() {
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      widgetId,
      data,
    }: {
      widgetId: number;
      data: UpdateDashboardWidgetRequest;
    }) => dashboardWidgetsApi.update(widgetId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
    },
  });
}

/**
 * ダッシュボードウィジェットを削除するフック
 */
export function useDeleteDashboardWidget() {
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (widgetId: number) => dashboardWidgetsApi.delete(widgetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
    },
  });
}

/**
 * ダッシュボードウィジェットを並び替えるフック
 */
export function useReorderDashboardWidgets() {
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: ReorderWidgetsRequest) =>
      dashboardWidgetsApi.reorder(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
    },
  });
}

/**
 * ダッシュボードウィジェットの表示/非表示を切り替えるフック
 */
export function useToggleDashboardWidgetVisibility() {
  const dashboardWidgetsApi = useDashboardWidgetsApi();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (widgetId: number) =>
      dashboardWidgetsApi.toggleVisibility(widgetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard", "widgets"] });
    },
  });
}
