/**
 * アプリ管理画面のテスト（認証済み）
 */

import { expect, test } from "@playwright/test";

test.describe("アプリ管理（認証済み）", () => {
  test("アプリ一覧が表示される", async ({ page }) => {
    await page.goto("/apps");

    // アプリ一覧の見出しが表示されることを確認
    await expect(
      page.getByRole("heading", { name: "アプリ一覧" })
    ).toBeVisible();
  });

  test("アプリ作成ページが表示される", async ({ page }) => {
    await page.goto("/apps/new");

    // アプリ作成画面の要素が表示されることを確認
    await expect(
      page.getByRole("heading", { name: "アプリを作成" })
    ).toBeVisible();
    await expect(page.getByText("フィールドパレット")).toBeVisible();
    await expect(page.getByText("基本情報")).toBeVisible();
  });

  test("アプリカードをクリックするとレコード画面に遷移する", async ({ page }) => {
    await page.goto("/apps");

    // 最初のアプリカードをクリック（存在する場合）
    const appCard = page.locator(".chakra-card").first();
    if (await appCard.isVisible({ timeout: 3000 })) {
      await appCard.click();

      // レコード画面に遷移していることを確認
      await expect(page).toHaveURL(/\/apps\/\d+\/records/);
    }
  });
});
