package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWTトークン内のクレームを表す構造体
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManagerInterface JWT操作のインターフェースを定義
type JWTManagerInterface interface {
	GenerateToken(userID uint64, email, role string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(claims *JWTClaims) (string, error)
}

// JWTManager JWT操作を処理する構造体
type JWTManager struct {
	secret      []byte
	expiryHours int
}

// JWTManagerがJWTManagerInterfaceを実装していることを確認
var _ JWTManagerInterface = (*JWTManager)(nil)

// NewJWTManager 新しいJWTManagerを作成する
func NewJWTManager(secret string, expiryHours int) *JWTManager {
	return &JWTManager{
		secret:      []byte(secret),
		expiryHours: expiryHours,
	}
}

// GenerateToken ユーザー用の新しいJWTトークンを生成する
func (m *JWTManager) GenerateToken(userID uint64, email, role string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "nocode-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken JWTトークンを検証しクレームを返す
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("予期しない署名方式です")
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("無効なトークンです")
	}

	return claims, nil
}

// RefreshToken 有効期限を延長した新しいトークンを生成する
func (m *JWTManager) RefreshToken(claims *JWTClaims) (string, error) {
	return m.GenerateToken(claims.UserID, claims.Email, claims.Role)
}
