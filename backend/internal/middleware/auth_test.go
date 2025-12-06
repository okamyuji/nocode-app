package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/utils"
)

func TestNewAuthMiddleware(t *testing.T) {
	jwtManager := utils.NewJWTManager("test-secret", 24)
	m := middleware.NewAuthMiddleware(jwtManager)
	assert.NotNil(t, m)
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	jwtManager := utils.NewJWTManager("test-secret", 24)
	m := middleware.NewAuthMiddleware(jwtManager)

	// Generate a valid token
	validToken, err := jwtManager.GenerateToken(1, "test@example.com", "user")
	require.NoError(t, err)

	// Create a test handler that will be wrapped
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r.Context())
		if ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(claims.Email))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing header",
			authHeader:     "",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid format - no bearer",
			authHeader:     validToken,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid format - wrong prefix",
			authHeader:     "Basic " + validToken,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "bearer lowercase",
			authHeader:     "bearer " + validToken,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "bearer mixed case",
			authHeader:     "BeArEr " + validToken,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "only bearer word",
			authHeader:     "Bearer",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			handler := m.Authenticate(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("with valid claims", func(t *testing.T) {
		claims := &utils.JWTClaims{
			UserID: 1,
			Email:  "test@example.com",
			Role:   "user",
		}
		ctx := context.WithValue(context.Background(), middleware.UserContextKey, claims)

		got, ok := middleware.GetUserFromContext(ctx)
		assert.True(t, ok)
		require.NotNil(t, got)
		assert.Equal(t, uint64(1), got.UserID)
		assert.Equal(t, "test@example.com", got.Email)
		assert.Equal(t, "user", got.Role)
	})

	t.Run("without claims", func(t *testing.T) {
		ctx := context.Background()

		got, ok := middleware.GetUserFromContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("with wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserContextKey, "not-claims")

		got, ok := middleware.GetUserFromContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	jwtManager := utils.NewJWTManager("test-secret", 24)
	m := middleware.NewAuthMiddleware(jwtManager)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		userRole       string
		requiredRole   string
		hasClaims      bool
		wantStatusCode int
	}{
		{
			name:           "matching role",
			userRole:       "editor",
			requiredRole:   "editor",
			hasClaims:      true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "admin can access any role",
			userRole:       "admin",
			requiredRole:   "editor",
			hasClaims:      true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "insufficient permissions",
			userRole:       "user",
			requiredRole:   "editor",
			hasClaims:      true,
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "no claims",
			userRole:       "",
			requiredRole:   "editor",
			hasClaims:      false,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.hasClaims {
				claims := &utils.JWTClaims{
					UserID: 1,
					Email:  "test@example.com",
					Role:   tt.userRole,
				}
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler := m.RequireRole(tt.requiredRole)(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		userRole       string
		hasClaims      bool
		wantStatusCode int
	}{
		{
			name:           "admin user allowed",
			userRole:       "admin",
			hasClaims:      true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "regular user denied",
			userRole:       "user",
			hasClaims:      true,
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "no claims - unauthorized",
			userRole:       "",
			hasClaims:      false,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.hasClaims {
				claims := &utils.JWTClaims{
					UserID: 1,
					Email:  "test@example.com",
					Role:   tt.userRole,
				}
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler := middleware.RequireAdmin(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name      string
		userRole  string
		hasClaims bool
		want      bool
	}{
		{
			name:      "admin user",
			userRole:  "admin",
			hasClaims: true,
			want:      true,
		},
		{
			name:      "regular user",
			userRole:  "user",
			hasClaims: true,
			want:      false,
		},
		{
			name:      "no claims",
			userRole:  "",
			hasClaims: false,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.hasClaims {
				claims := &utils.JWTClaims{
					UserID: 1,
					Email:  "test@example.com",
					Role:   tt.userRole,
				}
				ctx = context.WithValue(context.Background(), middleware.UserContextKey, claims)
			} else {
				ctx = context.Background()
			}

			got := middleware.IsAdmin(ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetUserInContext(t *testing.T) {
	claims := &utils.JWTClaims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "admin",
	}

	ctx := middleware.SetUserInContext(context.Background(), claims)

	got, ok := middleware.GetUserFromContext(ctx)
	assert.True(t, ok)
	require.NotNil(t, got)
	assert.Equal(t, claims.UserID, got.UserID)
	assert.Equal(t, claims.Email, got.Email)
	assert.Equal(t, claims.Role, got.Role)
}
