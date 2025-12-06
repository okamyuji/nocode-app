package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/utils"
)

func TestNewJWTManager(t *testing.T) {
	manager := utils.NewJWTManager("test-secret", 24)
	assert.NotNil(t, manager)
}

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := utils.NewJWTManager("test-secret", 24)

	tests := []struct {
		name    string
		userID  uint64
		email   string
		role    string
		wantErr bool
	}{
		{
			name:    "valid token generation",
			userID:  1,
			email:   "test@example.com",
			role:    "user",
			wantErr: false,
		},
		{
			name:    "admin user token",
			userID:  2,
			email:   "admin@example.com",
			role:    "admin",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tt.userID, tt.email, tt.role)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := utils.NewJWTManager("test-secret", 24)

	// Generate a valid token first
	validToken, err := manager.GenerateToken(1, "test@example.com", "user")
	require.NoError(t, err)

	// Create manager with different secret for invalid token test
	otherManager := utils.NewJWTManager("other-secret", 24)
	tokenWithDifferentSecret, err := otherManager.GenerateToken(2, "other@example.com", "user")
	require.NoError(t, err)

	tests := []struct {
		name       string
		token      string
		wantErr    bool
		wantUserID uint64
		wantEmail  string
		wantRole   string
	}{
		{
			name:       "valid token",
			token:      validToken,
			wantErr:    false,
			wantUserID: 1,
			wantEmail:  "test@example.com",
			wantRole:   "user",
		},
		{
			name:    "invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "token signed with different secret",
			token:   tokenWithDifferentSecret,
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not-a-jwt-token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, claims)
				assert.Equal(t, tt.wantUserID, claims.UserID)
				assert.Equal(t, tt.wantEmail, claims.Email)
				assert.Equal(t, tt.wantRole, claims.Role)
			}
		})
	}
}

func TestJWTManager_RefreshToken(t *testing.T) {
	manager := utils.NewJWTManager("test-secret", 24)

	// Generate original token
	originalToken, err := manager.GenerateToken(1, "test@example.com", "user")
	require.NoError(t, err)

	// Validate and get claims
	claims, err := manager.ValidateToken(originalToken)
	require.NoError(t, err)

	// Wait to ensure different timestamp (JWT uses second precision)
	time.Sleep(1100 * time.Millisecond)

	// Refresh token
	newToken, err := manager.RefreshToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)

	// Validate new token
	newClaims, err := manager.ValidateToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, newClaims.UserID)
	assert.Equal(t, claims.Email, newClaims.Email)
	assert.Equal(t, claims.Role, newClaims.Role)
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	// Create manager with 0 hour expiry (token will be expired immediately)
	manager := utils.NewJWTManager("test-secret", 0)

	token, err := manager.GenerateToken(1, "test@example.com", "user")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(100 * time.Millisecond)

	// Validate should fail for expired token
	claims, err := manager.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
