package services

import (
	"context"
	"errors"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
)

// レコード関連エラー
var (
	ErrRecordNotFound = errors.New("レコードが見つかりません")
)

// RecordService レコード操作を処理する構造体
type RecordService struct {
	appRepo      repositories.AppRepositoryInterface
	fieldRepo    repositories.FieldRepositoryInterface
	dynamicQuery repositories.DynamicQueryExecutorInterface
}

// NewRecordService 新しいRecordServiceを作成する
func NewRecordService(
	appRepo repositories.AppRepositoryInterface,
	fieldRepo repositories.FieldRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
) *RecordService {
	return &RecordService{
		appRepo:      appRepo,
		fieldRepo:    fieldRepo,
		dynamicQuery: dynamicQuery,
	}
}

// GetRecords ページネーションとフィルタリング付きでレコードを取得する
func (s *RecordService) GetRecords(ctx context.Context, appID uint64, opts repositories.RecordQueryOptions) (*models.RecordListResponse, error) {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	// フィールドを取得
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// レコードを取得
	records, total, err := s.dynamicQuery.GetRecords(ctx, tableName, fields, opts)
	if err != nil {
		return nil, err
	}

	return &models.RecordListResponse{
		Records:    records,
		Pagination: models.NewPagination(opts.Page, opts.Limit, total),
	}, nil
}

// GetRecord 単一のレコードを取得する
func (s *RecordService) GetRecord(ctx context.Context, appID, recordID uint64) (*models.RecordResponse, error) {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	// フィールドを取得
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// レコードを取得
	record, err := s.dynamicQuery.GetRecordByID(ctx, tableName, fields, recordID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrRecordNotFound
	}

	return record, nil
}

// CreateRecord 新しいレコードを作成する
func (s *RecordService) CreateRecord(ctx context.Context, appID, userID uint64, req *models.CreateRecordRequest) (*models.RecordResponse, error) {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	// レコードを挿入
	recordID, err := s.dynamicQuery.InsertRecord(ctx, tableName, req.Data, userID)
	if err != nil {
		return nil, err
	}

	// フィールドを取得して作成したレコードを返す
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return s.dynamicQuery.GetRecordByID(ctx, tableName, fields, recordID)
}

// UpdateRecord レコードを更新する
func (s *RecordService) UpdateRecord(ctx context.Context, appID, recordID uint64, req *models.UpdateRecordRequest) (*models.RecordResponse, error) {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	// レコードを更新
	if err := s.dynamicQuery.UpdateRecord(ctx, tableName, recordID, req.Data); err != nil {
		return nil, err
	}

	// フィールドを取得して更新したレコードを返す
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return s.dynamicQuery.GetRecordByID(ctx, tableName, fields, recordID)
}

// DeleteRecord レコードを削除する
func (s *RecordService) DeleteRecord(ctx context.Context, appID, recordID uint64) error {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return err
	}
	if tableName == "" {
		return ErrAppNotFound
	}

	return s.dynamicQuery.DeleteRecord(ctx, tableName, recordID)
}

// BulkCreateRecords 複数のレコードを作成する
func (s *RecordService) BulkCreateRecords(ctx context.Context, appID, userID uint64, req *models.BulkCreateRecordRequest) ([]models.RecordResponse, error) {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, ErrAppNotFound
	}

	// フィールドを取得
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// レコードスライスを事前確保
	records := make([]models.RecordResponse, 0, len(req.Records))
	for _, data := range req.Records {
		recordID, err := s.dynamicQuery.InsertRecord(ctx, tableName, data, userID)
		if err != nil {
			return nil, err
		}

		record, err := s.dynamicQuery.GetRecordByID(ctx, tableName, fields, recordID)
		if err != nil {
			return nil, err
		}

		records = append(records, *record)
	}

	return records, nil
}

// BulkDeleteRecords 複数のレコードを削除する
func (s *RecordService) BulkDeleteRecords(ctx context.Context, appID uint64, req *models.BulkDeleteRecordRequest) error {
	// アプリ情報を取得
	tableName, err := s.appRepo.GetTableName(ctx, appID)
	if err != nil {
		return err
	}
	if tableName == "" {
		return ErrAppNotFound
	}

	return s.dynamicQuery.DeleteRecords(ctx, tableName, req.IDs)
}
