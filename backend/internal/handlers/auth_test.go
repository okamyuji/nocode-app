package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

func TestAuthHandler_Register(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful registration", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		req := models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}
		resp := &models.AuthResponse{
			Token: "jwt-token",
			User: &models.UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Test User",
			},
		}

		mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Register(rr, httpReq)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result models.AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "jwt-token", result.Token)
		assert.Equal(t, "test@example.com", result.User.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		req := models.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		mockService.On("Register", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(nil, services.ErrEmailAlreadyExists)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Register(rr, httpReq)

		assert.Equal(t, http.StatusConflict, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil)
		rr := httptest.NewRecorder()

		handler.Register(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Register(rr, httpReq)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful login", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		resp := &models.AuthResponse{
			Token: "jwt-token",
			User: &models.UserResponse{
				ID:    1,
				Email: "test@example.com",
				Name:  "Test User",
			},
		}

		mockService.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Login(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.AuthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "jwt-token", result.Token)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockService.On("Login", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, services.ErrInvalidCredentials)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Login(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
		rr := httptest.NewRecorder()

		handler.Login(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful get current user", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		resp := &models.UserResponse{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockService.On("GetCurrentUser", mock.Anything, uint64(1)).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		// Add claims to context
		claims := &utils.JWTClaims{UserID: 1}
		ctx := context.WithValue(httpReq.Context(), middleware.UserContextKey, claims)
		httpReq = httpReq.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.Me(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.UserResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - no claims", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		rr := httptest.NewRecorder()

		handler.Me(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		mockService.On("GetCurrentUser", mock.Anything, uint64(999)).Return(nil, services.ErrUserNotFound)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		claims := &utils.JWTClaims{UserID: 999}
		ctx := context.WithValue(httpReq.Context(), middleware.UserContextKey, claims)
		httpReq = httpReq.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.Me(rr, httpReq)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/me", nil)
		rr := httptest.NewRecorder()

		handler.Me(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	validator := utils.NewValidator()

	t.Run("successful token refresh", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		claims := &utils.JWTClaims{UserID: 1}
		mockService.On("RefreshToken", claims).Return("new-jwt-token", nil)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		ctx := context.WithValue(httpReq.Context(), middleware.UserContextKey, claims)
		httpReq = httpReq.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.Refresh(rr, httpReq)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "new-jwt-token", result["token"])

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - no claims", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		rr := httptest.NewRecorder()

		handler.Refresh(rr, httpReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockService := new(mocks.MockAuthService)
		handler := handlers.NewAuthHandler(mockService, validator)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/refresh", nil)
		rr := httptest.NewRecorder()

		handler.Refresh(rr, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
