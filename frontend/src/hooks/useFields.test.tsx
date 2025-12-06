import { useAppStore } from "@/stores";
import { FieldType } from "@/types";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  useCreateField,
  useDeleteField,
  useFields,
  useUpdateField,
  useUpdateFieldOrder,
} from "./useFields";

// fieldsApiのモック
vi.mock("@/api", () => ({
  fieldsApi: {
    getByAppId: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    updateOrder: vi.fn(),
  },
}));

import { fieldsApi } from "@/api";

const mockField = {
  id: 1,
  app_id: 1,
  field_code: "title",
  field_name: "Title",
  field_type: "text" as FieldType,
  required: true,
  display_order: 1,
  options: {},
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

const mockFieldsResponse = {
  fields: [mockField],
};

describe("useFields hooks", () => {
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
    useAppStore.setState({
      currentApp: null,
      currentFields: [],
    });
    vi.clearAllMocks();
  });

  describe("useFields", () => {
    it("should fetch fields and update store", async () => {
      vi.mocked(fieldsApi.getByAppId).mockResolvedValueOnce(mockFieldsResponse);

      const { result } = renderHook(() => useFields(1), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(fieldsApi.getByAppId).toHaveBeenCalledWith(1);
      expect(result.current.data).toEqual(mockFieldsResponse);
      expect(useAppStore.getState().currentFields).toHaveLength(1);
    });

    it("should not fetch when appId is 0", () => {
      const { result } = renderHook(() => useFields(0), { wrapper });

      expect(result.current.isFetching).toBe(false);
      expect(fieldsApi.getByAppId).not.toHaveBeenCalled();
    });
  });

  describe("useCreateField", () => {
    it("should create field, update store, and invalidate queries", async () => {
      vi.mocked(fieldsApi.create).mockResolvedValueOnce(mockField);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useCreateField(), { wrapper });

      result.current.mutate({
        appId: 1,
        data: {
          field_code: "title",
          field_name: "Title",
          field_type: "text",
          required: true,
        },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(fieldsApi.create).toHaveBeenCalled();
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["fields", 1] });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["app", 1] });
      expect(useAppStore.getState().currentFields).toContainEqual(mockField);
    });
  });

  describe("useUpdateField", () => {
    it("should update field and update store", async () => {
      // Pre-populate store with the field
      useAppStore.setState({ currentFields: [mockField] });

      const updatedField = { ...mockField, field_name: "Updated Title" };
      vi.mocked(fieldsApi.update).mockResolvedValueOnce(updatedField);

      const { result } = renderHook(() => useUpdateField(), { wrapper });

      result.current.mutate({
        appId: 1,
        fieldId: 1,
        data: { field_name: "Updated Title" },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(fieldsApi.update).toHaveBeenCalledWith(1, 1, {
        field_name: "Updated Title",
      });
      const storeField = useAppStore
        .getState()
        .currentFields.find((f) => f.id === 1);
      expect(storeField?.field_name).toBe("Updated Title");
    });
  });

  describe("useDeleteField", () => {
    it("should delete field and remove from store", async () => {
      // Pre-populate store with the field
      useAppStore.setState({ currentFields: [mockField] });

      vi.mocked(fieldsApi.delete).mockResolvedValueOnce(undefined);

      const { result } = renderHook(() => useDeleteField(), { wrapper });

      result.current.mutate({ appId: 1, fieldId: 1 });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(fieldsApi.delete).toHaveBeenCalledWith(1, 1);
      expect(useAppStore.getState().currentFields).toHaveLength(0);
    });
  });

  describe("useUpdateFieldOrder", () => {
    it("should update field order and invalidate queries", async () => {
      vi.mocked(fieldsApi.updateOrder).mockResolvedValueOnce(undefined);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useUpdateFieldOrder(), { wrapper });

      result.current.mutate({
        appId: 1,
        data: { fields: [{ id: 1, display_order: 1 }] },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(fieldsApi.updateOrder).toHaveBeenCalledWith(1, {
        fields: [{ id: 1, display_order: 1 }],
      });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["fields", 1] });
    });
  });
});
