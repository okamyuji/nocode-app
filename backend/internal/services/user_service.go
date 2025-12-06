package services

import (
	"context"
	"errors"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// ユーザー管理エラー
var (
	ErrNotAdmin             = errors.New("管理者権限が必要です")
	ErrCannotDeleteSelf     = errors.New("自分のアカウントは削除できません")
	ErrInvalidPassword      = errors.New("現在のパスワードが正しくありません")
	ErrCannotChangeSelfRole = errors.New("自分のロールは変更できません")
)

// UserService ユーザー管理操作を処理する構造体
type UserService struct {
	userRepo       repositories.UserRepositoryInterface
	passwordHasher PasswordHasher
}

// NewUserService 新しいUserServiceを作成する
func NewUserService(userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo:       userRepo,
		passwordHasher: NewDefaultPasswordHasher(),
	}
}

// NewUserServiceWithHasher カスタムパスワードハッシャー付きで新しいUserServiceを作成する（テスト用）
func NewUserServiceWithHasher(userRepo repositories.UserRepositoryInterface, hasher PasswordHasher) *UserService {
	return &UserService{
		userRepo:       userRepo,
		passwordHasher: hasher,
	}
}

// GetUsers ページネーション付きで全ユーザーを取得する（管理者専用）
func (s *UserService) GetUsers(ctx context.Context, callerRole string, page, limit int) (*models.UserListResponse, error) {
	if callerRole != "admin" {
		return nil, ErrNotAdmin
	}

	users, total, err := s.userRepo.GetAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	userResponses := make([]*models.UserResponse, len(users))
	for i := range users {
		userResponses[i] = users[i].ToResponse()
	}

	return &models.UserListResponse{
		Users:      userResponses,
		Pagination: models.NewPagination(page, limit, total),
	}, nil
}

// GetUser IDでユーザーを取得する（管理者専用）
func (s *UserService) GetUser(ctx context.Context, callerRole string, userID uint64) (*models.UserResponse, error) {
	if callerRole != "admin" {
		return nil, ErrNotAdmin
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user.ToResponse(), nil
}

// CreateUser 新しいユーザーを作成する（管理者専用）
func (s *UserService) CreateUser(ctx context.Context, callerRole string, req *models.CreateUserRequest) (*models.UserResponse, error) {
	if callerRole != "admin" {
		return nil, ErrNotAdmin
	}

	// メールアドレスの存在確認
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// パスワードをハッシュ化
	hashedPassword, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// ユーザーを作成
	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Name:         req.Name,
		Role:         req.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// UpdateUser ユーザーを更新する（管理者専用）
func (s *UserService) UpdateUser(ctx context.Context, callerID uint64, callerRole string, userID uint64, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	if callerRole != "admin" {
		return nil, ErrNotAdmin
	}

	// 既存ユーザーを取得
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 自分のロール変更を防止
	if callerID == userID && req.Role != "" && req.Role != user.Role {
		return nil, ErrCannotChangeSelfRole
	}

	// フィールドを更新
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// DeleteUser ユーザーを削除する（管理者専用）
func (s *UserService) DeleteUser(ctx context.Context, callerID uint64, callerRole string, userID uint64) error {
	if callerRole != "admin" {
		return ErrNotAdmin
	}

	// 自分の削除を防止
	if callerID == userID {
		return ErrCannotDeleteSelf
	}

	// ユーザーの存在確認
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, userID)
}

// UpdateProfile 現在のユーザーのプロフィールを更新する
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, req *models.UpdateProfileRequest) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.Name = req.Name
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// ChangePassword 現在のユーザーのパスワードを変更する
func (s *UserService) ChangePassword(ctx context.Context, userID uint64, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 現在のパスワードを検証
	if !s.passwordHasher.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}

	// 新しいパスワードをハッシュ化
	hashedPassword, err := s.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}
