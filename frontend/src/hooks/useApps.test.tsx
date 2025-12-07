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
import type { App, AppListResponse } from "@/types";

const mockApp: App = {
  id: 1,
  name: "Test App",
  description: "A test application",
  table_name: "app_data_1",
  icon: "default",
  is_external: false,
  created_by: 1,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  fields: [],
  field_count: 0,
};

const mockAppListResponse: AppListResponse = {
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
          gcTime: Infinity, // テスト中はキャッシュを保持
          staleTime: Infinity,
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
    it("should create app and add to apps cache", async () => {
      const newApp: App = { ...mockApp, id: 2, name: "New App" };
      vi.mocked(appsApi.create).mockResolvedValueOnce(newApp);

      // 事前にキャッシュを設定
      queryClient.setQueryData(["apps", 1, 20], mockAppListResponse);

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

      // appsキャッシュが更新されていることを確認
      const cachedApps = queryClient.getQueryData<AppListResponse>([
        "apps",
        1,
        20,
      ]);
      expect(cachedApps).toBeDefined();
      expect(cachedApps!.apps).toHaveLength(2);
      expect(cachedApps!.apps[0]).toEqual(newApp); // 新しいアプリが先頭
      expect(cachedApps!.apps[1]).toEqual(mockApp);
      expect(cachedApps!.pagination.total).toBe(2);

      // 個別アプリのキャッシュが設定されていることを確認
      const cachedApp = queryClient.getQueryData<App>(["app", 2]);
      expect(cachedApp).toEqual(newApp);
    });

    it("should handle empty cache gracefully", async () => {
      const newApp: App = { ...mockApp, id: 2, name: "New App" };
      vi.mocked(appsApi.create).mockResolvedValueOnce(newApp);

      // キャッシュなしの状態でmutate
      const { result } = renderHook(() => useCreateApp(), { wrapper });

      result.current.mutate({
        name: "New App",
        description: "",
        fields: [],
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      // 個別アプリのキャッシュは設定される
      const cachedApp = queryClient.getQueryData<App>(["app", 2]);
      expect(cachedApp).toEqual(newApp);
    });
  });

  describe("useUpdateApp", () => {
    it("should update app and reflect in apps cache", async () => {
      const updatedApp: App = {
        ...mockApp,
        name: "Updated App",
        icon: "calendar",
      };
      vi.mocked(appsApi.update).mockResolvedValueOnce(updatedApp);

      // 事前にキャッシュを設定
      queryClient.setQueryData(["apps", 1, 20], mockAppListResponse);
      queryClient.setQueryData(["app", 1], mockApp);

      const { result } = renderHook(() => useUpdateApp(), { wrapper });

      result.current.mutate({
        id: 1,
        data: { name: "Updated App", icon: "calendar" },
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.update).toHaveBeenCalledWith(1, {
        name: "Updated App",
        icon: "calendar",
      });

      // appsキャッシュが更新されていることを確認
      const cachedApps = queryClient.getQueryData<AppListResponse>([
        "apps",
        1,
        20,
      ]);
      expect(cachedApps).toBeDefined();
      expect(cachedApps!.apps).toHaveLength(1);
      expect(cachedApps!.apps[0].name).toBe("Updated App");
      expect(cachedApps!.apps[0].icon).toBe("calendar");

      // 個別アプリのキャッシュが更新されていることを確認
      const cachedApp = queryClient.getQueryData<App>(["app", 1]);
      expect(cachedApp).toEqual(updatedApp);
    });

    it("should update correct app when multiple apps in cache", async () => {
      const app2: App = { ...mockApp, id: 2, name: "App 2" };
      const app3: App = { ...mockApp, id: 3, name: "App 3" };
      const initialCache: AppListResponse = {
        apps: [mockApp, app2, app3],
        pagination: { total: 3, page: 1, limit: 20, total_pages: 1 },
      };

      const updatedApp2: App = { ...app2, name: "Updated App 2" };
      vi.mocked(appsApi.update).mockResolvedValueOnce(updatedApp2);

      queryClient.setQueryData(["apps", 1, 20], initialCache);

      const { result } = renderHook(() => useUpdateApp(), { wrapper });

      result.current.mutate({ id: 2, data: { name: "Updated App 2" } });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      const cachedApps = queryClient.getQueryData<AppListResponse>([
        "apps",
        1,
        20,
      ]);
      expect(cachedApps!.apps[0].name).toBe("Test App"); // 変更なし
      expect(cachedApps!.apps[1].name).toBe("Updated App 2"); // 更新
      expect(cachedApps!.apps[2].name).toBe("App 3"); // 変更なし
    });
  });

  describe("useDeleteApp", () => {
    it("should delete app and remove from apps cache", async () => {
      vi.mocked(appsApi.delete).mockResolvedValueOnce(undefined);

      // 事前にキャッシュを設定
      queryClient.setQueryData(["apps", 1, 20], mockAppListResponse);
      queryClient.setQueryData(["app", 1], mockApp);

      const { result } = renderHook(() => useDeleteApp(), { wrapper });

      result.current.mutate(1);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(appsApi.delete).toHaveBeenCalledWith(1);

      // appsキャッシュからアプリが削除されていることを確認
      const cachedApps = queryClient.getQueryData<AppListResponse>([
        "apps",
        1,
        20,
      ]);
      expect(cachedApps).toBeDefined();
      expect(cachedApps!.apps).toHaveLength(0);
      expect(cachedApps!.pagination.total).toBe(0);

      // 個別アプリのキャッシュが削除されていることを確認
      const cachedApp = queryClient.getQueryData<App>(["app", 1]);
      expect(cachedApp).toBeUndefined();
    });

    it("should delete correct app when multiple apps in cache", async () => {
      const app2: App = { ...mockApp, id: 2, name: "App 2" };
      const app3: App = { ...mockApp, id: 3, name: "App 3" };
      const initialCache: AppListResponse = {
        apps: [mockApp, app2, app3],
        pagination: { total: 3, page: 1, limit: 20, total_pages: 1 },
      };

      vi.mocked(appsApi.delete).mockResolvedValueOnce(undefined);

      queryClient.setQueryData(["apps", 1, 20], initialCache);

      const { result } = renderHook(() => useDeleteApp(), { wrapper });

      result.current.mutate(2);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      const cachedApps = queryClient.getQueryData<AppListResponse>([
        "apps",
        1,
        20,
      ]);
      expect(cachedApps!.apps).toHaveLength(2);
      expect(cachedApps!.apps.find((a) => a.id === 2)).toBeUndefined();
      expect(cachedApps!.apps[0].id).toBe(1);
      expect(cachedApps!.apps[1].id).toBe(3);
      expect(cachedApps!.pagination.total).toBe(2);
    });

    it("should not go below zero for pagination total", async () => {
      const emptyCache: AppListResponse = {
        apps: [],
        pagination: { total: 0, page: 1, limit: 20, total_pages: 0 },
      };

      vi.mocked(appsApi.delete).mockResolvedValueOnce(undefined);

      queryClient.setQueryData(["apps", 1, 20], emptyCache);

      const { result } = renderHook(() => useDeleteApp(), { wrapper });

      result.current.mutate(999);

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      const cachedApps = queryClient.getQueryData<AppListResponse>([
        "apps",
        1,
        20,
      ]);
      expect(cachedApps!.pagination.total).toBe(0); // 負にならない
    });
  });
});
