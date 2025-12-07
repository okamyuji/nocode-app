/**
 * アプリ操作フック
 */

import { appsApi } from "@/api";
import { AppListResponse, CreateAppRequest, UpdateAppRequest } from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

/**
 * アプリ一覧を取得するフック
 */
export function useApps(page = 1, limit = 20) {
  return useQuery({
    queryKey: ["apps", page, limit],
    queryFn: () => appsApi.getAll(page, limit),
  });
}

/**
 * 単一アプリを取得するフック
 */
export function useApp(id: number) {
  return useQuery({
    queryKey: ["app", id],
    queryFn: () => appsApi.getById(id),
    enabled: !!id,
  });
}

/**
 * アプリを作成するフック
 */
export function useCreateApp() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAppRequest) => appsApi.create(data),
    onSuccess: (newApp) => {
      // キャッシュを直接更新して即時反映
      queryClient.setQueriesData<AppListResponse>(
        { queryKey: ["apps"] },
        (oldData) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            apps: [newApp, ...oldData.apps],
            pagination: {
              ...oldData.pagination,
              total: oldData.pagination.total + 1,
            },
          };
        }
      );
      // 個別アプリのキャッシュも設定
      queryClient.setQueryData(["app", newApp.id], newApp);
    },
  });
}

/**
 * アプリを更新するフック
 */
export function useUpdateApp() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateAppRequest }) =>
      appsApi.update(id, data),
    onSuccess: (updatedApp, { id }) => {
      // キャッシュを直接更新して即時反映
      queryClient.setQueriesData<AppListResponse>(
        { queryKey: ["apps"] },
        (oldData) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            apps: oldData.apps.map((app) => (app.id === id ? updatedApp : app)),
          };
        }
      );
      // 個別アプリのキャッシュも更新
      queryClient.setQueryData(["app", id], updatedApp);
    },
  });
}

/**
 * アプリを削除するフック
 */
export function useDeleteApp() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => appsApi.delete(id),
    onSuccess: (_, deletedId) => {
      // キャッシュを直接更新して即時反映
      queryClient.setQueriesData<AppListResponse>(
        { queryKey: ["apps"] },
        (oldData) => {
          if (!oldData) return oldData;
          return {
            ...oldData,
            apps: oldData.apps.filter((app) => app.id !== deletedId),
            pagination: {
              ...oldData.pagination,
              total: Math.max(0, oldData.pagination.total - 1),
            },
          };
        }
      );
      // 個別アプリのキャッシュを削除
      queryClient.removeQueries({ queryKey: ["app", deletedId] });
    },
  });
}
