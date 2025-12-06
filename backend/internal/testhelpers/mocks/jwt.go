// Package mocks テスト用のモック実装を提供
package mocks

import (
	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/utils"
)

// MockJWTManager JWTManagerInterfaceのモック実装
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID uint64, email, role string) (string, error) {
	args := m.Called(userID, email, role)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.JWTClaims), args.Error(1)
}

func (m *MockJWTManager) RefreshToken(claims *utils.JWTClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}
