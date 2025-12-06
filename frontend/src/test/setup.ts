/**
 * テストセットアップファイル
 * Vitestのグローバル設定とモックの初期化
 */

import "@testing-library/jest-dom";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./mocks/server";

/**
 * Chakra UIコンポーネント用のwindow.matchMediaモック
 */
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {}, // 非推奨
    removeListener: () => {}, // 非推奨
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
});

/**
 * Chakra UIコンポーネント用のResizeObserverモック
 */
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
window.ResizeObserver = ResizeObserverMock;

// 全テスト実行前にAPIモックサーバーを起動
beforeAll(() => server.listen({ onUnhandledRequest: "warn" }));

// 各テスト終了後にリクエストハンドラーをリセット
// テスト間の影響を防ぐため
afterEach(() => {
  cleanup();
  server.resetHandlers();
});

// 全テスト終了後にモックサーバーを停止
afterAll(() => server.close());
