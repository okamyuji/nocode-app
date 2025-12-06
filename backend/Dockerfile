# ビルドステージ
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 依存関係をインストール
RUN apk add --no-cache git

# goモジュールファイルをコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 実行ステージ
FROM alpine:3.19

WORKDIR /app

# HTTPS用のca証明書とタイムゾーンデータをインストール
RUN apk --no-cache add ca-certificates tzdata

# ビルダーからバイナリをコピー
COPY --from=builder /app/main .

# ポートを公開
EXPOSE 8080

# アプリケーションを実行
CMD ["./main"]
