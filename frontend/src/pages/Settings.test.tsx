/**
 * 設定ページのテスト
 */

import { ApiClientContext } from "@/api/ApiClientContext";
import type { IApiClient } from "@/api/interfaces";
import { useAuthStore } from "@/stores";
import { theme } from "@/theme";
import { ChakraProvider } from "@chakra-ui/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BrowserRouter } from "react-router-dom";
import { vi } from "vitest";
import { SettingsPage } from "./Settings";

/**
 * カスタムAPIクライアントを使用するテスト用ラッパーを作成
 */
const createWrapper = (apiClient: IApiClient) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ChakraProvider theme={theme}>
          <ApiClientContext.Provider value={apiClient}>
            <BrowserRouter
              future={{
                v7_startTransition: true,
                v7_relativeSplatPath: true,
              }}
            >
              {children}
            </BrowserRouter>
          </ApiClientContext.Provider>
        </ChakraProvider>
      </QueryClientProvider>
    );
  };
};

describe("SettingsPage", () => {
  // モックAPIクライアント
  const mockApiClient: IApiClient = {
    auth: {
      register: vi.fn(),
      login: vi.fn(),
      me: vi.fn(),
      refresh: vi.fn(),
    },
    apps: {
      getAll: vi.fn().mockResolvedValue({
        apps: [
          {
            id: 1,
            name: "Test App",
            description: "Test description",
            table_name: "app_data_1",
            icon: "default",
            is_external: false,
            created_by: 1,
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
            fields: [],
            field_count: 1,
          },
        ],
        pagination: { page: 1, limit: 100, total: 1, total_pages: 1 },
      }),
      getById: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
    fields: {
      getByAppId: vi.fn().mockResolvedValue({
        fields: [
          {
            id: 1,
            app_id: 1,
            field_code: "title",
            field_name: "Title",
            field_type: "text",
            required: true,
            display_order: 1,
            options: {},
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        ],
      }),
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
      getAll: vi.fn().mockResolvedValue({
        users: [
          {
            id: 1,
            email: "admin@example.com",
            name: "Admin User",
            role: "admin",
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        ],
        pagination: { page: 1, limit: 20, total: 1, total_pages: 1 },
      }),
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

  // 各テスト前にモックをリセットし、管理者ユーザーを設定
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({
      user: {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        role: "admin",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
      token: "test-token",
      isAuthenticated: true,
      isAdmin: true,
    });
  });

  it("設定ページがタブ付きで表示される", async () => {
    render(<SettingsPage />, { wrapper: createWrapper(mockApiClient) });

    expect(screen.getByRole("heading", { name: "設定" })).toBeInTheDocument();
    expect(
      screen.getByRole("tab", { name: "プロフィール" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("tab", { name: "パスワード変更" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("tab", { name: "ユーザー管理" })
    ).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "アプリ設定" })).toBeInTheDocument();
  });

  it("デフォルトでプロフィール設定が表示される", async () => {
    render(<SettingsPage />, { wrapper: createWrapper(mockApiClient) });

    expect(screen.getByText("プロフィール設定")).toBeInTheDocument();
  });

  it("管理者ユーザーにユーザー管理タブが表示される", async () => {
    render(<SettingsPage />, { wrapper: createWrapper(mockApiClient) });

    expect(
      screen.getByRole("tab", { name: "ユーザー管理" })
    ).toBeInTheDocument();
  });

  it("一般ユーザーには管理者専用タブが非表示になる", async () => {
    // 一般ユーザーを設定
    useAuthStore.setState({
      user: {
        id: 2,
        email: "user@example.com",
        name: "Regular User",
        role: "user",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
      token: "test-token",
      isAuthenticated: true,
      isAdmin: false,
    });

    render(<SettingsPage />, { wrapper: createWrapper(mockApiClient) });

    // ユーザー管理とアプリ設定タブは非管理者には非表示
    expect(
      screen.queryByRole("tab", { name: "ユーザー管理" })
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("tab", { name: "アプリ設定" })
    ).not.toBeInTheDocument();

    // プロフィールとパスワード変更タブのみ表示
    expect(
      screen.getByRole("tab", { name: "プロフィール" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("tab", { name: "パスワード変更" })
    ).toBeInTheDocument();
  });

  it("管理者がアプリ設定タブをクリックするとコンテンツが表示される", async () => {
    // 管理者ユーザーを設定
    useAuthStore.setState({
      user: {
        id: 1,
        email: "admin@example.com",
        name: "Admin User",
        role: "admin",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
      token: "test-token",
      isAuthenticated: true,
      isAdmin: true,
    });

    const user = userEvent.setup();
    render(<SettingsPage />, { wrapper: createWrapper(mockApiClient) });

    const appSettingsTab = screen.getByRole("tab", { name: "アプリ設定" });
    await user.click(appSettingsTab);

    await waitFor(() => {
      // アプリ設定内の見出しを確認
      expect(
        screen.getByRole("heading", { name: "アプリ設定" })
      ).toBeInTheDocument();
    });
  });

  it("管理者のアプリ設定でアプリ一覧が表示される", async () => {
    // 管理者ユーザーを設定
    useAuthStore.setState({
      user: {
        id: 1,
        email: "admin@example.com",
        name: "Admin User",
        role: "admin",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
      token: "test-token",
      isAuthenticated: true,
      isAdmin: true,
    });

    const user = userEvent.setup();
    render(<SettingsPage />, { wrapper: createWrapper(mockApiClient) });

    const appSettingsTab = screen.getByRole("tab", { name: "アプリ設定" });
    await user.click(appSettingsTab);

    await waitFor(() => {
      expect(screen.getByText("Test App")).toBeInTheDocument();
    });
  });
});
