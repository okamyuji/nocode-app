package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/testhelpers/mocks"
	"nocode-app/backend/internal/utils"
)

func setupEncryption(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	err := utils.SetEncryptionKey(key)
	require.NoError(t, err)
}

func TestDataSourceService_CreateDataSource(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("NameExists", ctx, "test-ds").Return(false, nil)
		mockDSRepo.On("Create", ctx, mock.AnythingOfType("*models.DataSource")).Return(nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.CreateDataSourceRequest{
			Name:         "test-ds",
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		resp, err := service.CreateDataSource(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "test-ds", resp.Name)
		assert.Equal(t, models.DBTypePostgreSQL, resp.DBType)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("name already exists", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("NameExists", ctx, "existing-ds").Return(true, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.CreateDataSourceRequest{
			Name:         "existing-ds",
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		_, err := service.CreateDataSource(ctx, 1, req)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNameExists, err)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("invalid db type", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("NameExists", ctx, "test-ds").Return(false, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.CreateDataSourceRequest{
			Name:         "test-ds",
			DBType:       "mongodb", // invalid
			Host:         "localhost",
			Port:         27017,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		_, err := service.CreateDataSource(ctx, 1, req)
		require.Error(t, err)
		assert.Equal(t, services.ErrInvalidDBType, err)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_GetDataSource(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		ds := &models.DataSource{
			ID:           1,
			Name:         "test-ds",
			DBType:       models.DBTypePostgreSQL,
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			CreatedBy:    1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		resp, err := service.GetDataSource(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, "test-ds", resp.Name)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		_, err := service.GetDataSource(ctx, 999)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNotFound, err)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_GetDataSources(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		dataSources := []models.DataSource{
			{ID: 1, Name: "ds1", DBType: models.DBTypePostgreSQL},
			{ID: 2, Name: "ds2", DBType: models.DBTypeMySQL},
		}

		mockDSRepo.On("GetAll", ctx, 1, 20).Return(dataSources, int64(2), nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		resp, err := service.GetDataSources(ctx, 1, 20)
		require.NoError(t, err)
		assert.Len(t, resp.DataSources, 2)
		assert.Equal(t, int64(2), resp.Pagination.Total)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_UpdateDataSource(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		ds := &models.DataSource{
			ID:                1,
			Name:              "old-name",
			DBType:            models.DBTypePostgreSQL,
			Host:              "localhost",
			Port:              5432,
			DatabaseName:      "testdb",
			Username:          "testuser",
			EncryptedPassword: "encrypted",
			CreatedBy:         1,
		}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)
		mockDSRepo.On("NameExistsExcludingDataSource", ctx, "new-name", uint64(1)).Return(false, nil)
		mockDSRepo.On("Update", ctx, mock.AnythingOfType("*models.DataSource")).Return(nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.UpdateDataSourceRequest{
			Name: "new-name",
			Host: "newhost",
		}

		resp, err := service.UpdateDataSource(ctx, 1, req)
		require.NoError(t, err)
		assert.Equal(t, "new-name", resp.Name)
		assert.Equal(t, "newhost", resp.Host)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.UpdateDataSourceRequest{
			Name: "new-name",
		}

		_, err := service.UpdateDataSource(ctx, 999, req)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNotFound, err)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("name conflict", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		ds := &models.DataSource{
			ID:   1,
			Name: "old-name",
		}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)
		mockDSRepo.On("NameExistsExcludingDataSource", ctx, "existing-name", uint64(1)).Return(true, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.UpdateDataSourceRequest{
			Name: "existing-name",
		}

		_, err := service.UpdateDataSource(ctx, 1, req)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNameExists, err)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_DeleteDataSource(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		ds := &models.DataSource{ID: 1, Name: "test-ds"}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)
		mockDSRepo.On("Delete", ctx, uint64(1)).Return(nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		err := service.DeleteDataSource(ctx, 1)
		require.NoError(t, err)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		err := service.DeleteDataSource(ctx, 999)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNotFound, err)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_TestConnection(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful connection", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockExtQuery.On("TestConnection", ctx, mock.AnythingOfType("*models.DataSource"), "testpass").Return(nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.TestConnectionRequest{
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		resp, err := service.TestConnection(ctx, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "接続に成功しました", resp.Message)

		mockExtQuery.AssertExpectations(t)
	})

	t.Run("connection failed", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockExtQuery.On("TestConnection", ctx, mock.AnythingOfType("*models.DataSource"), "testpass").
			Return(errors.New("connection refused"))

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.TestConnectionRequest{
			DBType:       "postgresql",
			Host:         "localhost",
			Port:         5432,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		resp, err := service.TestConnection(ctx, req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "connection refused")

		mockExtQuery.AssertExpectations(t)
	})

	t.Run("invalid db type", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		req := &models.TestConnectionRequest{
			DBType:       "mongodb",
			Host:         "localhost",
			Port:         27017,
			DatabaseName: "testdb",
			Username:     "testuser",
			Password:     "testpass",
		}

		resp, err := service.TestConnection(ctx, req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "無効なデータベースタイプ")
	})
}

func TestDataSourceService_GetTables(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful get tables", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		encryptedPass, _ := utils.Encrypt("testpass")
		ds := &models.DataSource{
			ID:                1,
			Name:              "test-ds",
			DBType:            models.DBTypePostgreSQL,
			EncryptedPassword: encryptedPass,
		}

		tables := []models.TableInfo{
			{Name: "users", Schema: "public"},
			{Name: "orders", Schema: "public"},
		}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)
		mockExtQuery.On("GetTables", ctx, ds, "testpass").Return(tables, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		resp, err := service.GetTables(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, resp.Tables, 2)

		mockDSRepo.AssertExpectations(t)
		mockExtQuery.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		_, err := service.GetTables(ctx, 999)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNotFound, err)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_GetColumns(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful get columns", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		encryptedPass, _ := utils.Encrypt("testpass")
		ds := &models.DataSource{
			ID:                1,
			Name:              "test-ds",
			DBType:            models.DBTypePostgreSQL,
			EncryptedPassword: encryptedPass,
		}

		columns := []models.ColumnInfo{
			{Name: "id", DataType: "integer", IsPrimaryKey: true},
			{Name: "name", DataType: "varchar", IsNullable: true},
		}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)
		mockExtQuery.On("GetColumns", ctx, ds, "testpass", "users").Return(columns, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		resp, err := service.GetColumns(ctx, 1, "users")
		require.NoError(t, err)
		assert.Len(t, resp.Columns, 2)
		assert.True(t, resp.Columns[0].IsPrimaryKey)

		mockDSRepo.AssertExpectations(t)
		mockExtQuery.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		_, err := service.GetColumns(ctx, 999, "users")
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNotFound, err)

		mockDSRepo.AssertExpectations(t)
	})
}

func TestDataSourceService_GetDecryptedPassword(t *testing.T) {
	setupEncryption(t)
	ctx := context.Background()

	t.Run("successful decrypt", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		encryptedPass, _ := utils.Encrypt("mypassword")
		ds := &models.DataSource{
			ID:                1,
			EncryptedPassword: encryptedPass,
		}

		mockDSRepo.On("GetByID", ctx, uint64(1)).Return(ds, nil)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		password, err := service.GetDecryptedPassword(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, "mypassword", password)

		mockDSRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockDSRepo := new(mocks.MockDataSourceRepository)
		mockExtQuery := new(mocks.MockExternalQueryExecutor)

		mockDSRepo.On("GetByID", ctx, uint64(999)).Return(nil, sql.ErrNoRows)

		service := services.NewDataSourceService(mockDSRepo, mockExtQuery)

		_, err := service.GetDecryptedPassword(ctx, 999)
		require.Error(t, err)
		assert.Equal(t, services.ErrDataSourceNotFound, err)

		mockDSRepo.AssertExpectations(t)
	})
}
