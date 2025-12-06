/**
 * Playwright設定ファイル
 * E2Eテストの実行設定を定義
 */

import { defineConfig, devices } from "@playwright/test";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * 認証状態を保存するファイルパス
 */
export const STORAGE_STATE = path.join(__dirname, ".auth/admin.json");

export default defineConfig({
  testDir: "./tests",

  // global-setupで認証状態を保存
  globalSetup: "./global-setup.ts",

  // テストの並列実行を有効化
  fullyParallel: true,

  // CI環境ではtest.onlyを禁止
  forbidOnly: !!process.env.CI,

  // リトライ設定
  retries: process.env.CI ? 2 : 0,

  // ワーカー数
  workers: process.env.CI ? 1 : undefined,

  // レポーター
  reporter: [["html", { open: "never" }], ["list"]],

  // 共通設定
  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  // テストのタイムアウト
  timeout: 30000,
  expect: {
    timeout: 5000,
  },

  // プロジェクト設定
  projects: [
    // 認証不要のテスト（ログイン・登録画面など）
    {
      name: "unauthenticated",
      testMatch: /.*\.unauth\.spec\.ts/,
      use: { ...devices["Desktop Chrome"] },
    },
    // 認証が必要なテスト
    {
      name: "authenticated",
      testMatch: /.*\.auth\.spec\.ts/,
      use: {
        ...devices["Desktop Chrome"],
        storageState: STORAGE_STATE, // 保存された認証状態を使用
      },
    },
  ],
});
