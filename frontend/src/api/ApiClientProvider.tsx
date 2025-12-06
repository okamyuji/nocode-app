/**
 * APIクライアント コンテキストプロバイダー コンポーネント
 */

import { ReactNode } from "react";
import { ApiClientContext } from "./ApiClientContext";
import { apiClient as defaultApiClient } from "./apiClientInstance";
import type { IApiClient } from "./interfaces";

interface ApiClientProviderProps {
  children: ReactNode;
  client?: IApiClient;
}

/**
 * APIクライアントの依存性注入用プロバイダーコンポーネント
 * @param children - 子コンポーネント
 * @param client - カスタムAPIクライアント（テスト/モック用、省略可能）
 */
export function ApiClientProvider({
  children,
  client = defaultApiClient,
}: ApiClientProviderProps) {
  return (
    <ApiClientContext.Provider value={client}>
      {children}
    </ApiClientContext.Provider>
  );
}
