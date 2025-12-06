/**
 * アプリ操作フック
 */

import { appsApi } from "@/api";
import { CreateAppRequest, UpdateAppRequest } from "@/types";
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
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
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
      queryClient.invalidateQueries({ queryKey: ["app", id] });
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["apps"] });
    },
  });
}
