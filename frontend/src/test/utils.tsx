/**
 * テストユーティリティ
 * カスタムレンダー関数とプロバイダーラッパー
 */

import { ApiClientProvider } from "@/api/ApiClientProvider";
import { theme } from "@/theme";
import { ChakraProvider } from "@chakra-ui/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, RenderOptions } from "@testing-library/react";
import { ReactElement, ReactNode } from "react";
import { BrowserRouter } from "react-router-dom";

/**
 * テスト用のQueryClientを作成
 * 各テストで独立したインスタンスを使用
 */
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });

interface AllProvidersProps {
  children: ReactNode;
}

/**
 * テスト用の全プロバイダーラッパー
 * QueryClient、ChakraUI、APIClient、Routerを提供
 */
const AllProviders = ({ children }: AllProvidersProps) => {
  const queryClient = createTestQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      <ChakraProvider theme={theme}>
        <ApiClientProvider>
          <BrowserRouter
            future={{
              v7_startTransition: true,
              v7_relativeSplatPath: true,
            }}
          >
            {children}
          </BrowserRouter>
        </ApiClientProvider>
      </ChakraProvider>
    </QueryClientProvider>
  );
};

/**
 * カスタムレンダー関数
 * 全プロバイダーをラップした状態でコンポーネントをレンダリング
 */
const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, "wrapper">
) => render(ui, { wrapper: AllProviders, ...options });

// testing-libraryの全エクスポートを再エクスポート
export * from "@testing-library/react";
export { createTestQueryClient, customRender as render };
