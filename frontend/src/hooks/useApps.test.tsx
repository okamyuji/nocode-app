import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  useApp,
  useApps,
  useCreateApp,
  useDeleteApp,
  useUpdateApp,
} from "./useApps";

// appsApiのモック
vi.mock("@/api", () => ({
  appsApi: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}));

import { appsApi } from "@/api";

const mockApp = {
  id: 1,
  name: "Test App",
  description: "A test application",
  table_name: "app_data_1",
  icon: "📋",
  created_by: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  fields: [],
  field_count: 0,
};

const mockAppListResponse = {
  apps: [mockApp],
  pagination: {
    total: 1,
    page: 1,
    limit: 20,
    total_pages: 1,
  },
};

describe("useApps hooks", () => {
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

  describe("useApps", () => {
    it("should fetch apps list", async () => {
      vi.mocked(appsApi.getAll).mockResolvedValueOnce(mockAppListResponse);

      const { result } = renderHook(() => useApps(), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.getAll).toHaveBeenCalledWith(1, 20);
      expect(result.current.data).toEqual(mockAppListResponse);
    });

    it("should fetch apps with custom pagination", async () => {
      vi.mocked(appsApi.getAll).mockResolvedValueOnce(mockAppListResponse);

      const { result } = renderHook(() => useApps(2, 10), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.getAll).toHaveBeenCalledWith(2, 10);
    });

    it("should handle error", async () => {
      vi.mocked(appsApi.getAll).mockRejectedValueOnce(
        new Error("Network error")
      );

      const { result } = renderHook(() => useApps(), { wrapper });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toBeDefined();
    });
  });

  describe("useApp", () => {
    it("should fetch single app by id", async () => {
      vi.mocked(appsApi.getById).mockResolvedValueOnce(mockApp);

      const { result } = renderHook(() => useApp(1), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.getById).toHaveBeenCalledWith(1);
      expect(result.current.data).toEqual(mockApp);
    });

    it("should not fetch when id is 0", () => {
      const { result } = renderHook(() => useApp(0), { wrapper });

      expect(result.current.isFetching).toBe(false);
      expect(appsApi.getById).not.toHaveBeenCalled();
    });
  });

  describe("useCreateApp", () => {
    it("should create app and invalidate queries", async () => {
      vi.mocked(appsApi.create).mockResolvedValueOnce(mockApp);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useCreateApp(), { wrapper });

      result.current.mutate({
        name: "New App",
        description: "New description",
        fields: [],
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.create).toHaveBeenCalledWith({
        name: "New App",
        description: "New description",
        fields: [],
      });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["apps"] });
    });
  });

  describe("useUpdateApp", () => {
    it("should update app and invalidate queries", async () => {
      const updatedApp = { ...mockApp, name: "Updated App" };
      vi.mocked(appsApi.update).mockResolvedValueOnce(updatedApp);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useUpdateApp(), { wrapper });

      result.current.mutate({ id: 1, data: { name: "Updated App" } });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.update).toHaveBeenCalledWith(1, { name: "Updated App" });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["apps"] });
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["app", 1] });
    });
  });

  describe("useDeleteApp", () => {
    it("should delete app and invalidate queries", async () => {
      vi.mocked(appsApi.delete).mockResolvedValueOnce(undefined);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

      const { result } = renderHook(() => useDeleteApp(), { wrapper });

      result.current.mutate(1);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.delete).toHaveBeenCalledWith(1);
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ["apps"] });
    });
  });
});
