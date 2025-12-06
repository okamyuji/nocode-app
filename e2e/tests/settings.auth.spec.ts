/**
 * 設定画面のテスト（認証済み）
 */

import { expect, test } from "@playwright/test";

test.describe("設定（認証済み）", () => {
  test("設定ページが表示される", async ({ page }) => {
    await page.goto("/settings");

    // タブが表示されることを確認
    await expect(page.getByRole("tab", { name: "プロフィール" })).toBeVisible();
    await expect(
      page.getByRole("tab", { name: "パスワード変更" })
    ).toBeVisible();
    await expect(page.getByRole("tab", { name: "ユーザー管理" })).toBeVisible();
  });

  test("プロフィールタブで名前を更新できる", async ({ page }) => {
    await page.goto("/settings?tab=profile");

    // 名前を更新
    const nameInput = page.getByLabel("名前");
    await nameInput.clear();
    await nameInput.fill("Updated Admin");
    await page.getByRole("button", { name: "保存" }).click();

    // トースト通知が表示されることを確認
    await expect(page.locator(".chakra-toast").first()).toBeVisible();

    // 元に戻す
    await nameInput.clear();
    await nameInput.fill("Admin");
    await page.getByRole("button", { name: "保存" }).click();
  });

  test("ユーザー管理タブでユーザー一覧が表示される", async ({ page }) => {
    await page.goto("/settings?tab=users");

    // ユーザー管理の見出しが表示されることを確認
    await expect(
      page.getByRole("heading", { name: "ユーザー管理" })
    ).toBeVisible();

    // テーブル内に管理者ユーザーが表示されることを確認
    await expect(
      page.getByRole("cell", { name: "admin@example.com" })
    ).toBeVisible();
  });

  test("アプリ設定タブが表示される", async ({ page }) => {
    await page.goto("/settings?tab=apps");

    // アプリ設定の見出しが表示されることを確認
    await expect(
      page.getByRole("heading", { name: "アプリ設定" })
    ).toBeVisible();
  });
});
