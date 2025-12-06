import { AuthResponse, LoginRequest, RegisterRequest, User } from "@/types";
import client from "./client";

/**
 * 認証API
 */
export const authApi = {
  // ユーザー登録
  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await client.post<AuthResponse>("/auth/register", data);
    return response.data;
  },

  // ログイン
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await client.post<AuthResponse>("/auth/login", data);
    return response.data;
  },

  // 現在のユーザー情報を取得
  me: async (): Promise<User> => {
    const response = await client.get<User>("/auth/me");
    return response.data;
  },

  // トークンを更新
  refresh: async (): Promise<{ token: string }> => {
    const response = await client.post<{ token: string }>("/auth/refresh");
    return response.data;
  },
};
