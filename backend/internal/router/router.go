package router

import (
	"net/http"
	"strings"

	"nocode-app/backend/internal/handlers"
	"nocode-app/backend/internal/middleware"
)

// Router HTTPルーティングを処理する構造体
type Router struct {
	mux            *http.ServeMux
	authMiddleware *middleware.AuthMiddleware
	corsConfig     *middleware.CORSConfig

	// ハンドラー
	authHandler       *handlers.AuthHandler
	appHandler        *handlers.AppHandler
	fieldHandler      *handlers.FieldHandler
	recordHandler     *handlers.RecordHandler
	viewHandler       *handlers.ViewHandler
	chartHandler      *handlers.ChartHandler
	userHandler       *handlers.UserHandler
	dashboardHandler  *handlers.DashboardHandler
	dataSourceHandler *handlers.DataSourceHandler
}

// NewRouter 新しいRouterを作成する
func NewRouter(
	authMiddleware *middleware.AuthMiddleware,
	corsConfig *middleware.CORSConfig,
	authHandler *handlers.AuthHandler,
	appHandler *handlers.AppHandler,
	fieldHandler *handlers.FieldHandler,
	recordHandler *handlers.RecordHandler,
	viewHandler *handlers.ViewHandler,
	chartHandler *handlers.ChartHandler,
	userHandler *handlers.UserHandler,
	dashboardHandler *handlers.DashboardHandler,
	dataSourceHandler *handlers.DataSourceHandler,
) *Router {
	return &Router{
		mux:               http.NewServeMux(),
		authMiddleware:    authMiddleware,
		corsConfig:        corsConfig,
		authHandler:       authHandler,
		appHandler:        appHandler,
		fieldHandler:      fieldHandler,
		recordHandler:     recordHandler,
		viewHandler:       viewHandler,
		chartHandler:      chartHandler,
		userHandler:       userHandler,
		dashboardHandler:  dashboardHandler,
		dataSourceHandler: dataSourceHandler,
	}
}

// Setup 全ルートをセットアップする
func (r *Router) Setup() http.Handler {
	// ヘルスチェック
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			http.Error(w, "レスポンスの書き込みに失敗しました", http.StatusInternalServerError)
		}
	})

	// APIルート
	r.mux.HandleFunc("/api/v1/", r.routeAPI)

	// ミドルウェアを適用
	handler := middleware.LoggerMiddleware(r.mux)
	handler = middleware.CORSMiddleware(r.corsConfig)(handler)
	handler = middleware.RecoveryMiddleware(handler)

	return handler
}

// routeAPI 全APIリクエストをルーティングする
func (r *Router) routeAPI(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// 認証ルート（登録/ログインは認証不要）
	if strings.HasPrefix(path, "/api/v1/auth/") {
		r.routeAuth(w, req)
		return
	}

	// 保護されたルート（認証必須）
	r.authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.routeProtected(w, req)
	})).ServeHTTP(w, req)
}

// routeAuth 認証エンドポイントをルーティングする
func (r *Router) routeAuth(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	switch path {
	case "/api/v1/auth/register":
		r.authHandler.Register(w, req)
	case "/api/v1/auth/login":
		r.authHandler.Login(w, req)
	case "/api/v1/auth/me":
		// /meは認証必須
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Me)).ServeHTTP(w, req)
	case "/api/v1/auth/refresh":
		// /refreshは認証必須
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Refresh)).ServeHTTP(w, req)
	case "/api/v1/auth/profile":
		// /profileは認証必須
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.UpdateProfile)).ServeHTTP(w, req)
	case "/api/v1/auth/password":
		// /passwordは認証必須
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.ChangePassword)).ServeHTTP(w, req)
	default:
		http.NotFound(w, req)
	}
}

// routeProtected 保護されたエンドポイントをルーティングする
func (r *Router) routeProtected(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// ダッシュボードルート
	if path == "/api/v1/dashboard/stats" {
		r.dashboardHandler.GetStats(w, req)
		return
	}

	// アプリルート
	if strings.HasPrefix(path, "/api/v1/apps") {
		r.routeApps(w, req)
		return
	}

	// ユーザールート（管理者専用）
	if strings.HasPrefix(path, "/api/v1/users") {
		r.routeUsers(w, req)
		return
	}

	// データソースルート（管理者専用）
	if strings.HasPrefix(path, "/api/v1/datasources") {
		r.routeDataSources(w, req)
		return
	}

	http.NotFound(w, req)
}

// routeUsers ユーザー管理エンドポイントをルーティングする
func (r *Router) routeUsers(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// /api/v1/users
	if len(parts) == 3 {
		switch req.Method {
		case http.MethodGet:
			r.userHandler.List(w, req)
		case http.MethodPost:
			r.userHandler.Create(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/users/{id}
	if len(parts) == 4 {
		switch req.Method {
		case http.MethodGet:
			r.userHandler.Get(w, req)
		case http.MethodPut:
			r.userHandler.Update(w, req)
		case http.MethodDelete:
			r.userHandler.Delete(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, req)
}

// routeApps アプリ関連エンドポイントをルーティングする
func (r *Router) routeApps(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// /api/v1/apps/external
	if len(parts) == 4 && parts[3] == "external" {
		if req.Method == http.MethodPost {
			// 管理者専用: 外部データソースからアプリ作成
			middleware.RequireAdmin(r.appHandler.CreateExternal)(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/apps
	if len(parts) == 3 {
		switch req.Method {
		case http.MethodGet:
			r.appHandler.List(w, req)
		case http.MethodPost:
			// 管理者専用: アプリ作成
			middleware.RequireAdmin(r.appHandler.Create)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}
	if len(parts) == 4 {
		switch req.Method {
		case http.MethodGet:
			r.appHandler.Get(w, req)
		case http.MethodPut:
			// 管理者専用: アプリ更新
			middleware.RequireAdmin(r.appHandler.Update)(w, req)
		case http.MethodDelete:
			// 管理者専用: アプリ削除
			middleware.RequireAdmin(r.appHandler.Delete)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}/fields, /api/v1/apps/{id}/records など
	if len(parts) >= 5 {
		resource := parts[4]

		switch resource {
		case "fields":
			r.routeFields(w, req, parts)
		case "records":
			r.routeRecords(w, req, parts)
		case "views":
			r.routeViews(w, req, parts)
		case "charts":
			r.routeCharts(w, req, parts)
		default:
			http.NotFound(w, req)
		}
		return
	}

	http.NotFound(w, req)
}

// routeFields フィールドエンドポイントをルーティングする
func (r *Router) routeFields(w http.ResponseWriter, req *http.Request, parts []string) {
	// /api/v1/apps/{id}/fields
	if len(parts) == 5 {
		switch req.Method {
		case http.MethodGet:
			r.fieldHandler.List(w, req)
		case http.MethodPost:
			// 管理者専用: フィールド作成
			middleware.RequireAdmin(r.fieldHandler.Create)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}/fields/order
	if len(parts) == 6 && parts[5] == "order" {
		if req.Method == http.MethodPut {
			// 管理者専用: フィールド順序更新
			middleware.RequireAdmin(r.fieldHandler.UpdateOrder)(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/apps/{id}/fields/{fieldId}
	if len(parts) == 6 {
		switch req.Method {
		case http.MethodPut:
			// 管理者専用: フィールド更新
			middleware.RequireAdmin(r.fieldHandler.Update)(w, req)
		case http.MethodDelete:
			// 管理者専用: フィールド削除
			middleware.RequireAdmin(r.fieldHandler.Delete)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, req)
}

// routeRecords レコードエンドポイントをルーティングする
func (r *Router) routeRecords(w http.ResponseWriter, req *http.Request, parts []string) {
	// /api/v1/apps/{id}/records
	if len(parts) == 5 {
		switch req.Method {
		case http.MethodGet:
			r.recordHandler.List(w, req)
		case http.MethodPost:
			// 管理者専用: レコード作成
			middleware.RequireAdmin(r.recordHandler.Create)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}/records/bulk
	if len(parts) == 6 && parts[5] == "bulk" {
		switch req.Method {
		case http.MethodPost:
			// 管理者専用: レコード一括作成
			middleware.RequireAdmin(r.recordHandler.BulkCreate)(w, req)
		case http.MethodDelete:
			// 管理者専用: レコード一括削除
			middleware.RequireAdmin(r.recordHandler.BulkDelete)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}/records/{recordId}
	if len(parts) == 6 {
		switch req.Method {
		case http.MethodGet:
			r.recordHandler.Get(w, req)
		case http.MethodPut:
			// 管理者専用: レコード更新
			middleware.RequireAdmin(r.recordHandler.Update)(w, req)
		case http.MethodDelete:
			// 管理者専用: レコード削除
			middleware.RequireAdmin(r.recordHandler.Delete)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, req)
}

// routeViews ビューエンドポイントをルーティングする
func (r *Router) routeViews(w http.ResponseWriter, req *http.Request, parts []string) {
	// /api/v1/apps/{id}/views
	if len(parts) == 5 {
		switch req.Method {
		case http.MethodGet:
			r.viewHandler.List(w, req)
		case http.MethodPost:
			// 管理者専用: ビュー作成
			middleware.RequireAdmin(r.viewHandler.Create)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}/views/{viewId}
	if len(parts) == 6 {
		switch req.Method {
		case http.MethodPut:
			// 管理者専用: ビュー更新
			middleware.RequireAdmin(r.viewHandler.Update)(w, req)
		case http.MethodDelete:
			// 管理者専用: ビュー削除
			middleware.RequireAdmin(r.viewHandler.Delete)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, req)
}

// routeCharts チャートエンドポイントをルーティングする
func (r *Router) routeCharts(w http.ResponseWriter, req *http.Request, parts []string) {
	// /api/v1/apps/{id}/charts/data
	if len(parts) == 6 && parts[5] == "data" {
		if req.Method == http.MethodPost {
			r.chartHandler.GetData(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/apps/{id}/charts/config
	if len(parts) == 6 && parts[5] == "config" {
		switch req.Method {
		case http.MethodGet:
			r.chartHandler.GetConfigs(w, req)
		case http.MethodPost:
			// 管理者専用: チャート設定保存
			middleware.RequireAdmin(r.chartHandler.SaveConfig)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/apps/{id}/charts/config/{configId}
	if len(parts) == 7 && parts[5] == "config" {
		if req.Method == http.MethodDelete {
			// 管理者専用: チャート設定削除
			middleware.RequireAdmin(r.chartHandler.DeleteConfig)(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, req)
}

// routeDataSources データソースエンドポイントをルーティングする
func (r *Router) routeDataSources(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// /api/v1/datasources/test（テスト接続）
	if len(parts) == 4 && parts[3] == "test" {
		if req.Method == http.MethodPost {
			middleware.RequireAdmin(r.dataSourceHandler.TestConnection)(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/datasources
	if len(parts) == 3 {
		switch req.Method {
		case http.MethodGet:
			middleware.RequireAdmin(r.dataSourceHandler.List)(w, req)
		case http.MethodPost:
			middleware.RequireAdmin(r.dataSourceHandler.Create)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/datasources/{id}
	if len(parts) == 4 {
		switch req.Method {
		case http.MethodGet:
			middleware.RequireAdmin(r.dataSourceHandler.Get)(w, req)
		case http.MethodPut:
			middleware.RequireAdmin(r.dataSourceHandler.Update)(w, req)
		case http.MethodDelete:
			middleware.RequireAdmin(r.dataSourceHandler.Delete)(w, req)
		default:
			http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/datasources/{id}/tables
	if len(parts) == 5 && parts[4] == "tables" {
		if req.Method == http.MethodGet {
			middleware.RequireAdmin(r.dataSourceHandler.GetTables)(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/datasources/{id}/tables/{tableName}/columns
	if len(parts) == 7 && parts[4] == "tables" && parts[6] == "columns" {
		if req.Method == http.MethodGet {
			middleware.RequireAdmin(r.dataSourceHandler.GetColumns)(w, req)
			return
		}
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, req)
}
