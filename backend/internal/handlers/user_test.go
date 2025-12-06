package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

func TestUserHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name: "success",
			setupMock: func(m *mocks.MockUserService) {
				m.On("GetUsers", mock.Anything, "admin", 1, 20).Return(&models.UserListResponse{
					Users: []*models.UserResponse{
						{ID: 1, Email: "admin@example.com", Name: "Admin", Role: "admin"},
					},
					Pagination: &models.Pagination{Page: 1, Limit: 20, Total: 1, TotalPages: 1},
				}, nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "forbidden - not admin",
			setupMock: func(m *mocks.MockUserService) {
				m.On("GetUsers", mock.Anything, "user", 1, 20).Return(nil, services.ErrNotAdmin)
			},
			claims:         &utils.JWTClaims{UserID: 2, Email: "user@example.com", Role: "user"},
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&limit=20", nil)
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.List(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name:   "success",
			userID: "1",
			setupMock: func(m *mocks.MockUserService) {
				m.On("GetUser", mock.Anything, "admin", uint64(1)).Return(&models.UserResponse{
					ID: 1, Email: "user@example.com", Name: "User", Role: "user",
					CreatedAt: time.Now(), UpdatedAt: time.Now(),
				}, nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "not found",
			userID: "999",
			setupMock: func(m *mocks.MockUserService) {
				m.On("GetUser", mock.Anything, "admin", uint64(999)).Return(nil, services.ErrUserNotFound)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.Get(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name: "success",
			body: models.CreateUserRequest{
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
				Role:     "user",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("CreateUser", mock.Anything, "admin", mock.AnythingOfType("*models.CreateUserRequest")).
					Return(&models.UserResponse{
						ID: 3, Email: "new@example.com", Name: "New User", Role: "user",
						CreatedAt: time.Now(), UpdatedAt: time.Now(),
					}, nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "conflict - email exists",
			body: models.CreateUserRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Name:     "New User",
				Role:     "user",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("CreateUser", mock.Anything, "admin", mock.AnythingOfType("*models.CreateUserRequest")).
					Return(nil, services.ErrEmailAlreadyExists)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.Create(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		body           interface{}
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name:   "success",
			userID: "2",
			body:   models.UpdateUserRequest{Name: "Updated Name"},
			setupMock: func(m *mocks.MockUserService) {
				m.On("UpdateUser", mock.Anything, uint64(1), "admin", uint64(2), mock.AnythingOfType("*models.UpdateUserRequest")).
					Return(&models.UserResponse{
						ID: 2, Email: "user@example.com", Name: "Updated Name", Role: "user",
						CreatedAt: time.Now(), UpdatedAt: time.Now(),
					}, nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "bad request - cannot change own role",
			userID: "1",
			body:   models.UpdateUserRequest{Role: "user"},
			setupMock: func(m *mocks.MockUserService) {
				m.On("UpdateUser", mock.Anything, uint64(1), "admin", uint64(1), mock.AnythingOfType("*models.UpdateUserRequest")).
					Return(nil, services.ErrCannotChangeSelfRole)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tt.userID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.Update(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name:   "success",
			userID: "2",
			setupMock: func(m *mocks.MockUserService) {
				m.On("DeleteUser", mock.Anything, uint64(1), "admin", uint64(2)).Return(nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "bad request - cannot delete self",
			userID: "1",
			setupMock: func(m *mocks.MockUserService) {
				m.On("DeleteUser", mock.Anything, uint64(1), "admin", uint64(1)).Return(services.ErrCannotDeleteSelf)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "admin@example.com", Role: "admin"},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+tt.userID, nil)
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.Delete(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name: "success",
			body: models.UpdateProfileRequest{Name: "New Name"},
			setupMock: func(m *mocks.MockUserService) {
				m.On("UpdateProfile", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateProfileRequest")).
					Return(&models.UserResponse{
						ID: 1, Email: "user@example.com", Name: "New Name", Role: "user",
						CreatedAt: time.Now(), UpdatedAt: time.Now(),
					}, nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "user@example.com", Role: "user"},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			body: models.UpdateProfileRequest{Name: "New Name"},
			setupMock: func(m *mocks.MockUserService) {
				m.On("UpdateProfile", mock.Anything, uint64(1), mock.AnythingOfType("*models.UpdateProfileRequest")).
					Return(nil, services.ErrUserNotFound)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "user@example.com", Role: "user"},
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/profile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.UpdateProfile(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ChangePassword(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(*mocks.MockUserService)
		claims         *utils.JWTClaims
		wantStatusCode int
	}{
		{
			name: "success",
			body: models.ChangePasswordRequest{
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("ChangePassword", mock.Anything, uint64(1), mock.AnythingOfType("*models.ChangePasswordRequest")).
					Return(nil)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "user@example.com", Role: "user"},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "bad request - invalid password",
			body: models.ChangePasswordRequest{
				CurrentPassword: "wrongpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.On("ChangePassword", mock.Anything, uint64(1), mock.AnythingOfType("*models.ChangePasswordRequest")).
					Return(services.ErrInvalidPassword)
			},
			claims:         &utils.JWTClaims{UserID: 1, Email: "user@example.com", Role: "user"},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockUserService)
			tt.setupMock(mockService)

			handler := handlers.NewUserHandler(mockService, utils.NewValidator())

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserInContext(req.Context(), tt.claims)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ChangePassword(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// Helper to set user in context
func init() {
	// This is just to ensure the context setting works
}
