package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nocode-app/backend/internal/config"
	"nocode-app/backend/internal/database"
	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/middleware"
	"nocode-app/backend/internal/repositories"
	"nocode-app/backend/internal/router"
	"nocode-app/backend/internal/services"
	"nocode-app/backend/internal/utils"
)

func main() {
	// 設定の読み込み
	cfg := config.Load()

	// データベースの起動を待機（コンテナ起動時に有用）
	if err := waitForDB(&cfg.DB, 30); err != nil {
		log.Fatalf("データベース接続に失敗しました（リトライ後）: %v", err)
	}

	// データベースの初期化
	db, err := database.NewDB(&cfg.DB)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Printf("データベースクローズエラー: %v", err)
		}
	}()

	log.Println("データベースに接続しました")

	// 暗号化の初期化
	if err := utils.InitEncryption(); err != nil {
		log.Printf("警告: 暗号化キーの初期化に失敗しました: %v", err)
		log.Println("外部データソース機能は利用できません。ENCRYPTION_KEY環境変数を設定してください。")
	}

	// ユーティリティの初期化
	jwtManager := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	validator := utils.NewValidator()

	// リポジトリの初期化
	userRepo := repositories.NewUserRepository(db)
	appRepo := repositories.NewAppRepository(db)
	fieldRepo := repositories.NewFieldRepository(db)
	viewRepo := repositories.NewViewRepository(db)
	chartRepo := repositories.NewChartRepository(db)
	dynamicQuery := repositories.NewDynamicQueryExecutor(db)
	dataSourceRepo := repositories.NewDataSourceRepository(db)
	externalQuery := repositories.NewExternalQueryExecutor()

	// サービスの初期化
	authService := services.NewAuthService(userRepo, jwtManager)
	appService := services.NewAppService(appRepo, fieldRepo, dynamicQuery, dataSourceRepo)
	fieldService := services.NewFieldService(fieldRepo, appRepo, dynamicQuery)
	recordService := services.NewRecordService(appRepo, fieldRepo, dynamicQuery, dataSourceRepo, externalQuery)
	viewService := services.NewViewService(viewRepo, appRepo)
	chartService := services.NewChartService(chartRepo, appRepo, fieldRepo, dynamicQuery, dataSourceRepo, externalQuery)
	userService := services.NewUserService(userRepo)
	dashboardService := services.NewDashboardService(userRepo, appRepo, dynamicQuery)
	dataSourceService := services.NewDataSourceService(dataSourceRepo, externalQuery)

	// ハンドラーの初期化
	authHandler := handlers.NewAuthHandler(authService, validator)
	appHandler := handlers.NewAppHandler(appService, validator)
	fieldHandler := handlers.NewFieldHandler(fieldService, validator)
	recordHandler := handlers.NewRecordHandler(recordService, validator)
	viewHandler := handlers.NewViewHandler(viewService, validator)
	chartHandler := handlers.NewChartHandler(chartService, validator)
	userHandler := handlers.NewUserHandler(userService, validator)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	dataSourceHandler := handlers.NewDataSourceHandler(dataSourceService, validator)

	// ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	corsConfig := &middleware.CORSConfig{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
	}

	// ルーターの初期化
	r := router.NewRouter(
		authMiddleware,
		corsConfig,
		authHandler,
		appHandler,
		fieldHandler,
		recordHandler,
		viewHandler,
		chartHandler,
		userHandler,
		dashboardHandler,
		dataSourceHandler,
	)

	// ルートの設定
	handler := r.Setup()

	// サーバーの作成
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// サーバーをgoroutineで起動
	go func() {
		log.Printf("サーバーをポート%sで起動します", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバー起動に失敗しました: %v", err)
		}
	}()

	// 割り込みシグナルを待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("サーバーをシャットダウンします...")

	// タイムアウト付きグレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("サーバーの強制シャットダウン: %v", err)
	}

	log.Println("サーバーを停止しました")
}

// waitForDB データベースの起動を待機する（コンテナ起動時用）
func waitForDB(cfg *config.DBConfig, maxAttempts int) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		db, err := database.NewDB(cfg)
		if err == nil {
			if closeErr := database.Close(db); closeErr != nil {
				log.Printf("テスト接続のクローズエラー: %v", closeErr)
			}
			return nil
		}
		lastErr = err
		log.Printf("データベースを待機中... %d/%d回目", i+1, maxAttempts)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("%d回の試行後に接続に失敗しました: %w", maxAttempts, lastErr)
}
