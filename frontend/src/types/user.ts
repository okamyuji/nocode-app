/**
 * ユーザー関連の型定義
 */

import { Pagination } from "./app";

/**
 * ユーザー
 */
export interface User {
  id: number;
  email: string;
  name: string;
  role: "admin" | "user";
  created_at: string;
  updated_at: string;
}

/**
 * ユーザー登録リクエスト
 */
export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

/**
 * ログインリクエスト
 */
export interface LoginRequest {
  email: string;
  password: string;
}

/**
 * 認証レスポンス
 */
export interface AuthResponse {
  token: string;
  user: User;
}

/**
 * プロフィール更新リクエスト
 */
export interface UpdateProfileRequest {
  name: string;
}

/**
 * パスワード変更リクエスト
 */
export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

/**
 * ユーザー作成リクエスト（管理者用）
 */
export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role: "admin" | "user";
}

/**
 * ユーザー更新リクエスト（管理者用）
 */
export interface UpdateUserRequest {
  name?: string;
  role?: "admin" | "user";
}

/**
 * ユーザー一覧レスポンス
 */
export interface UserListResponse {
  users: User[];
  pagination: Pagination;
}
