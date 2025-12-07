package services

import (
	"context"
	"errors"

	"nocode-app/backend/internal/models"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/utils"
)

// レコード関連エラー
var (
	ErrRecordNotFound      = errors.New("レコードが見つかりません")
	ErrExternalAppReadOnly = errors.New("外部データソースのアプリは読み取り専用です")
)

// RecordService レコード操作を処理する構造体
type RecordService struct {
	appRepo       repositories.AppRepositoryInterface
	fieldRepo     repositories.FieldRepositoryInterface
	dynamicQuery  repositories.DynamicQueryExecutorInterface
	dsRepo        repositories.DataSourceRepositoryInterface
	externalQuery repositories.ExternalQueryExecutorInterface
}

// NewRecordService 新しいRecordServiceを作成する
func NewRecordService(
	appRepo repositories.AppRepositoryInterface,
	fieldRepo repositories.FieldRepositoryInterface,
	dynamicQuery repositories.DynamicQueryExecutorInterface,
	dsRepo repositories.DataSourceRepositoryInterface,
	externalQuery repositories.ExternalQueryExecutorInterface,
) *RecordService {
	return &RecordService{
		appRepo:       appRepo,
		fieldRepo:     fieldRepo,
		dynamicQuery:  dynamicQuery,
		dsRepo:        dsRepo,
		externalQuery: externalQuery,
	}
}

// GetRecords ページネーションとフィルタリング付きでレコードを取得する
func (s *RecordService) GetRecords(ctx context.Context, appID uint64, opts repositories.RecordQueryOptions) (*models.RecordListResponse, error) {
	// アプリ情報を取得
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// フィールドを取得
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	var records []models.RecordResponse
	var total int64

	// 外部データソースの場合は外部クエリを使用
	if app.IsExternal && app.DataSourceID != nil && app.SourceTableName != nil {
		// 暗号化が初期化されているか確認
		if !utils.IsEncryptionInitialized() {
			return nil, ErrEncryptionNotInitialized
		}

		ds, err := s.dsRepo.GetByID(ctx, *app.DataSourceID)
		if err != nil {
			return nil, err
		}
		if ds == nil {
			return nil, ErrDataSourceNotFound
		}

		password, err := utils.Decrypt(ds.EncryptedPassword)
		if err != nil {
			return nil, err
		}

		records, total, err = s.externalQuery.GetRecords(ctx, ds, password, *app.SourceTableName, fields, opts)
		if err != nil {
			return nil, err
		}
	} else {
		// 内部アプリの場合は動的クエリを使用
		records, total, err = s.dynamicQuery.GetRecords(ctx, app.TableName, fields, opts)
		if err != nil {
			return nil, err
		}
	}

	return &models.RecordListResponse{
		Records:    records,
		Pagination: models.NewPagination(opts.Page, opts.Limit, total),
	}, nil
}

// GetRecord 単一のレコードを取得する
func (s *RecordService) GetRecord(ctx context.Context, appID, recordID uint64) (*models.RecordResponse, error) {
	// アプリ情報を取得
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// フィールドを取得
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	var record *models.RecordResponse

	// 外部データソースの場合は外部クエリを使用
	if app.IsExternal && app.DataSourceID != nil && app.SourceTableName != nil {
		// 暗号化が初期化されているか確認
		if !utils.IsEncryptionInitialized() {
			return nil, ErrEncryptionNotInitialized
		}

		ds, err := s.dsRepo.GetByID(ctx, *app.DataSourceID)
		if err != nil {
			return nil, err
		}
		if ds == nil {
			return nil, ErrDataSourceNotFound
		}

		password, err := utils.Decrypt(ds.EncryptedPassword)
		if err != nil {
			return nil, err
		}

		record, err = s.externalQuery.GetRecordByID(ctx, ds, password, *app.SourceTableName, fields, recordID)
		if err != nil {
			return nil, err
		}
	} else {
		// 内部アプリの場合は動的クエリを使用
		record, err = s.dynamicQuery.GetRecordByID(ctx, app.TableName, fields, recordID)
		if err != nil {
			return nil, err
		}
	}

	if record == nil {
		return nil, ErrRecordNotFound
	}

	return record, nil
}

// CreateRecord 新しいレコードを作成する
func (s *RecordService) CreateRecord(ctx context.Context, appID, userID uint64, req *models.CreateRecordRequest) (*models.RecordResponse, error) {
	// アプリ情報を取得
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// 外部データソースのアプリは読み取り専用
	if app.IsExternal {
		return nil, ErrExternalAppReadOnly
	}

	// レコードを挿入
	recordID, err := s.dynamicQuery.InsertRecord(ctx, app.TableName, req.Data, userID)
	if err != nil {
		return nil, err
	}

	// フィールドを取得して作成したレコードを返す
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return s.dynamicQuery.GetRecordByID(ctx, app.TableName, fields, recordID)
}

// UpdateRecord レコードを更新する
func (s *RecordService) UpdateRecord(ctx context.Context, appID, recordID uint64, req *models.UpdateRecordRequest) (*models.RecordResponse, error) {
	// アプリ情報を取得
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// 外部データソースのアプリは読み取り専用
	if app.IsExternal {
		return nil, ErrExternalAppReadOnly
	}

	// レコードを更新
	if err := s.dynamicQuery.UpdateRecord(ctx, app.TableName, recordID, req.Data); err != nil {
		return nil, err
	}

	// フィールドを取得して更新したレコードを返す
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return s.dynamicQuery.GetRecordByID(ctx, app.TableName, fields, recordID)
}

// DeleteRecord レコードを削除する
func (s *RecordService) DeleteRecord(ctx context.Context, appID, recordID uint64) error {
	// アプリ情報を取得
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrAppNotFound
	}

	// 外部データソースのアプリは読み取り専用
	if app.IsExternal {
		return ErrExternalAppReadOnly
	}

	return s.dynamicQuery.DeleteRecord(ctx, app.TableName, recordID)
}

// BulkCreateRecords 複数のレコードを作成する
func (s *RecordService) BulkCreateRecords(ctx context.Context, appID, userID uint64, req *models.BulkCreateRecordRequest) ([]models.RecordResponse, error) {
	// アプリ情報を取得
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}

	// 外部データソースのアプリは読み取り専用
	if app.IsExternal {
		return nil, ErrExternalAppReadOnly
	}

	// フィールドを取得
	fields, err := s.fieldRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// レコードスライスを事前確保
	records := make([]models.RecordResponse, 0, len(req.Records))
	for _, data := range req.Records {
		recordID, err := s.dynamicQuery.InsertRecord(ctx, app.TableName, data, userID)
		if err != nil {
			return nil, err
		}

		record, err := s.dynamicQuery.GetRecordByID(ctx, app.TableName, fields, recordID)
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
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrAppNotFound
	}

	// 外部データソースのアプリは読み取り専用
	if app.IsExternal {
		return ErrExternalAppReadOnly
	}

	return s.dynamicQuery.DeleteRecords(ctx, app.TableName, req.IDs)
}
