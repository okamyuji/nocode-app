# ビルドステージ
FROM node:22-alpine AS builder

# pnpm用にcorepackを有効化
RUN corepack enable && corepack prepare pnpm@10.24.0 --activate

WORKDIR /app

# パッケージファイルをコピー
COPY package.json pnpm-lock.yaml* ./

# 依存関係をインストール
RUN pnpm install --frozen-lockfile

# ソースコードをコピー
COPY . .

# 本番用API URLを設定（nginxプロキシ経由）
ARG VITE_API_URL=/api/v1
ENV VITE_API_URL=$VITE_API_URL

# アプリケーションをビルド
RUN pnpm run build

# 本番ステージ
FROM nginx:alpine

# ビルド済みファイルをコピー
COPY --from=builder /app/dist /usr/share/nginx/html

# nginx設定ファイルをコピー
COPY nginx.conf /etc/nginx/conf.d/default.conf

# ポートを公開
EXPOSE 80

# nginxを起動
CMD ["nginx", "-g", "daemon off;"]
