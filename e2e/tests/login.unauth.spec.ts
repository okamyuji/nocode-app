/**
 * ログイン・登録画面のテスト（未認証）
 */

import { expect, test } from "@playwright/test";

test.describe("ログイン・登録（未認証）", () => {
  test("正しい認証情報でログインできる", async ({ page }) => {
    await page.goto("/login");

    // ログインフォームに入力
    await page
      .getByRole("textbox", { name: "メールアドレス" })
      .fill("admin@example.com");
    await page.getByRole("textbox", { name: "パスワード" }).fill("admin123");
    await page.getByRole("button", { name: "ログイン" }).click();

    // ダッシュボードに遷移していることを確認
    await expect(page).toHaveURL("/");
    await expect(
      page.getByRole("heading", { name: "ダッシュボード" })
    ).toBeVisible();
  });

  test("未認証ユーザーはダッシュボードにアクセスできない", async ({ page }) => {
    // ダッシュボードにアクセス
    await page.goto("/");

    // ログインページにリダイレクトされることを確認
    await expect(page).toHaveURL("/login");
  });

  test("ログインページから登録ページに遷移できる", async ({ page }) => {
    await page.goto("/login");

    // 新規登録リンクをクリック
    await page.getByRole("link", { name: "新規登録" }).click();

    // 登録ページに遷移していることを確認
    await expect(page).toHaveURL("/register");
  });

  test("登録ページからログインページに遷移できる", async ({ page }) => {
    await page.goto("/register");

    // ログインリンクをクリック
    await page.getByRole("link", { name: "ログイン" }).click();

    // ログインページに遷移していることを確認
    await expect(page).toHaveURL("/login");
  });
});
