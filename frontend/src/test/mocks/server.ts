import { setupServer } from "msw/node";
import { handlers } from "./handlers";

// リクエストモックサーバーを設定
export const server = setupServer(...handlers);
