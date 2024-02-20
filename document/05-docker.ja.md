# STEP5: 仮想環境でアプリを動かす

このステップでは docker の使い方を学びます。

**:book: Reference**

* (JA)[docker docs](https://matsuand.github.io/docs.docker.jp.onthefly/get-started/overview/)
* (JA)[Udemy Business - ゼロからはじめる Dockerによるアプリケーション実行環境構築](https://mercari.udemy.com/course/docker-k/)

* (EN)[docker docs](https://docs.docker.com/get-started/overview/)
* (EN)[Udemy Business - Docker for the Absolute Beginner - Hands On - DevOps](https://mercari.udemy.com/course/learn-docker/)
## 1. Docker をインストールする
**最新のdockerをインストールし、`docker -v` が実行できることを確認しましょう。**

**:book: Reference**

* [Docker のダウンロードとインストール](https://matsuand.github.io/docs.docker.jp.onthefly/get-started/#download-and-install-docker)
## 2. Docker を触ってみる
**`mercari-build-training/` 以下にいることを確認して次のコマンドを実行してみましょう。**

```shell
$ docker run -v $(pwd)/data/text_en.png:/tmp/img.png wakanapo/tesseract-ocr tesseract /tmp/img.png stdout -l eng
```

メッセージが表示されましたか？

このコマンドを実行すると[レジストリ](https://hub.docker.com/repository/docker/wakanapo/tesseract-ocr)にある docker image がローカルにダウンロードされ実行されます。

この docker image には画像から文字を読み取る機能 (OCR) が実装されています。
このように docker image を使うと手元の環境を変更することなく、docker image 内に構築された環境を使ってシステムを実行することができます。

ちなみに、以下のコマンドで日本語も読み取ることができます。

```shell
$ docker run -v $(pwd)/data/text_ja.png:/tmp/img.png wakanapo/tesseract-ocr tesseract /tmp/img.png stdout -l jpn
```

**英語か日本語の文字が含まれる好きな画像を用意して、文字が読み取れるか試してみましょう。**

**:beginner: Point**

* [Dockerのvolume](https://matsuand.github.io/docs.docker.jp.onthefly/storage/volumes/) について理解しましょう

## 3. Docker Image を取得する

**次のコマンドを実行してみましょう。**
```shell
$ docker images
```

これはローカルのホスト上にあるイメージの一覧を表示するコマンドです。
先程使った `wakanapo/tesseract-ocr` という image があることが確認できるはずです。

**次のコマンドを実行し、docker にはどんなコマンドあるか確認しましょう。**
```
$ docker help
```

docker はホスト上に存在しないイメージを使う際には、自動的に image をダウンロードしてくれます。しかしながら、image を予めダウンロードしておくこともできます。

**レジストリから image を取得するコマンドを調べ、`alpine`　という名前の image を取得してみましょう。**

イメージ一覧の中に `alpine` という image があることを確認しましょう。

**:book: Reference**

* [Docker コマンド](https://docs.docker.jp/engine/reference/commandline/index.html)

**:beginner: Point**

どのようなときに使われるコマンドか理解できているか確認しましょう。

* images
* help
* pull


## 4. Docker Image を Build する
**pythonで開発をしている人は`python/`, Goの人は`go/`以下にある`Dockerfile`をbuildしてみましょう。**

* 名前（リポジトリ名）は `build2024/app`, タグは`latest` とします。

イメージ一覧の中に `build2024/app` という image があれば成功です。


**:book: Reference**

* [Dockerfile リファレンス](https://docs.docker.jp/engine/reference/builder.html)

## 5. Dockerfile を 変更する
**STEP4-4 で Build した Image を実行し、次のようなerrorが出ることを確認しましょう。**

```
docker: Error response from daemon: OCI runtime create failed: container_linux.go:380: starting container process caused: exec: "python": executable file not found in $PATH: unknown.
ERRO[0000] error waiting for container: context canceled 
```
Goの場合は、上のエラーメッセージの`"python"`の部分が`"go"`になります。


**`Dockerfile` を変更し STEP2 でインストールしたのと同じバージョンの Python や Go が docker image で使えるようにしましょう。**

変更した `Dockerfile` で build した Image を実行し、STEP2-2 と同じ結果が表示されれば成功です。

**:book: Reference**

* [docker docs - 言語別ガイド (Python)](https://matsuand.github.io/docs.docker.jp.onthefly/language/python/)
* [docker docs - 言語別ガイド (Go)](https://matsuand.github.io/docs.docker.jp.onthefly/language/golang/)

## 6. 出品 API を docker 上で動かす

STEP4-5 までで docker image の中は STEP2-2 と同じ状態になっています。

**`Dockerfile`を変更し、必要なファイルをコピーしたり依存ライブラリをインストールしたりして, docker image 上で 出品 API が動くようにしましょう。**

`$ docker run -d -p 9000:9000 build2024/app:latest`

を実行しSTEP3と同様にしてAPIを叩ければ成功です。

---
**:beginner: Point**

以下のキーワードについて理解できているか確認しましょう。

* images
* pull
* build
* run
* Dockerfile

---

### Next

[STEP5: Webのフロントエンドを実装する](07-frontend.ja.md)