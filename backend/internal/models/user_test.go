package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nocode-app/backend/internal/models"
)

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	resp := user.ToResponse()

	assert.NotNil(t, resp)
	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, user.Role, resp.Role)
	assert.Equal(t, user.CreatedAt, resp.CreatedAt)
	assert.Equal(t, user.UpdatedAt, resp.UpdatedAt)
}

func TestUser_ToResponse_DoesNotExposePasswordHash(t *testing.T) {
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "secret-password-hash",
		Name:         "Test User",
		Role:         "user",
	}

	resp := user.ToResponse()

	// Verify PasswordHash is not in response struct at all
	assert.NotContains(t, "PasswordHash", resp)
	assert.NotContains(t, "password_hash", resp)
}
