import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  useBulkCreateRecords,
  useBulkDeleteRecords,
  useCreateRecord,
  useDeleteRecord,
  useRecord,
  useRecords,
  useUpdateRecord,
} from "./useRecords";

// recordsApiのモック
vi.mock("@/api", () => ({
  recordsApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    bulkCreate: vi.fn(),
    bulkDelete: vi.fn(),
  },
}));

import { recordsApi } from "@/api";

const mockRecord = {
  id: 1,
  data: { title: "Test Record" },
  created_by: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

const mockRecordsResponse = {
  records: [mockRecord],
  pagination: {
    total: 1,
    page: 1,
    limit: 20,
    total_pages: 1,
  },
};

describe("useRecords hooks", () => {
  let queryClient: QueryClient;

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          gcTime: 0,
        },
        mutations: {
          retry: false,
        },
      },
    });
    vi.clearAllMocks();
  });

  describe("useRecords", () => {
    it("should fetch records list", async () => {
      vi.mocked(recordsApi.getAll).mockResolvedValueOnce(mockRecordsResponse);

      const { result } = renderHook(() => useRecords(1), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.getAll).toHaveBeenCalledWith(1, {});
      expect(result.current.data).toEqual(mockRecordsResponse);
    });

    it("should fetch records with options", async () => {
      vi.mocked(recordsApi.getAll).mockResolvedValueOnce(mockRecordsResponse);

      const options = {
        page: 2,
        limit: 10,
        sort: "title",
        order: "desc" as const,
      };
      const { result } = renderHook(() => useRecords(1, options), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.getAll).toHaveBeenCalledWith(1, options);
    });

    it("should not fetch when appId is 0", () => {
      const { result } = renderHook(() => useRecords(0), { wrapper });

      expect(result.current.isFetching).toBe(false);
      expect(recordsApi.getAll).not.toHaveBeenCalled();
    });
  });

  describe("useRecord", () => {
    it("should fetch single record", async () => {
      vi.mocked(recordsApi.getById).mockResolvedValueOnce(mockRecord);

      const { result } = renderHook(() => useRecord(1, 1), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.getById).toHaveBeenCalledWith(1, 1);
      expect(result.current.data).toEqual(mockRecord);
    });

    it("should not fetch when recordId is 0", () => {
      const { result } = renderHook(() => useRecord(1, 0), { wrapper });

      expect(result.current.isFetching).toBe(false);
      expect(recordsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe("useCreateRecord", () => {
    it("should create record and invalidate queries", async () => {
      vi.mocked(recordsApi.create).mockResolvedValueOnce(mockRecord);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useCreateRecord(), { wrapper });

      result.current.mutate({
        appId: 1,
        data: { data: { title: "New Record" } },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.create).toHaveBeenCalled();
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["records", 1] });
    });
  });

  describe("useUpdateRecord", () => {
    it("should update record and invalidate queries", async () => {
      const updatedRecord = { ...mockRecord, data: { title: "Updated" } };
      vi.mocked(recordsApi.update).mockResolvedValueOnce(updatedRecord);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useUpdateRecord(), { wrapper });

      result.current.mutate({
        appId: 1,
        recordId: 1,
        data: { data: { title: "Updated" } },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.update).toHaveBeenCalledWith(1, 1, {
        data: { title: "Updated" },
      });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["records", 1] });
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: ["record", 1, 1],
      });
    });
  });

  describe("useDeleteRecord", () => {
    it("should delete record and invalidate queries", async () => {
      vi.mocked(recordsApi.delete).mockResolvedValueOnce(undefined);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useDeleteRecord(), { wrapper });

      result.current.mutate({ appId: 1, recordId: 1 });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.delete).toHaveBeenCalledWith(1, 1);
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["records", 1] });
    });
  });

  describe("useBulkCreateRecords", () => {
    it("should bulk create records and invalidate queries", async () => {
      const response = { records: [mockRecord, { ...mockRecord, id: 2 }] };
      vi.mocked(recordsApi.bulkCreate).mockResolvedValueOnce(response);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useBulkCreateRecords(), { wrapper });

      result.current.mutate({
        appId: 1,
        data: {
          records: [
            { data: { title: "Record 1" } },
            { data: { title: "Record 2" } },
          ],
        },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.bulkCreate).toHaveBeenCalled();
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["records", 1] });
    });
  });

  describe("useBulkDeleteRecords", () => {
    it("should bulk delete records and invalidate queries", async () => {
      vi.mocked(recordsApi.bulkDelete).mockResolvedValueOnce(undefined);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useBulkDeleteRecords(), { wrapper });

      result.current.mutate({
        appId: 1,
        data: { ids: [1, 2, 3] },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(recordsApi.bulkDelete).toHaveBeenCalledWith(1, { ids: [1, 2, 3] });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["records", 1] });
    });
  });
});
