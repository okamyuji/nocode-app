package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter ステータスコードをキャプチャするためのhttp.ResponseWriterラッパー
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

// newResponseWriter 新しいresponseWriterを作成
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader ステータスコードを書き込み、記録
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write レスポンスボディを書き込み、バイト数を記録
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggerMiddleware ロギングミドルウェアを作成
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// レスポンスライターをラップ
		rw := newResponseWriter(w)

		// リクエストを処理
		next.ServeHTTP(rw, r)

		// 処理時間を計算
		duration := time.Since(start)

		// リクエストをログ出力
		log.Printf(
			"[%s] %s %s %d %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rw.statusCode,
			rw.written,
			duration,
		)
	})
}

// RecoveryMiddleware パニックリカバリーミドルウェアを作成
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
