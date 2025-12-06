import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useAuthStore } from "./authStore";

// authApiをモック
vi.mock("@/api", () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    me: vi.fn(),
  },
}));

import { authApi } from "@/api";

describe("authStore", () => {
  beforeEach(() => {
    // 各テスト前にストアの状態をリセット
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    });
    localStorage.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("setUser", () => {
    it("should set user and update isAuthenticated", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        role: "admin" as const,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      useAuthStore.getState().setUser(mockUser);

      expect(useAuthStore.getState().user).toEqual(mockUser);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
    });

    it("should clear user and set isAuthenticated to false when null", () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        role: "admin" as const,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      useAuthStore.getState().setUser(mockUser);
      useAuthStore.getState().setUser(null);

      expect(useAuthStore.getState().user).toBeNull();
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });

  describe("setToken", () => {
    it("should set token and save to localStorage", () => {
      useAuthStore.getState().setToken("test-token");

      expect(useAuthStore.getState().token).toBe("test-token");
      expect(localStorage.getItem("token")).toBe("test-token");
    });

    it("should remove token from localStorage when null", () => {
      localStorage.setItem("token", "test-token");
      useAuthStore.getState().setToken(null);

      expect(useAuthStore.getState().token).toBeNull();
      expect(localStorage.getItem("token")).toBeNull();
    });
  });

  describe("login", () => {
    it("should login successfully", async () => {
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

      await useAuthStore.getState().login("test@example.com", "password123");

      expect(authApi.login).toHaveBeenCalledWith({
        email: "test@example.com",
        password: "password123",
      });
      expect(useAuthStore.getState().user).toEqual(mockResponse.user);
      expect(useAuthStore.getState().token).toBe("jwt-token");
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().isLoading).toBe(false);
    });

    it("should handle login error", async () => {
      vi.mocked(authApi.login).mockRejectedValueOnce(
        new Error("Invalid credentials")
      );

      await expect(
        useAuthStore.getState().login("test@example.com", "wrongpassword")
      ).rejects.toThrow("Invalid credentials");

      expect(useAuthStore.getState().isLoading).toBe(false);
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });

  describe("register", () => {
    it("should register successfully", async () => {
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

      await useAuthStore
        .getState()
        .register("new@example.com", "password123", "New User");

      expect(authApi.register).toHaveBeenCalledWith({
        email: "new@example.com",
        password: "password123",
        name: "New User",
      });
      expect(useAuthStore.getState().user).toEqual(mockResponse.user);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
    });
  });

  describe("logout", () => {
    it("should clear user, token and reset isAuthenticated", () => {
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
      localStorage.setItem("token", "jwt-token");

      useAuthStore.getState().logout();

      expect(useAuthStore.getState().user).toBeNull();
      expect(useAuthStore.getState().token).toBeNull();
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(localStorage.getItem("token")).toBeNull();
    });
  });

  describe("fetchUser", () => {
    it("should fetch user when token exists", async () => {
      const mockUser = {
        id: 1,
        email: "test@example.com",
        name: "Test User",
        role: "admin" as const,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      useAuthStore.setState({ token: "jwt-token" });
      vi.mocked(authApi.me).mockResolvedValueOnce(mockUser);

      await useAuthStore.getState().fetchUser();

      expect(authApi.me).toHaveBeenCalled();
      expect(useAuthStore.getState().user).toEqual(mockUser);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
    });

    it("should not fetch user when no token exists", async () => {
      await useAuthStore.getState().fetchUser();

      expect(authApi.me).not.toHaveBeenCalled();
    });

    it("should logout on fetch error", async () => {
      useAuthStore.setState({ token: "jwt-token" });
      vi.mocked(authApi.me).mockRejectedValueOnce(new Error("Unauthorized"));

      await useAuthStore.getState().fetchUser();

      expect(useAuthStore.getState().user).toBeNull();
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
    });
  });
});
