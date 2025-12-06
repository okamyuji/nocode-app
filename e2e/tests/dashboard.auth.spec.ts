/**
 * ダッシュボード画面のテスト（認証済み）
 */

import { expect, test } from "@playwright/test";

test.describe("ダッシュボード（認証済み）", () => {
  test("ダッシュボードが表示される", async ({ page }) => {
    await page.goto("/");

    // ダッシュボードの見出しが表示されることを確認
    await expect(
      page.getByRole("heading", { name: "ダッシュボード" })
    ).toBeVisible();

    // 統計情報が表示されることを確認
    await expect(page.getByText("アプリ数")).toBeVisible();
    await expect(page.getByText("総レコード数")).toBeVisible();
    await expect(page.getByText("ユーザー数")).toBeVisible();
  });

  test("サイドバーからアプリ一覧に遷移できる", async ({ page }) => {
    await page.goto("/");

    // サイドバー内のリンクをクリック
    await page.locator("aside").getByRole("link", { name: "アプリ一覧" }).click();

    // アプリ一覧ページに遷移していることを確認
    await expect(page).toHaveURL("/apps");
    await expect(
      page.getByRole("heading", { name: "アプリ一覧" })
    ).toBeVisible();
  });

  test("サイドバーから設定に遷移できる", async ({ page }) => {
    await page.goto("/");

    // サイドバー内のリンクをクリック
    await page.locator("aside").getByRole("link", { name: "設定" }).click();

    // 設定ページに遷移していることを確認
    await expect(page).toHaveURL(/\/settings/);
  });
});
