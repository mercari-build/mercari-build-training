# STEP3: 出品APIを作る

## 1. APIをたたいてみる

**:book: Reference**

* (JA) [Udemy -【基礎からわかる！】Webアプリケーションの仕組み](https://www.udemy.com/course/tobach_01_webapp_structure/)
* (JA) [HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (JA) [HTTP リクエストメソッド](https://developer.mozilla.org/ja/docs/Web/HTTP/Methods)
* (JA) [APIとは？意味やメリット、使い方を世界一わかりやすく解説](https://www.sejuku.net/blog/7087)

* (EN) [API and Web Service Introduction](https://www.udemy.com/course/api-and-web-service-introduction/)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
* (EN) [HTTP request methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)

### GETリクエスト

サービスをローカルで立ち上げると、ブラウザで `http://127.0.0.1:9000` からサービスにアクセスすることができるようになりました。
次に、curlというコマンドを使ってアクセスをしてみます。 curlがインストールされていない場合、インストールしてください。

```shell
curl -X GET 'http://127.0.0.1:9000'
```

ブラウザと同じように`{"message": "Hello, world!"}` がコンソール上で返ってくることを確認します。

### POSTリクエスト

サンプルコードには `/items` というエンドポイントが用意されています。 こちらのエンドポイントをcurlで叩いてみます。

```shell
$ curl -X POST 'http://127.0.0.1:9000/items'
```

このエンドポイントは、コールに成功すると`{"message": "item received: <name>"}`
というレスポンスが返ってくることが期待されていますが、違ったレスポンスが返ってきてしまいます。

コマンドを以下のように修正することで、`{"message": "item received: jacket"}`が返ってきますが、なぜそのような結果になるのか調べてみましょう。

```shell
$ curl -X POST \
  --url 'http://localhost:9000/items' \
  -d name=jacket
```

**:beginner: Point**

* POSTとGETのリクエストの違いについて調べてみましょう
* ブラウザで `http://127.0.0.1:9000/items` にアクセスしても `{"message": "item received: <name>"}`
  が返ってこないのはなぜでしょうか？
  * アクセスしたときに返ってくる**HTTPステータスコード**はいくつですか？
  * それはどんな意味をもつステータスコードですか？

## 2. 新しい商品を登録する

商品を登録するエンドポイントを作成します。

**:book: Reference**

* (JA)[RESTful Web API の設計](https://docs.microsoft.com/ja-jp/azure/architecture/best-practices/api-design)
* (JA)[HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (EN) [RESTful web API design](https://docs.microsoft.com/en-us/azure/architecture/best-practices/api-design)
* (EN) [HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)


準備されている`POST /items`のエンドポイントはnameという情報を受け取れます。 ここにcategoryの情報も受け取れるように変更を加えます。

* name: 商品の名前 (string)
* category: 商品のカテゴリ(string)

このままではデータの保存ができないので、jsonファイルに保存するようにしましょう。
`items.json` というファイルを作り、そこの`items`というキーに新しく登録された商品を追加するようにしましょう。

商品を追加すると、items.jsonの中身は以下のようになることを期待しています。
```json
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 3. 商品一覧を取得する

GETで`/items`にアクセスしたときに、登録された商品一覧を取得できるようにエンドポイントを実装しましょう。 以下のようなレスポンスを期待しています。

```shell
# 商品の登録
$ curl -X POST \
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

## 4. Databaseに保存する

ここまで`items.json`に情報を保存してきましたが、このデータをデータベースに移し替えます。  
今回は **SQLite**というデータベースを使います。

**:book: Reference**

* (JA)[SQLite入門](https://www.dbonline.jp/sqlite/)
* (JA)[Udemy -【SQLiteで学ぶ】ゼロから始めるデータベースとSQL超入門](https://www.udemy.com/course/basic_database_sqlite/)
* (JA)[Udemy - はじめてのSQLserver データベース　SQL未経験者〜初心者向けコース](https://www.udemy.com/course/sqlserver-for-beginner/)
* (EN)[https://www.sqlitetutorial.net/](https://www.sqlitetutorial.net/)
* (EN)[Udemy - Intro To SQLite Databases for Python Programming](https://www.udemy.com/course/using-sqlite3-databases-with-python/)

SQLiteをインストールし、dbフォルダに、`mercari.sqlite3` というデータベースファイルを作成します。
`mercari.sqlite3`を開き、`items`テーブルを作成します。

itemsテーブルは以下のように定義し、スキーマを `db/items.db` に保存します。

* id: int 商品ごとにユニークなID
* name: string 商品の名前
* category: string 商品のカテゴリ

データがデータベースに保存され、商品一覧情報を取り出すことができるように、`GET /items`と`POST /items`のエンドポイントを変更しましょう。

`items.db`はgitの管理対象にしますが、`mercari.sqlite3`はgitの管理対象として追加しないようにしてください。

**:beginner: Point**

* jsonファイルではなくデータベース(SQLite)にデータを保存する利点は何がありますか？

## 5. 商品を検索する

指定したキーワードを含む商品一覧を返す、`GET /search`エンドポイントを作ります。

```shell
# "jacket"という文字を含む商品一覧をリクエストする
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
# "jacket"をnameに含む商品一覧が返ってくる
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 6. 画像を登録する

商品情報に画像(image)を登録できるように、`GET /items`と`POST /items`のエンドポイントを変更します。

* 画像は `images` というフォルダを作成し保存します
* ポストされた画像のファイル名を sha256 で hash化し、`<hash>.jpg`という名前で保存します
* itemsテーブルに画像のファイル名をstringで保存できるように変更を加えます

```shell
# ローカルから.jpgをポストする
curl -X POST \
  --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
```


Items table example:

| id   | name   | category | image_filename                                                       |
| :--- | :----- | :------- | :------------------------------------------------------------------- |
| 1    | jacket | fashion  | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |          |                                                                      |

**:beginner: Point**

* Hash化とはなにか？
* sha256以外にどんなハッシュ関数があるか調べてみましょう

## 7. 商品の詳細を返す

商品の詳細情報を取得する  `GET /items/<item_id>` というエンドポイントを作成します。

```shell
$ curl -X GET 'http://127.0.0.1:9000/items/1'
{"name": "jacket", "category": "fashion", "image_filename": "..."}
```

## 8. (Optional) カテゴリの情報を別のテーブルに移す

データベースを以下のように構成しなおします。
これによってカテゴリの名前が？途中で変わったとしても、全部のitemsテーブルのcategoryを修正する必要がなくなります。  
`GET items`ではcategoryの名前を変わらず取得したいので、テーブルをjoinしてレスポンス用のデータを作って返しましょう。

**items table**

| id   | name   | category_id | image_filename                                                       |
| :--- | :----- | :---------- | :------------------------------------------------------------------- |
| 1    | jacket | 1           | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |             |                                                                      |

**category table**

| id   | name    |
| :--- | :------ |
| 1    | fashion |
| ...  |         |

**:beginner: Point**
* データベースの**正規化**とは何でしょうか？


## 9. (Optional) Loggerについて調べる
`http://127.0.0.1:9000/image/no_image.jpg`にアクセスしてみましょう。
`no image`という画像が帰ってきますが、 コード中にある
```
Image not found: <image path>
```
というデバッグログがコンソールに表示されません。
これはなぜか、調べてみましょう。
これを表示するためには、どこを変更したらいいでしょうか？

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

[STEP4: 仮想環境でアプリを動かす](step4.ja.md)