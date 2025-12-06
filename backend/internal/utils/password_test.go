package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/utils"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "short password",
			password: "abc",
			wantErr:  false,
		},
		{
			name:     "normal password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-that-should-still-work",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "P@$$w0rd!#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: "パスワード123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	// Hash a known password
	password := "test-password-123"
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrong-password",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "not-a-valid-hash",
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
		{
			name:     "both empty",
			password: "",
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.CheckPassword(tt.password, tt.hash)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHashPassword_DifferentResults(t *testing.T) {
	// Same password should produce different hashes due to salt
	password := "same-password"
	hash1, err := utils.HashPassword(password)
	require.NoError(t, err)

	hash2, err := utils.HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)

	// But both should validate correctly
	assert.True(t, utils.CheckPassword(password, hash1))
	assert.True(t, utils.CheckPassword(password, hash2))
}
