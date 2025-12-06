/**
 * Playwrightグローバルセットアップ
 * テスト実行前に管理者ユーザーでログインし、認証状態を保存する
 */

import { chromium, type FullConfig } from "@playwright/test";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const STORAGE_STATE = path.join(__dirname, ".auth/admin.json");

/**
 * 管理者の認証情報
 */
const ADMIN = {
  email: "admin@example.com",
  password: "admin123",
};

async function globalSetup(_config: FullConfig) {
  const baseURL = "http://localhost:3000";

  console.log("グローバルセットアップ: 開始...");

  // .authディレクトリを作成
  const authDir = path.dirname(STORAGE_STATE);
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true });
  }

  // バックエンドの起動を待機
  let retries = 30;
  while (retries > 0) {
    try {
      const response = await fetch(`${baseURL}/api/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: "", password: "" }),
      });
      if (response.status === 400 || response.status === 401) {
        console.log("バックエンドが準備完了");
        break;
      }
    } catch {
      // バックエンドがまだ準備できていない
    }
    console.log(`バックエンドを待機中... (残り ${retries} 回)`);
    await new Promise((r) => setTimeout(r, 2000));
    retries--;
  }

  if (retries === 0) {
    throw new Error("バックエンドが準備できませんでした");
  }

  // ブラウザを起動してログイン
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    await page.goto(`${baseURL}/login`);

    // ログインフォームに入力
    await page.getByRole("textbox", { name: "メールアドレス" }).fill(ADMIN.email);
    await page.getByRole("textbox", { name: "パスワード" }).fill(ADMIN.password);
    await page.getByRole("button", { name: "ログイン" }).click();

    // ダッシュボードの要素が表示されるまで待機
    await page
      .getByRole("heading", { name: "ダッシュボード" })
      .waitFor({ timeout: 10000 });

    console.log("ログイン成功、認証状態を保存中...");

    // 認証状態を保存
    await context.storageState({ path: STORAGE_STATE });

    console.log("グローバルセットアップ完了!");
  } catch (error) {
    console.error("グローバルセットアップ失敗:", error);
    throw error;
  } finally {
    await browser.close();
  }
}

export default globalSetup;
