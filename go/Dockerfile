FROM golang:1.24.0-alpine3.20

# github.com/mattn/go-sqlite3 を使う場合、CGO_ENABLED=1 が必要
ENV CGO_ENABLED=1

# gcc, musl-dev, sqlite をインストール
# go-sqlite3 は C言語ベースのライブラリ
RUN apk add --no-cache gcc musl-dev sqlite

WORKDIR /app

# go.mod, go.sum をコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# アプリのソースコードをコピー
COPY . .

# ユーザー・グループの作成と権限設定
RUN addgroup -S mercari && adduser -S trainee -G mercari \
    && chown -R trainee:mercari db images

# バイナリをビルド
RUN go build -o server cmd/api/main.go

USER trainee

# `go run` ではなく `./server` を実行
CMD ["./server"]