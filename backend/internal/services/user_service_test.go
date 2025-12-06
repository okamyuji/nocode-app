package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
)

// テスト用MockPasswordHasher
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) CheckPassword(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

func TestUserService_GetUsers(t *testing.T) {
	tests := []struct {
		name       string
		callerRole string
		page       int
		limit      int
		setupMock  func(*mocks.MockUserRepository)
		wantErr    error
		wantCount  int
	}{
		{
			name:       "success - admin gets users",
			callerRole: "admin",
			page:       1,
			limit:      20,
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetAll", mock.Anything, 1, 20).Return([]models.User{
					{ID: 1, Email: "admin@example.com", Name: "Admin", Role: "admin"},
					{ID: 2, Email: "user@example.com", Name: "User", Role: "user"},
				}, int64(2), nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:       "error - non-admin denied",
			callerRole: "user",
			page:       1,
			limit:      20,
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantErr:    services.ErrNotAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo)
			result, err := svc.GetUsers(context.Background(), tt.callerRole, tt.page, tt.limit)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Users, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		callerRole string
		userID     uint64
		setupMock  func(*mocks.MockUserRepository)
		wantErr    error
	}{
		{
			name:       "success",
			callerRole: "admin",
			userID:     1,
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(1)).Return(&models.User{
					ID: 1, Email: "user@example.com", Name: "User", Role: "user",
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:       "error - not admin",
			callerRole: "user",
			userID:     1,
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantErr:    services.ErrNotAdmin,
		},
		{
			name:       "error - user not found",
			callerRole: "admin",
			userID:     999,
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			wantErr: services.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo)
			result, err := svc.GetUser(context.Background(), tt.callerRole, tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		callerRole string
		req        *models.CreateUserRequest
		setupMock  func(*mocks.MockUserRepository, *MockPasswordHasher)
		wantErr    error
	}{
		{
			name:       "success",
			callerRole: "admin",
			req: &models.CreateUserRequest{
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
				Role:     "user",
			},
			setupMock: func(m *mocks.MockUserRepository, h *MockPasswordHasher) {
				m.On("EmailExists", mock.Anything, "new@example.com").Return(false, nil)
				h.On("HashPassword", "password123").Return("hashed", nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "error - not admin",
			callerRole: "user",
			req: &models.CreateUserRequest{
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
				Role:     "user",
			},
			setupMock: func(_ *mocks.MockUserRepository, _ *MockPasswordHasher) {},
			wantErr:   services.ErrNotAdmin,
		},
		{
			name:       "error - email exists",
			callerRole: "admin",
			req: &models.CreateUserRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Name:     "New User",
				Role:     "user",
			},
			setupMock: func(m *mocks.MockUserRepository, _ *MockPasswordHasher) {
				m.On("EmailExists", mock.Anything, "existing@example.com").Return(true, nil)
			},
			wantErr: services.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockHasher := new(MockPasswordHasher)
			tt.setupMock(mockRepo, mockHasher)

			svc := services.NewUserServiceWithHasher(mockRepo, mockHasher)
			result, err := svc.CreateUser(context.Background(), tt.callerRole, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	tests := []struct {
		name       string
		callerID   uint64
		callerRole string
		userID     uint64
		req        *models.UpdateUserRequest
		setupMock  func(*mocks.MockUserRepository)
		wantErr    error
	}{
		{
			name:       "success - update name",
			callerID:   1,
			callerRole: "admin",
			userID:     2,
			req:        &models.UpdateUserRequest{Name: "Updated Name"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(2)).Return(&models.User{
					ID: 2, Email: "user@example.com", Name: "Original", Role: "user",
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "error - cannot change own role",
			callerID:   1,
			callerRole: "admin",
			userID:     1,
			req:        &models.UpdateUserRequest{Role: "user"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(1)).Return(&models.User{
					ID: 1, Email: "admin@example.com", Name: "Admin", Role: "admin",
				}, nil)
			},
			wantErr: services.ErrCannotChangeSelfRole,
		},
		{
			name:       "error - not admin",
			callerID:   2,
			callerRole: "user",
			userID:     3,
			req:        &models.UpdateUserRequest{Name: "New Name"},
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantErr:    services.ErrNotAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo)
			result, err := svc.UpdateUser(context.Background(), tt.callerID, tt.callerRole, tt.userID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		callerID   uint64
		callerRole string
		userID     uint64
		setupMock  func(*mocks.MockUserRepository)
		wantErr    error
	}{
		{
			name:       "success",
			callerID:   1,
			callerRole: "admin",
			userID:     2,
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(2)).Return(&models.User{
					ID: 2, Email: "user@example.com", Name: "User", Role: "user",
				}, nil)
				m.On("Delete", mock.Anything, uint64(2)).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:       "error - cannot delete self",
			callerID:   1,
			callerRole: "admin",
			userID:     1,
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantErr:    services.ErrCannotDeleteSelf,
		},
		{
			name:       "error - not admin",
			callerID:   2,
			callerRole: "user",
			userID:     3,
			setupMock:  func(_ *mocks.MockUserRepository) {},
			wantErr:    services.ErrNotAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo)
			err := svc.DeleteUser(context.Background(), tt.callerID, tt.callerRole, tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint64
		req       *models.UpdateProfileRequest
		setupMock func(*mocks.MockUserRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			req:    &models.UpdateProfileRequest{Name: "New Name"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(1)).Return(&models.User{
					ID: 1, Email: "user@example.com", Name: "Old Name", Role: "user",
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "error - user not found",
			userID: 999,
			req:    &models.UpdateProfileRequest{Name: "New Name"},
			setupMock: func(m *mocks.MockUserRepository) {
				m.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			wantErr: services.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			tt.setupMock(mockRepo)

			svc := services.NewUserService(mockRepo)
			result, err := svc.UpdateProfile(context.Background(), tt.userID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "New Name", result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint64
		req       *models.ChangePasswordRequest
		setupMock func(*mocks.MockUserRepository, *MockPasswordHasher)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			req: &models.ChangePasswordRequest{
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(m *mocks.MockUserRepository, h *MockPasswordHasher) {
				m.On("GetByID", mock.Anything, uint64(1)).Return(&models.User{
					ID: 1, Email: "user@example.com", Name: "User", Role: "user",
					PasswordHash: "oldhash", CreatedAt: time.Now(), UpdatedAt: time.Now(),
				}, nil)
				h.On("CheckPassword", "oldpassword", "oldhash").Return(true)
				h.On("HashPassword", "newpassword").Return("newhash", nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "error - invalid current password",
			userID: 1,
			req: &models.ChangePasswordRequest{
				CurrentPassword: "wrongpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(m *mocks.MockUserRepository, h *MockPasswordHasher) {
				m.On("GetByID", mock.Anything, uint64(1)).Return(&models.User{
					ID: 1, Email: "user@example.com", Name: "User", Role: "user",
					PasswordHash: "oldhash", CreatedAt: time.Now(), UpdatedAt: time.Now(),
				}, nil)
				h.On("CheckPassword", "wrongpassword", "oldhash").Return(false)
			},
			wantErr: services.ErrInvalidPassword,
		},
		{
			name:   "error - user not found",
			userID: 999,
			req: &models.ChangePasswordRequest{
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(m *mocks.MockUserRepository, _ *MockPasswordHasher) {
				m.On("GetByID", mock.Anything, uint64(999)).Return(nil, nil)
			},
			wantErr: services.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockHasher := new(MockPasswordHasher)
			tt.setupMock(mockRepo, mockHasher)

			svc := services.NewUserServiceWithHasher(mockRepo, mockHasher)
			err := svc.ChangePassword(context.Background(), tt.userID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
		})
	}
}
