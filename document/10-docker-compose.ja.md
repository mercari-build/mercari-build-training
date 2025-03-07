# STEP10: docker-composeを利用して複数のサービスを動かす

このステップでは docker-compose の使い方を学びます。

**:book: Reference**

* (JA)[Docker Compose の概要](https://matsuand.github.io/docs.docker.jp.onthefly/compose/)
* (JA)[Udemy Business - 駆け出しエンジニアのためのDocker入門](https://mercari.udemy.com/course/docker-startup/)

* (EN)[Docker Compose Overview](https://docs.docker.com/compose/)
* (EN)[Udemy Business - Docker for the Absolute Beginner - Hands On - DevOps](https://mercari.udemy.com/course/learn-docker/)
## 1. (復習) フロントエンドの docker image を作成する

**STEP7を思い出しながらフロントエンドの docker image を作成しましょう。**

`typescript/simple-mercari-web`以下にフロントエンド用の `Dockerfile` がすでに用意されています。これを変更しフロントエンドが docker 上で立ち上がるようにしましょう。

* 名前（リポジトリ名）は `mercari-build-training/web`, タグは`latest` とします。

`$ docker run -d -p 3000:3000 mercari-build-training/web:latest`

を実行し、ブラウザから[http://localhost:3000/](http://localhost:3000/)が正しく開ければ成功です。

## 2. Docker Compose をインストールする
**Docker Composeをインストールし、`docker-compose -v` が実行できることを確認しましょう**

**:book: Reference**

* [Docker Compose のインストール](https://matsuand.github.io/docs.docker.jp.onthefly/compose/install/)

## 3. Docker Compose のチュートリアルをやってみる
**[Docker Compose のチュートリアル](https://matsuand.github.io/docs.docker.jp.onthefly/compose/gettingstarted/)を一通りやってみましょう。**

:pushpin: チュートリアルではサンプルが Python で書かれていますが、Pythonの理解や環境は必須ではありません。これまでGoで開発していた人もこのチュートリアルに則って進めてください。

**:beginner: Point**

以下の質問に答えられるか確認しましょう。

* チュートリアルのdocker-composeファイルにはいくつのサービスが定義されていますか？それらはどのようなサービスですか？
* webサービスとredisサービスは異なる方法で image を取得しています。`docker-compose up`を実行した際に、各imageはどこから取得されているか確認しましょう。
* docker-composeでは、サービスから他のサービスのコンテナに接続することができます。webサービスは、redisサービスとどのように名前解決をし、接続していますか？

## 4. Docker ComposeでAPIとフロントエンドを動かす
**チュートリアルを参考にしながら、今回作成したサービスのフロントエンドとバックエンドのAPIをDocker Composeで動かせるようにしましょう**

`docker-compose.yml` は `mercari-build-training/` 以下に作成することにします。

以下の点を参考にしながら `docker-compose.yml` を作成しましょう。

* 使用する docker image
    * (Option 1: 難易度 ☆) STEP7 と STEP10-1 でそれぞれ build した `mercari-build-training/app:latest` と `mercari-build-training/web:latest` を使う
    * (Option 2: 難易度 ☆☆☆) `{go|python}/Dockerfile` と `typescript/simple-mercari-web/Dockerfile` から build するようにする
* 使用する port
    * API : 9000
    * フロントエンド : 3000
* サービス間の接続
    * フロントエンドは`REACT_APP_API_URL`という環境変数で設定されたURLのAPIにリクエストを送ります
    * APIはフロントエンドにリクエストは送りませんが[CORS](https://developer.mozilla.org/ja/docs/Web/HTTP/CORS)という仕組みのために、どこからリクエストが来るのか知っておく必要があります
    `FRONT_URL`という環境変数でフロントエンドのURLを指定しています

`docker-compose up` でサービスを起動して以下のことができれば成功です。
- [http://localhost:3000/](http://localhost:3000/)でページが正しく表示される
- 新しい商品の登録 (Listing)
- 商品の一覧の閲覧 (ItemList)
