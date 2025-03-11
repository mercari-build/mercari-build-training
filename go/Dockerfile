# ベースイメージを最新の Go 1.24 に変更
FROM golang:1.24-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache gcc musl-dev sqlite sqlite-dev

# 作業ディレクトリを設定
WORKDIR /app

# go.mod と go.sum をコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod tidy

# ソースコードをコピー
COPY . .

# 作業ディレクトリを API のルートに変更
WORKDIR /app/cmd/api

# アプリをビルド
RUN go build -o /app/server .

# 軽量なランタイム用のイメージを作成
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk add --no-cache sqlite sqlite-dev

# アプリケーション用のユーザーを作成
RUN addgroup -S mercari && adduser -S trainee -G mercari

# 作業ディレクトリを設定
WORKDIR /app

# 画像保存ディレクトリとDBを作成
RUN mkdir -p images db && chown -R trainee:mercari /app

# ユーザーを変更
USER trainee

# バイナリをコピー
COPY --from=builder /app/server .

# ポートを公開
EXPOSE 9000

# アプリケーションを実行
CMD ["./server"]