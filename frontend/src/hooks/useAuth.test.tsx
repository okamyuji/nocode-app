import { useAuthStore } from "@/stores";
import { act, renderHook, waitFor } from "@testing-library/react";
import { ReactNode } from "react";
import { BrowserRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useAuth } from "./useAuth";

// react-router-dom navigateのモック
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// authApiのモック
vi.mock("@/api", () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    me: vi.fn(),
  },
}));

import { authApi } from "@/api";

describe("useAuth", () => {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <BrowserRouter
      future={{
        v7_startTransition: true,
        v7_relativeSplatPath: true,
      }}
    >
      {children}
    </BrowserRouter>
  );

  beforeEach(() => {
    // ストアの状態をリセット
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    });
    localStorage.clear();
    vi.clearAllMocks();
  });

  describe("initial state", () => {
    it("should return initial auth state", () => {
      const { result } = renderHook(() => useAuth(), { wrapper });

      expect(result.current.user).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isAuthenticated).toBe(false);
    });
  });

  describe("login", () => {
    it("should login and navigate to home", async () => {
      const mockResponse = {
        user: {
          id: 1,
          email: "test@example.com",
          name: "Test User",
          role: "admin" as const,
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
        token: "jwt-token",
      };

      vi.mocked(authApi.login).mockResolvedValueOnce(mockResponse);

      const { result } = renderHook(() => useAuth(), { wrapper });

      await act(async () => {
        await result.current.login("test@example.com", "password123");
      });

      await waitFor(() => {
        expect(result.current.isAuthenticated).toBe(true);
      });

      expect(mockNavigate).toHaveBeenCalledWith("/");
    });
  });

  describe("register", () => {
    it("should register and navigate to home", async () => {
      const mockResponse = {
        user: {
          id: 1,
          email: "new@example.com",
          name: "New User",
          role: "user" as const,
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
        token: "jwt-token",
      };

      vi.mocked(authApi.register).mockResolvedValueOnce(mockResponse);

      const { result } = renderHook(() => useAuth(), { wrapper });

      await act(async () => {
        await result.current.register(
          "new@example.com",
          "password123",
          "New User"
        );
      });

      await waitFor(() => {
        expect(result.current.isAuthenticated).toBe(true);
      });

      expect(mockNavigate).toHaveBeenCalledWith("/");
    });
  });

  describe("logout", () => {
    it("should logout and navigate to login", async () => {
      // First set up authenticated state
      useAuthStore.setState({
        user: {
          id: 1,
          email: "test@example.com",
          name: "Test User",
          role: "admin",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
        token: "jwt-token",
        isAuthenticated: true,
      });

      const { result } = renderHook(() => useAuth(), { wrapper });

      act(() => {
        result.current.logout();
      });

      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.user).toBeNull();
      expect(mockNavigate).toHaveBeenCalledWith("/login");
    });
  });

  describe("auto fetch user", () => {
    it("should fetch user when token exists but user is null", async () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        role: "admin" as const,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      vi.mocked(authApi.me).mockResolvedValueOnce(mockUser);

      useAuthStore.setState({
        token: "jwt-token",
        user: null,
      });

      renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(authApi.me).toHaveBeenCalled();
      });
    });

    it("should not fetch user when both token and user exist", () => {
      useAuthStore.setState({
        token: "jwt-token",
        user: {
          id: 1,
          email: "test@example.com",
          name: "Test User",
          role: "admin",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
      });

      renderHook(() => useAuth(), { wrapper });

      expect(authApi.me).not.toHaveBeenCalled();
    });
  });
});
