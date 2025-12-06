import type {
  ChangePasswordRequest,
  CreateUserRequest,
  UpdateProfileRequest,
  UpdateUserRequest,
  User,
  UserListResponse,
} from "@/types";
import client from "./client";

/**
 * ユーザー管理API（管理者専用）
 */
export const usersApi = {
  // ユーザー一覧を取得（管理者専用）
  getAll: async (page = 1, limit = 20): Promise<UserListResponse> => {
    const response = await client.get<UserListResponse>("/users", {
      params: { page, limit },
    });
    return response.data;
  },

  // IDでユーザーを取得（管理者専用）
  getById: async (id: number): Promise<User> => {
    const response = await client.get<User>(`/users/${id}`);
    return response.data;
  },

  // 新しいユーザーを作成（管理者専用）
  create: async (data: CreateUserRequest): Promise<User> => {
    const response = await client.post<User>("/users", data);
    return response.data;
  },

  // ユーザーを更新（管理者専用）
  update: async (id: number, data: UpdateUserRequest): Promise<User> => {
    const response = await client.put<User>(`/users/${id}`, data);
    return response.data;
  },

  // ユーザーを削除（管理者専用）
  delete: async (id: number): Promise<void> => {
    await client.delete(`/users/${id}`);
  },
};

/**
 * プロフィール管理API
 */
export const profileApi = {
  // 自分のプロフィールを更新
  updateProfile: async (data: UpdateProfileRequest): Promise<User> => {
    const response = await client.put<User>("/auth/profile", data);
    return response.data;
  },

  // 自分のパスワードを変更
  changePassword: async (
    data: ChangePasswordRequest
  ): Promise<{ message: string }> => {
    const response = await client.put<{ message: string }>(
      "/auth/password",
      data
    );
    return response.data;
  },
};
