package models

import (
	"time"

	"github.com/uptrace/bun"
)

// User システムユーザーを表す構造体
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           uint64    `bun:"id,pk,autoincrement" json:"id"`
	Email        string    `bun:"email,notnull,unique" json:"email"`
	PasswordHash string    `bun:"password_hash,notnull" json:"-"`
	Name         string    `bun:"name,notnull" json:"name"`
	Role         string    `bun:"role,notnull,default:'user'" json:"role"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// UserResponse ユーザーデータのレスポンス構造体（機密フィールドを除外）
type UserResponse struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse UserをUserResponseに変換する
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// RegisterRequest ユーザー登録リクエストの構造体
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
}

// LoginRequest ユーザーログインリクエストの構造体
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse 認証レスポンスの構造体
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// UpdateProfileRequest プロフィール更新リクエストの構造体
type UpdateProfileRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// ChangePasswordRequest パスワード変更リクエストの構造体
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// CreateUserRequest ユーザー作成リクエストの構造体（管理者専用）
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

// UpdateUserRequest ユーザー更新リクエストの構造体（管理者専用）
type UpdateUserRequest struct {
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Role string `json:"role,omitempty" validate:"omitempty,oneof=admin user"`
}

// UserListResponse ユーザー一覧のレスポンス構造体
type UserListResponse struct {
	Users      []*UserResponse `json:"users"`
	Pagination *Pagination     `json:"pagination"`
}
