/**
 * useApiClientフックのテスト
 */

import { renderHook } from "@testing-library/react";
import { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { ApiClientProvider } from "./ApiClientProvider";
import { IApiClient } from "./interfaces";
import { useApiClient } from "./useApiClient";

// テスト用モックAPIクライアント
const mockApiClient: IApiClient = {
  auth: {
    register: vi.fn(),
    login: vi.fn(),
    me: vi.fn(),
    refresh: vi.fn(),
  },
  apps: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
  fields: {
    getByAppId: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    updateOrder: vi.fn(),
  },
  records: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    bulkCreate: vi.fn(),
    bulkDelete: vi.fn(),
  },
  views: {
    getByAppId: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
  charts: {
    getData: vi.fn(),
    getConfigs: vi.fn(),
    saveConfig: vi.fn(),
    deleteConfig: vi.fn(),
  },
  users: {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
  profile: {
    updateProfile: vi.fn(),
    changePassword: vi.fn(),
  },
  dashboard: {
    getStats: vi.fn(),
  },
  dashboardWidgets: {
    getAll: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    reorder: vi.fn(),
    toggleVisibility: vi.fn(),
  },
  dataSources: {
    getDataSources: vi.fn(),
    getDataSource: vi.fn(),
    createDataSource: vi.fn(),
    updateDataSource: vi.fn(),
    deleteDataSource: vi.fn(),
    testConnection: vi.fn(),
    getTables: vi.fn(),
    getColumns: vi.fn(),
  },
};

describe("useApiClient", () => {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <ApiClientProvider client={mockApiClient}>{children}</ApiClientProvider>
  );

  it("注入されたAPIクライアントを返す", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current).toBe(mockApiClient);
  });

  it("認証APIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.auth).toBeDefined();
    expect(result.current.auth.login).toBeDefined();
    expect(result.current.auth.register).toBeDefined();
    expect(result.current.auth.me).toBeDefined();
    expect(result.current.auth.refresh).toBeDefined();
  });

  it("アプリAPIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.apps).toBeDefined();
    expect(result.current.apps.getAll).toBeDefined();
    expect(result.current.apps.getById).toBeDefined();
    expect(result.current.apps.create).toBeDefined();
    expect(result.current.apps.update).toBeDefined();
    expect(result.current.apps.delete).toBeDefined();
  });

  it("フィールドAPIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.fields).toBeDefined();
    expect(result.current.fields.getByAppId).toBeDefined();
    expect(result.current.fields.create).toBeDefined();
    expect(result.current.fields.update).toBeDefined();
    expect(result.current.fields.delete).toBeDefined();
    expect(result.current.fields.updateOrder).toBeDefined();
  });

  it("レコードAPIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.records).toBeDefined();
    expect(result.current.records.getAll).toBeDefined();
    expect(result.current.records.getById).toBeDefined();
    expect(result.current.records.create).toBeDefined();
    expect(result.current.records.update).toBeDefined();
    expect(result.current.records.delete).toBeDefined();
    expect(result.current.records.bulkCreate).toBeDefined();
    expect(result.current.records.bulkDelete).toBeDefined();
  });

  it("ビューAPIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.views).toBeDefined();
    expect(result.current.views.getByAppId).toBeDefined();
    expect(result.current.views.create).toBeDefined();
    expect(result.current.views.update).toBeDefined();
    expect(result.current.views.delete).toBeDefined();
  });

  it("チャートAPIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.charts).toBeDefined();
    expect(result.current.charts.getData).toBeDefined();
    expect(result.current.charts.getConfigs).toBeDefined();
    expect(result.current.charts.saveConfig).toBeDefined();
    expect(result.current.charts.deleteConfig).toBeDefined();
  });

  it("ダッシュボードAPIにアクセスできる", () => {
    const { result } = renderHook(() => useApiClient(), { wrapper });

    expect(result.current.dashboard).toBeDefined();
    expect(result.current.dashboard.getStats).toBeDefined();
  });
});
