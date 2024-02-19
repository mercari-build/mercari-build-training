# ベースイメージとしてPythonの公式イメージを使用
FROM python:3.8-slim

# 必要なパッケージをインストール
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        gcc \
        libc-dev \
    && rm -rf /var/lib/apt/lists/*

# 必要なPythonパッケージをインストール
COPY python/requirements.txt requirements.txt
RUN pip install -r requirements.txt

# アプリケーションのディレクトリを作成
WORKDIR /app

# dbディレクトリ内のファイルをコピー
COPY db/ /app/db

# Pythonファイルをコピー
COPY python/ /app/python

# ポート9000番を公開
EXPOSE 9000

# 環境変数を設定
ENV DB_PATH=/app/db/main.sqlite3
ENV CATEGORY_DB_PATH=/app/db/category.sqlite3

# FastAPIアプリケーションを実行
CMD ["uvicorn", "python.main:app", "--host", "0.0.0.0", "--port", "9000"]

