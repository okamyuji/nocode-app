import { ApiClientContext } from "@/api/ApiClientContext";
import type { IApiClient } from "@/api/interfaces";
import { useAuthStore } from "@/stores";
import { server } from "@/test/mocks/server";
import { theme } from "@/theme";
import { ChakraProvider } from "@chakra-ui/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { BrowserRouter } from "react-router-dom";
import { vi } from "vitest";
import { DataManagementPage } from "./DataManagement";

// Mock useNavigate
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockApps = [
  {
    id: 1,
    name: "顧客管理",
    description: "顧客情報を管理するアプリ",
    table_name: "app_data_1",
    icon: "default",
    is_external: false,
    created_by: 1,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    fields: [
      {
        id: 1,
        app_id: 1,
        field_code: "name",
        field_name: "顧客名",
        field_type: "text",
        required: true,
        display_order: 1,
        options: {},
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
    ],
    field_count: 1,
  },
  {
    id: 2,
    name: "商品管理",
    description: "商品在庫を管理",
    table_name: "app_data_2",
    icon: "grid",
    is_external: false,
    created_by: 1,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    fields: [],
    field_count: 0,
  },
];

// Create a wrapper
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  // Create mock API client that won't be used (MSW handles requests)
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

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ChakraProvider theme={theme}>
          <ApiClientContext.Provider value={mockApiClient}>
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

describe("DataManagementPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Set admin user state
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
    // Set default handler for apps with multiple apps
    server.use(
      http.get("/api/v1/apps", () => {
        return HttpResponse.json({
          apps: mockApps,
          pagination: { page: 1, limit: 100, total: 2, total_pages: 1 },
        });
      })
    );
  });

  it("renders data management page with title", async () => {
    render(<DataManagementPage />, { wrapper: createWrapper() });

    expect(
      screen.getByRole("heading", { name: "データ管理" })
    ).toBeInTheDocument();
    expect(
      screen.getByText("アプリを選択してレコードを管理します")
    ).toBeInTheDocument();
  });

  it("displays app list after loading", async () => {
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客管理")).toBeInTheDocument();
      expect(screen.getByText("商品管理")).toBeInTheDocument();
    });
  });

  it("shows field count badges", async () => {
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("1 フィールド")).toBeInTheDocument();
      expect(screen.getByText("0 フィールド")).toBeInTheDocument();
    });
  });

  it("shows app description", async () => {
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客情報を管理するアプリ")).toBeInTheDocument();
      expect(screen.getByText("商品在庫を管理")).toBeInTheDocument();
    });
  });

  it("navigates to records page when app card is clicked", async () => {
    const user = userEvent.setup();
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客管理")).toBeInTheDocument();
    });

    // Click the heading which should trigger navigation
    await user.click(screen.getByText("顧客管理"));

    expect(mockNavigate).toHaveBeenCalledWith("/apps/1/records");
  });

  it("navigates to settings page when settings icon is clicked", async () => {
    const user = userEvent.setup();
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客管理")).toBeInTheDocument();
    });

    // Find and click the settings button
    const settingsButtons = screen.getAllByRole("button", {
      name: "アプリ設定",
    });
    await user.click(settingsButtons[0]);

    expect(mockNavigate).toHaveBeenCalledWith("/settings?tab=apps&appId=1");
  });

  it("filters apps by search query", async () => {
    const user = userEvent.setup();
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客管理")).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText("アプリを検索...");
    await user.type(searchInput, "顧客");

    expect(screen.getByText("顧客管理")).toBeInTheDocument();
    expect(screen.queryByText("商品管理")).not.toBeInTheDocument();
  });

  it("shows no results message when search finds nothing", async () => {
    const user = userEvent.setup();
    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客管理")).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText("アプリを検索...");
    await user.type(searchInput, "存在しないアプリ");

    expect(
      screen.getByText("検索条件に一致するアプリがありません")
    ).toBeInTheDocument();
  });

  it("shows empty state when no apps exist", async () => {
    // Override handler for this specific test
    server.use(
      http.get("/api/v1/apps", () => {
        return HttpResponse.json({
          apps: [],
          pagination: { page: 1, limit: 100, total: 0, total_pages: 0 },
        });
      })
    );

    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("アプリがありません")).toBeInTheDocument();
      expect(
        screen.getByText("「アプリ一覧」から新しいアプリを作成してください")
      ).toBeInTheDocument();
    });
  });

  it("hides settings button for non-admin users", async () => {
    // Set non-admin user state
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

    render(<DataManagementPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("顧客管理")).toBeInTheDocument();
    });

    // Settings button should not be visible for non-admin users
    expect(
      screen.queryByRole("button", { name: "アプリ設定" })
    ).not.toBeInTheDocument();
  });
});
