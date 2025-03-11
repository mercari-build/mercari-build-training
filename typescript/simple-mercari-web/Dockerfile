# Node.js 20の軽量 Alpine イメージを使用
FROM node:20-alpine

# ユーザー作成
RUN addgroup -S mercari && adduser -S trainee -G mercari

# 作業ディレクトリを設定
WORKDIR /app

# package.json と package-lock.json をコピー
COPY package.json package-lock.json ./

# 依存関係をインストール
RUN npm ci

# ソースコードをコピー
COPY . .

# ファイルの権限を修正
RUN chown -R trainee:mercari /app

# ポート3000を開放
EXPOSE 3000

# 非 root ユーザーで実行
USER trainee

# アプリケーションを起動
CMD ["npm", "start", "--", "--host", "0.0.0.0"]
