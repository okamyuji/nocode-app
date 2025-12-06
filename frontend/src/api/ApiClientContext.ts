/**
 * APIクライアント コンテキスト定義
 */

import { createContext } from "react";
import type { IApiClient } from "./interfaces";

export const ApiClientContext = createContext<IApiClient | null>(null);
