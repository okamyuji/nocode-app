package repositories_test

import (
	"context"
	"os"
	"testing"

	"nocode-app/backend/internal/testhelpers"
)

// TestMain 全リポジトリテスト用のMySQLコンテナをセットアップする
func TestMain(m *testing.M) {
	ctx := context.Background()

	// テストデータベースを初期化（初回呼び出しでコンテナを起動）
	_, err := testhelpers.GetTestDB(ctx)
	if err != nil {
		panic("テストデータベースの初期化に失敗しました: " + err.Error())
	}

	// テストを実行
	code := m.Run()

	// コンテナをクリーンアップ
	if err := testhelpers.CleanupContainer(ctx); err != nil {
		panic("コンテナのクリーンアップに失敗しました: " + err.Error())
	}

	os.Exit(code)
}
