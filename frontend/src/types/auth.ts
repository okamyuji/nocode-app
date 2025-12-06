/**
 * 認証関連の型定義
 */

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
 * 登録リクエスト
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

// プロフィール管理

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

// ユーザー管理（管理者専用）

/**
 * ユーザー作成リクエスト（管理者専用）
 */
export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role: "admin" | "user";
}

/**
 * ユーザー更新リクエスト（管理者専用）
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
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}
