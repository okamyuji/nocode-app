/**
 * 認証保護ルートコンポーネントのテスト
 */

import { useAuthStore } from "@/stores";
import { theme } from "@/theme";
import { ChakraProvider } from "@chakra-ui/react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ProtectedRoute } from "./ProtectedRoute";

// 認証ストアのモック
vi.mock("@/stores", () => ({
  useAuthStore: vi.fn(),
}));

/** テスト用ラッパーでルーターを初期化 */
const renderWithRouter = (initialRoute = "/protected") => {
  return render(
    <ChakraProvider theme={theme}>
      <MemoryRouter
        initialEntries={[initialRoute]}
        future={{
          v7_startTransition: true,
          v7_relativeSplatPath: true,
        }}
      >
        <Routes>
          <Route path="/login" element={<div>Login Page</div>} />
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    </ChakraProvider>
  );
};

describe("ProtectedRoute", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("読み込み中の場合はローディング表示される", () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: false,
      isLoading: true,
      token: null,
    } as ReturnType<typeof useAuthStore>);

    renderWithRouter();

    expect(screen.getByText("認証情報を確認中...")).toBeInTheDocument();
  });

  it("認証済みの場合は子コンポーネントが表示される", () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
      token: "jwt-token",
    } as ReturnType<typeof useAuthStore>);

    renderWithRouter();

    expect(screen.getByText("Protected Content")).toBeInTheDocument();
  });

  it("未認証の場合はログインページにリダイレクトされる", () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      token: null,
    } as ReturnType<typeof useAuthStore>);

    renderWithRouter();

    expect(screen.getByText("Login Page")).toBeInTheDocument();
    expect(screen.queryByText("Protected Content")).not.toBeInTheDocument();
  });

  it("トークンが存在する場合は完全に認証されていなくてもアクセスを許可する", () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      token: "jwt-token",
    } as ReturnType<typeof useAuthStore>);

    renderWithRouter();

    expect(screen.getByText("Protected Content")).toBeInTheDocument();
  });
});
