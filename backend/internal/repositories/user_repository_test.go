package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/testhelpers"
)

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewUserRepository(db)

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &models.User{
				Email:        "test@example.com",
				PasswordHash: "hashedpassword",
				Name:         "Test User",
				Role:         "user",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			user: &models.User{
				Email:        "admin@example.com", // Existing admin user
				PasswordHash: "hashedpassword",
				Name:         "Duplicate User",
				Role:         "user",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.ID)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewUserRepository(db)

	// Create a test user
	user := &models.User{
		Email:        "getbyid@example.com",
		PasswordHash: "hashedpassword",
		Name:         "GetByID User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	tests := []struct {
		name      string
		id        uint64
		wantUser  bool
		wantEmail string
	}{
		{
			name:      "existing user",
			id:        user.ID,
			wantUser:  true,
			wantEmail: "getbyid@example.com",
		},
		{
			name:     "non-existing user",
			id:       99999,
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)
			require.NoError(t, err)
			if tt.wantUser {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantEmail, result.Email)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewUserRepository(db)

	// Create a test user
	user := &models.User{
		Email:        "getbyemail@example.com",
		PasswordHash: "hashedpassword",
		Name:         "GetByEmail User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	tests := []struct {
		name     string
		email    string
		wantUser bool
		wantName string
	}{
		{
			name:     "existing email",
			email:    "getbyemail@example.com",
			wantUser: true,
			wantName: "GetByEmail User",
		},
		{
			name:     "admin email",
			email:    "admin@example.com",
			wantUser: true,
			wantName: "Admin",
		},
		{
			name:     "non-existing email",
			email:    "nonexistent@example.com",
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByEmail(ctx, tt.email)
			require.NoError(t, err)
			if tt.wantUser {
				require.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewUserRepository(db)

	// Create a test user
	user := &models.User{
		Email:        "update@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Original Name",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	// Update the user
	user.Name = "Updated Name"
	user.Role = "admin"
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify the update
	updated, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "admin", updated.Role)
}

func TestUserRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewUserRepository(db)

	// Create a test user
	user := &models.User{
		Email:        "delete@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Delete User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	// Delete the user
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestUserRepository_EmailExists(t *testing.T) {
	ctx := context.Background()
	db, err := testhelpers.GetTestDB(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testhelpers.ResetDatabase(ctx))
	})

	repo := repositories.NewUserRepository(db)

	// Create a test user
	user := &models.User{
		Email:        "exists@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Exists User",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	tests := []struct {
		name   string
		email  string
		exists bool
	}{
		{
			name:   "existing email",
			email:  "exists@example.com",
			exists: true,
		},
		{
			name:   "admin email",
			email:  "admin@example.com",
			exists: true,
		},
		{
			name:   "non-existing email",
			email:  "nonexistent@example.com",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.EmailExists(ctx, tt.email)
			require.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
		})
	}
}
