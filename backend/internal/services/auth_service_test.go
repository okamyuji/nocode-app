package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		// Setup expectations
		mockUserRepo.On("EmailExists", ctx, "test@example.com").Return(false, nil)
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = 1
		})
		mockJWT.On("GenerateToken", uint64(1), "test@example.com", "user").Return("test-token", nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		req := &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		resp, err := service.Register(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-token", resp.Token)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.Equal(t, "Test User", resp.User.Name)

		mockUserRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		mockUserRepo.On("EmailExists", ctx, "existing@example.com").Return(true, nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		req := &models.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		_, err := service.Register(ctx, req)
		assert.ErrorIs(t, err, services.ErrEmailAlreadyExists)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("repository error on email check", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		mockUserRepo.On("EmailExists", ctx, "test@example.com").Return(false, errors.New("db error"))

		service := services.NewAuthService(mockUserRepo, mockJWT)

		req := &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		_, err := service.Register(ctx, req)
		assert.Error(t, err)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		// Hash the password for comparison
		hashedPassword, _ := utils.HashPassword("password123")

		user := &models.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
			Name:         "Test User",
			Role:         "user",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
		mockJWT.On("GenerateToken", uint64(1), "test@example.com", "user").Return("test-token", nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		resp, err := service.Login(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-token", resp.Token)
		assert.Equal(t, "test@example.com", resp.User.Email)

		mockUserRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		mockUserRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		req := &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		_, err := service.Login(ctx, req)
		assert.ErrorIs(t, err, services.ErrInvalidCredentials)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		hashedPassword, _ := utils.HashPassword("correctpassword")

		user := &models.User{
			ID:           1,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
			Name:         "Test User",
			Role:         "user",
		}

		mockUserRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		_, err := service.Login(ctx, req)
		assert.ErrorIs(t, err, services.ErrInvalidCredentials)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get user", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
			Role:  "user",
		}

		mockUserRepo.On("GetByID", ctx, uint64(1)).Return(user, nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		resp, err := service.GetCurrentUser(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, "Test User", resp.Name)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		mockUserRepo.On("GetByID", ctx, uint64(999)).Return(nil, nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		_, err := service.GetCurrentUser(ctx, 999)
		assert.ErrorIs(t, err, services.ErrUserNotFound)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		claims := &utils.JWTClaims{
			UserID: 1,
			Email:  "test@example.com",
			Role:   "user",
		}

		mockJWT.On("RefreshToken", claims).Return("new-token", nil)

		service := services.NewAuthService(mockUserRepo, mockJWT)

		token, err := service.RefreshToken(claims)
		require.NoError(t, err)
		assert.Equal(t, "new-token", token)

		mockJWT.AssertExpectations(t)
	})

	t.Run("refresh error", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockJWT := new(mocks.MockJWTManager)

		claims := &utils.JWTClaims{
			UserID: 1,
			Email:  "test@example.com",
			Role:   "user",
		}

		mockJWT.On("RefreshToken", claims).Return("", errors.New("token error"))

		service := services.NewAuthService(mockUserRepo, mockJWT)

		_, err := service.RefreshToken(claims)
		assert.Error(t, err)

		mockJWT.AssertExpectations(t)
	})
}
