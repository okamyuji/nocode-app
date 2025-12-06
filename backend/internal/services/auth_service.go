package services

import (
	"context"
	"errors"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/utils"
)

// 共通エラー
var (
	ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが無効です")
	ErrEmailAlreadyExists = errors.New("メールアドレスは既に存在します")
	ErrUserNotFound       = errors.New("ユーザーが見つかりません")
)

// PasswordHasher パスワード操作のインターフェースを定義
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

// defaultPasswordHasher bcryptを使用してPasswordHasherを実装
type defaultPasswordHasher struct{}

// HashPassword パスワードをハッシュ化する
func (h *defaultPasswordHasher) HashPassword(password string) (string, error) {
	return utils.HashPassword(password)
}

// CheckPassword パスワードを検証する
func (h *defaultPasswordHasher) CheckPassword(password, hash string) bool {
	return utils.CheckPassword(password, hash)
}

// NewDefaultPasswordHasher 新しいデフォルトパスワードハッシャーを作成する
func NewDefaultPasswordHasher() PasswordHasher {
	return &defaultPasswordHasher{}
}

// AuthService 認証操作を処理する構造体
type AuthService struct {
	userRepo       repositories.UserRepositoryInterface
	jwtManager     utils.JWTManagerInterface
	passwordHasher PasswordHasher
}

// NewAuthService 新しいAuthServiceを作成する
func NewAuthService(userRepo repositories.UserRepositoryInterface, jwtManager utils.JWTManagerInterface) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		passwordHasher: NewDefaultPasswordHasher(),
	}
}

// NewAuthServiceWithHasher カスタムパスワードハッシャー付きで新しいAuthServiceを作成する（テスト用）
func NewAuthServiceWithHasher(userRepo repositories.UserRepositoryInterface, jwtManager utils.JWTManagerInterface, hasher PasswordHasher) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		passwordHasher: hasher,
	}
}

// Register 新しいユーザーを登録する
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
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
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// トークンを生成
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login ユーザーを認証する
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// メールアドレスでユーザーを取得
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// パスワードを確認
	if !s.passwordHasher.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// トークンを生成
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetCurrentUser クレームから現在のユーザーを取得する
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uint64) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user.ToResponse(), nil
}

// RefreshToken JWTトークンを更新する
func (s *AuthService) RefreshToken(claims *utils.JWTClaims) (string, error) {
	return s.jwtManager.RefreshToken(claims)
}
