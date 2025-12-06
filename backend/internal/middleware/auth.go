package middleware

import (
	"context"
	"net/http"
	"strings"

	"nocode-app/backend/internal/utils"
)

// ContextKey コンテキストキーの型
type ContextKey string

const (
	// UserContextKey コンテキスト内のユーザークレームのキー
	UserContextKey ContextKey = "user"
)

// AuthMiddleware JWT認証ミドルウェア
type AuthMiddleware struct {
	jwtManager utils.JWTManagerInterface
}

// NewAuthMiddleware 新しいAuthMiddlewareを作成する
func NewAuthMiddleware(jwtManager utils.JWTManagerInterface) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// Authenticate JWT認証でハンドラーをラップする
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// クレームをコンテキストに追加
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext コンテキストからユーザークレームを取得する
func GetUserFromContext(ctx context.Context) (*utils.JWTClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*utils.JWTClaims)
	return claims, ok
}

// SetUserInContext ユーザークレームをコンテキストに設定する（テスト用）
func SetUserInContext(ctx context.Context, claims *utils.JWTClaims) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}

// RequireRole 特定のロールを要求するミドルウェアを作成する
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if claims.Role != role && claims.Role != "admin" {
				utils.WriteErrorResponse(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin 管理者ロールを要求するミドルウェア
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserFromContext(r.Context())
		if !ok {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
			return
		}

		if claims.Role != "admin" {
			utils.WriteErrorResponse(w, http.StatusForbidden, "管理者権限が必要です")
			return
		}

		next(w, r)
	}
}

// IsAdmin コンテキスト内のユーザーが管理者かどうかを確認する
func IsAdmin(ctx context.Context) bool {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}
	return claims.Role == "admin"
}
