/**
 * フィールド操作フック
 */

import { fieldsApi } from "@/api";
import { useAppStore } from "@/stores";
import {
  CreateFieldRequest,
  UpdateFieldOrderRequest,
  UpdateFieldRequest,
} from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

/**
 * アプリのフィールド一覧を取得するフック
 */
export function useFields(appId: number) {
  const { setCurrentFields } = useAppStore();

  return useQuery({
    queryKey: ["fields", appId],
    queryFn: async () => {
      const response = await fieldsApi.getByAppId(appId);
      setCurrentFields(response.fields);
      return response;
    },
    enabled: !!appId,
  });
}

/**
 * フィールドを作成するフック
 */
export function useCreateField() {
  const queryClient = useQueryClient();
  const { addField } = useAppStore();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: number;
      data: CreateFieldRequest;
    }) => fieldsApi.create(appId, data),
    onSuccess: (field, { appId }) => {
      addField(field);
      queryClient.invalidateQueries({ queryKey: ["fields", appId] });
      queryClient.invalidateQueries({ queryKey: ["app", appId] });
    },
  });
}

/**
 * フィールドを更新するフック
 */
export function useUpdateField() {
  const queryClient = useQueryClient();
  const { updateField } = useAppStore();

  return useMutation({
    mutationFn: ({
      appId,
      fieldId,
      data,
    }: {
      appId: number;
      fieldId: number;
      data: UpdateFieldRequest;
    }) => fieldsApi.update(appId, fieldId, data),
    onSuccess: (field, { appId }) => {
      updateField(field);
      queryClient.invalidateQueries({ queryKey: ["fields", appId] });
    },
  });
}

/**
 * フィールドを削除するフック
 */
export function useDeleteField() {
  const queryClient = useQueryClient();
  const { removeField } = useAppStore();

  return useMutation({
    mutationFn: ({ appId, fieldId }: { appId: number; fieldId: number }) =>
      fieldsApi.delete(appId, fieldId),
    onSuccess: (_, { appId, fieldId }) => {
      removeField(fieldId);
      queryClient.invalidateQueries({ queryKey: ["fields", appId] });
      queryClient.invalidateQueries({ queryKey: ["app", appId] });
    },
  });
}

/**
 * フィールドの表示順序を更新するフック
 */
export function useUpdateFieldOrder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: number;
      data: UpdateFieldOrderRequest;
    }) => fieldsApi.updateOrder(appId, data),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["fields", appId] });
    },
  });
}
