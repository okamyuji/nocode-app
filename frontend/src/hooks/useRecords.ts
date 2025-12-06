/**
 * レコード操作フック
 */

import { recordsApi } from "@/api";
import {
  BulkCreateRecordRequest,
  BulkDeleteRecordRequest,
  CreateRecordRequest,
  RecordQueryOptions,
  UpdateRecordRequest,
} from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

/**
 * レコード一覧を取得するフック
 */
export function useRecords(appId: number, options: RecordQueryOptions = {}) {
  return useQuery({
    queryKey: ["records", appId, options],
    queryFn: () => recordsApi.getAll(appId, options),
    enabled: !!appId,
  });
}

/**
 * 単一レコードを取得するフック
 */
export function useRecord(appId: number, recordId: number) {
  return useQuery({
    queryKey: ["record", appId, recordId],
    queryFn: () => recordsApi.getById(appId, recordId),
    enabled: !!appId && !!recordId,
  });
}

/**
 * レコードを作成するフック
 */
export function useCreateRecord() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: number;
      data: CreateRecordRequest;
    }) => recordsApi.create(appId, data),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["records", appId] });
    },
  });
}

/**
 * レコードを更新するフック
 */
export function useUpdateRecord() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      recordId,
      data,
    }: {
      appId: number;
      recordId: number;
      data: UpdateRecordRequest;
    }) => recordsApi.update(appId, recordId, data),
    onSuccess: (_, { appId, recordId }) => {
      queryClient.invalidateQueries({ queryKey: ["records", appId] });
      queryClient.invalidateQueries({ queryKey: ["record", appId, recordId] });
    },
  });
}

/**
 * レコードを削除するフック
 */
export function useDeleteRecord() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, recordId }: { appId: number; recordId: number }) =>
      recordsApi.delete(appId, recordId),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["records", appId] });
    },
  });
}

/**
 * レコードを一括作成するフック
 */
export function useBulkCreateRecords() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: number;
      data: BulkCreateRecordRequest;
    }) => recordsApi.bulkCreate(appId, data),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["records", appId] });
    },
  });
}

/**
 * レコードを一括削除するフック
 */
export function useBulkDeleteRecords() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: number;
      data: BulkDeleteRecordRequest;
    }) => recordsApi.bulkDelete(appId, data),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: ["records", appId] });
    },
  });
}
