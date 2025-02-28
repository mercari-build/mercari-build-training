# STEP3: 出品APIを作る

## 1. APIを呼び出す

**:book: Reference**

* (JA) [Udemy Business - REST WebAPI サービス 設計](https://mercari.udemy.com/course/rest-webapi-development/)
* (JA) [HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (JA) [HTTP リクエストメソッド](https://developer.mozilla.org/ja/docs/Web/HTTP/Methods)
* (JA) [APIとは？意味やメリット、使い方を世界一わかりやすく解説](https://www.sejuku.net/blog/7087)

* (EN) [Udemy Business - API and Web Service Introduction](https://mercari.udemy.com/course/api-and-web-service-introduction/)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
* (EN) [HTTP request methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)

本節のゴールは、ツールを用いてAPIを呼び出すことです。

### API呼び出しに利用できるツールについて
APIの呼び出しはブラウザからも可能ですが、自由にリクエストを送るためにはコマンドラインツールを使うのが便利です。ツールとしては、GUIの[Insomnia](https://insomnia.rest/)や[Postman](https://www.postman.com/)、CUIの[HTTPie](https://github.com/httpie/cli)やcURLなどが存在しています。今回は、よく利用されるcURLを利用してみましょう。

### cURLのインストール
cURLが利用されているかは、以下のコマンドで確認できます。

```shell
$ curl --version
```

このコマンドを実行後にバージョンが表示されればcURLはインストールされています。インストールされていない場合は、各自調べてインストールしてください。

### GETリクエストの送信

cURLを用いて、前節で立ち上げたAPIサーバに対してGETリクエストを送ってみましょう。

cURLでリクエストを送る前に、HTTPブラウザで `http://127.0.0.1:9000` にアクセスしたときに、 `{"message": "Hello, world!"}` が表示されることを確認してください。仮に表示されない場合は、前章の4節: アプリにアクセスするを参照してください([Python](./02-local-env.ja.md#4-アプリにアクセスする), [Go](./02-local-env.ja.md#4-アプリにアクセスする-1))。

さて、実際にcURLコマンドを用いてリクエストを送ってみましょう。ここではGETリクエストとPOSTリクエストを送信します。

新しいターミナルを開き、以下のコマンドを実行してください。

```shell
$ curl -X GET 'http://127.0.0.1:9000'
```

ブラウザと同じように`{"message": "Hello, world!"}` がコンソール上で返ってくることを確認してください。

### POSTリクエストの送信と修正

次に、POSTリクエストを送ってみましょう。サンプルコードには `/items` というエンドポイントが用意されているので、こちらのエンドポイントに対してcURLでリクエストを送ります。以下のコマンドを実行してください。

```shell
$ curl -X POST 'http://127.0.0.1:9000/items'
```

このエンドポイントは、コールに成功すると`{"message": "item received: <name>"}`
というレスポンスが返ってくることが期待されています。しかし、ここでは異なるレスポンスが返ってくるはずです。

コマンドを以下のように修正することで、`{"message": "item received: jacket"}`が返ってきますが、なぜそのような結果になるのか調べてみましょう。

```shell
$ curl \
  -X POST \
  --url 'http://localhost:9000/items' \
  -d 'name=jacket'
```

**:beginner: Point**

* GETとPOSTのリクエストの違いについて調べてみましょう
* ブラウザで `http://127.0.0.1:9000/items` にアクセスしても `{"message": "item received: <name>"}` が返ってこないのはなぜでしょうか？
  * アクセスしたときに返ってくる**HTTPステータスコード**はいくつですか？
  * それはどんな意味をもつステータスコードですか？

## 2. 新しい商品を登録する

**:book: Reference**

* (JA)[RESTful Web API の設計](https://docs.microsoft.com/ja-jp/azure/architecture/best-practices/api-design)
* (JA)[HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (EN) [RESTful web API design](https://docs.microsoft.com/en-us/azure/architecture/best-practices/api-design)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)

本節のゴールは、 `POST /items` のエンドポイントの拡張と `items` に関連するデータの永続化です。

前提として、準備されている`POST /items`のエンドポイントは `name` という情報を受け取ることが出来ます。 ここで、`category` という情報も受け取れるように変更を加えましょう。

* `name`: 商品の名前 (string型)
* `category`: 商品のカテゴリ(string型)

このままではデータの保存ができないので、JSONファイルに保存するようにしましょう。
`items.json` というファイルを作り、ファイル内で保持されるJSONでは`items`というキーに新しく登録された商品を追加するようにしましょう。

商品を追加すると、`items.json` の中身は以下のようになることを期待しています。

```json
{
  "items": [
    {
      "name": "jacket",
      "category": "fashion"
    },
    ... (ここから別のアイテムが続く)
  ]
}
```

## 3. 商品一覧を取得する

本節のゴールは、登録された商品一覧を取得するための `GET /items` エンドポイントを実装することです。

GETで`/items`にアクセスしたときに、以下のようなレスポンスを期待しています。

```shell
# 商品の登録
$ curl \
  -X POST \
  --url 'http://localhost:9000/items' \
  -d 'name=jacket' \
  -d 'category=fashion'
# /itemsにPOSTリクエストを送った時のレスポンス
{"message": "item received: jacket"}
# 登録された商品一覧
$ curl -X GET 'http://127.0.0.1:9000/items'
# /itemsにGETリクエストを送った時のレスポンス
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```


## 4. 画像を登録する

本節のゴールは、商品画像に画像(image)を登録できるようにすることです。そのために、`GET /items`と`POST /items`のエンドポイントを変更しましょう。

* `images` というディレクトリを作成し、画像はそのディレクトリ以下に保存してください
* 送信された画像のファイルを SHA-256 でハッシュ化し、`<hashed-value>.jpg`という名前で保存します
* itemsに画像のファイル名をstringで保存できるように変更を加えます

```shell
# ローカルから.jpgをポストする
$ curl \
  -X POST \
  --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
```


```json
{"items": [{"name": "jacket", "category": "fashion", "image_name": "510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"}, ...]}
```


**:beginner: Point**

* ハッシュ化とはなにか？
* SHA-256 以外にどんなハッシュ関数があるか調べてみましょう

## 5. 商品の詳細を返す

本節のゴールは、1商品の詳細情報を取得できるエンドポイントを作成することです。
そのために、 `GET /items/<item_id>` というエンドポイントを作成します。

```shell
$ curl -X GET 'http://127.0.0.1:9000/items/1'
{"name": "jacket", "category": "fashion", "image_name": "..."}
```

## 6. (Optional) Loggerについて調べる
`http://127.0.0.1:9000/image/no_image.jpg`にアクセスしてみましょう。
`no image`という画像が帰ってきますが、 コード中にある
```
Image not found: <image path>
```
というデバッグログがコンソールに表示されません。
これはなぜか、調べてみましょう。
これを表示するためには、コードのどこを変更したらいいでしょうか？

**:beginner: Point**
* Log levelとは？
* webサーバーでは、本番はどのログレベルまで表示する？

---
**:beginner: Point**

以下のキーワードについて理解しましょう

* port (ポート番号)
* localhost, 127.0.0.1
* HTTPリクエストメソッド (GET, POST...)
* HTTPステータスコード (1XX, 2XX, 3XX, 4XX, 5XXはそれぞれどんな意味を持ちますか？)

---

### Next

[STEP4: データベース](04-database.ja.md)
