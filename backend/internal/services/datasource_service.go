package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/utils"
)

// データソース関連エラー
var (
	ErrDataSourceNotFound       = errors.New("データソースが見つかりません")
	ErrDataSourceNameExists     = errors.New("同じ名前のデータソースが既に存在します")
	ErrConnectionFailed         = errors.New("データベースへの接続に失敗しました")
	ErrInvalidDBType            = errors.New("無効なデータベースタイプです")
	ErrEncryptionNotInitialized = errors.New("外部データソース機能は利用できません。ENCRYPTION_KEY環境変数を設定してください")
)

// DataSourceService データソース操作を処理する構造体
type DataSourceService struct {
	dsRepo        repositories.DataSourceRepositoryInterface
	externalQuery repositories.ExternalQueryExecutorInterface
}

// NewDataSourceService 新しいDataSourceServiceを作成する
func NewDataSourceService(
	dsRepo repositories.DataSourceRepositoryInterface,
	externalQuery repositories.ExternalQueryExecutorInterface,
) *DataSourceService {
	return &DataSourceService{
		dsRepo:        dsRepo,
		externalQuery: externalQuery,
	}
}

// CreateDataSource 新しいデータソースを作成する
func (s *DataSourceService) CreateDataSource(ctx context.Context, userID uint64, req *models.CreateDataSourceRequest) (*models.DataSourceResponse, error) {
	// 暗号化が初期化されているか確認
	if !utils.IsEncryptionInitialized() {
		return nil, ErrEncryptionNotInitialized
	}

	// 名前の重複チェック
	exists, err := s.dsRepo.NameExists(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDataSourceNameExists
	}

	// DBタイプの検証
	if !models.IsValidDBType(req.DBType) {
		return nil, ErrInvalidDBType
	}

	// パスワードを暗号化
	encryptedPassword, err := utils.Encrypt(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ds := &models.DataSource{
		Name:              req.Name,
		DBType:            models.DBType(req.DBType),
		Host:              req.Host,
		Port:              req.Port,
		DatabaseName:      req.DatabaseName,
		Username:          req.Username,
		EncryptedPassword: encryptedPassword,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.dsRepo.Create(ctx, ds); err != nil {
		return nil, err
	}

	return ds.ToResponse(), nil
}

// GetDataSource IDでデータソースを取得する
func (s *DataSourceService) GetDataSource(ctx context.Context, id uint64) (*models.DataSourceResponse, error) {
	ds, err := s.dsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDataSourceNotFound
		}
		return nil, err
	}
	return ds.ToResponse(), nil
}

// GetDataSources ページネーション付きで全データソースを取得する
func (s *DataSourceService) GetDataSources(ctx context.Context, page, limit int) (*models.DataSourceListResponse, error) {
	dataSources, total, err := s.dsRepo.GetAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]models.DataSourceResponse, len(dataSources))
	for i := range dataSources {
		responses[i] = *dataSources[i].ToResponse()
	}

	return &models.DataSourceListResponse{
		DataSources: responses,
		Pagination:  models.NewPagination(page, limit, total),
	}, nil
}

// UpdateDataSource データソースを更新する
func (s *DataSourceService) UpdateDataSource(ctx context.Context, id uint64, req *models.UpdateDataSourceRequest) (*models.DataSourceResponse, error) {
	ds, err := s.dsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDataSourceNotFound
		}
		return nil, err
	}

	// 名前の重複チェック（自分自身を除く）
	if req.Name != "" && req.Name != ds.Name {
		exists, err := s.dsRepo.NameExistsExcludingDataSource(ctx, req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrDataSourceNameExists
		}
		ds.Name = req.Name
	}

	if req.Host != "" {
		ds.Host = req.Host
	}
	if req.Port > 0 {
		ds.Port = req.Port
	}
	if req.DatabaseName != "" {
		ds.DatabaseName = req.DatabaseName
	}
	if req.Username != "" {
		ds.Username = req.Username
	}
	if req.Password != "" {
		encryptedPassword, err := utils.Encrypt(req.Password)
		if err != nil {
			return nil, err
		}
		ds.EncryptedPassword = encryptedPassword
	}

	ds.UpdatedAt = time.Now()

	if err := s.dsRepo.Update(ctx, ds); err != nil {
		return nil, err
	}

	return ds.ToResponse(), nil
}

// DeleteDataSource データソースを削除する
func (s *DataSourceService) DeleteDataSource(ctx context.Context, id uint64) error {
	_, err := s.dsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDataSourceNotFound
		}
		return err
	}

	return s.dsRepo.Delete(ctx, id)
}

// TestConnection テスト接続を実行する
func (s *DataSourceService) TestConnection(ctx context.Context, req *models.TestConnectionRequest) (*models.TestConnectionResponse, error) {
	// DBタイプの検証
	if !models.IsValidDBType(req.DBType) {
		return &models.TestConnectionResponse{
			Success: false,
			Message: "無効なデータベースタイプです",
		}, nil
	}

	// テスト用の一時データソースを作成
	ds := &models.DataSource{
		DBType:       models.DBType(req.DBType),
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
	}

	err := s.externalQuery.TestConnection(ctx, ds, req.Password)
	if err != nil {
		return &models.TestConnectionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.TestConnectionResponse{
		Success: true,
		Message: "接続に成功しました",
	}, nil
}

// GetTables データソースのテーブル一覧を取得する
func (s *DataSourceService) GetTables(ctx context.Context, id uint64) (*models.TableListResponse, error) {
	// 暗号化が初期化されているか確認
	if !utils.IsEncryptionInitialized() {
		return nil, ErrEncryptionNotInitialized
	}

	ds, err := s.dsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDataSourceNotFound
		}
		return nil, err
	}

	// パスワードを復号
	password, err := utils.Decrypt(ds.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	tables, err := s.externalQuery.GetTables(ctx, ds, password)
	if err != nil {
		return nil, err
	}

	return &models.TableListResponse{
		Tables: tables,
	}, nil
}

// GetColumns テーブルのカラム一覧を取得する
func (s *DataSourceService) GetColumns(ctx context.Context, id uint64, tableName string) (*models.ColumnListResponse, error) {
	// 暗号化が初期化されているか確認
	if !utils.IsEncryptionInitialized() {
		return nil, ErrEncryptionNotInitialized
	}

	ds, err := s.dsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDataSourceNotFound
		}
		return nil, err
	}

	// パスワードを復号
	password, err := utils.Decrypt(ds.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	columns, err := s.externalQuery.GetColumns(ctx, ds, password, tableName)
	if err != nil {
		return nil, err
	}

	return &models.ColumnListResponse{
		Columns: columns,
	}, nil
}

// GetDecryptedPassword データソースの復号化されたパスワードを取得する（内部使用）
func (s *DataSourceService) GetDecryptedPassword(ctx context.Context, id uint64) (string, error) {
	ds, err := s.dsRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrDataSourceNotFound
		}
		return "", err
	}

	return utils.Decrypt(ds.EncryptedPassword)
}
