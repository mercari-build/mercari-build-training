# STEP4: データベース

ここまで`items.json`に情報を保存してきましたが、このデータをデータベースに移し替えます。

**:book: Reference**

* (JA)[SQLite入門](https://www.dbonline.jp/sqlite/)
* (JA)[Udemy Business - データベース講座１：データベース論理設計](https://mercari.udemy.com/course/database-logic/)

* (EN)[https://www.sqlitetutorial.net/](https://www.sqlitetutorial.net/)
* (EN)[Udemy Business - SQLite for beginners](https://mercari.udemy.com/course/sqlite-for-beginners/)
* (EN)[Udemy Business - Relational Database Designs](https://mercari.udemy.com/course/relational-database-design/)

## 1. SQLiteに情報を移行する
今回は **SQLite**というデータベースを使います。

* SQLiteをインストール
* dbフォルダに、`mercari.sqlite3` というデータベースファイルを作成
* `mercari.sqlite3`を開き、`items`テーブルを作成 
*  `items`テーブルは以下のように定義し、スキーマを `db/items.db` に保存します。
  * id: int 商品ごとにユニークなID
  * name: string 商品の名前
  * category: string 商品のカテゴリ
  * image_name: string 画像のパス

`items.db`はgitの管理対象にしますが、`mercari.sqlite3`はgitの管理対象として追加しないようにしてください。

データがデータベースに保存され、商品一覧情報を取り出すことができるように、`GET /items`と`POST /items`のエンドポイントを変更しましょう。


Items table example:

| id   | name   | category | image_name                                                           |
| :--- | :----- | :------- |:---------------------------------------------------------------------|
| 1    | jacket | fashion  | 510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg |
| 2    | ...    |          |                                                                      |


**:beginner: Point**

* jsonファイルではなくデータベース(SQLite)にデータを保存する利点は何がありますか？

## 2. 商品を検索する

指定したキーワードを含む商品一覧を返す、`GET /search`エンドポイントを作ります。

```shell
# "jacket"という文字を含む商品一覧をリクエストする
$ curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
# "jacket"をnameに含む商品一覧が返ってくる
{"items": [{"name": "jacket", "category": "fashion"}, ...]}
```

## 3. カテゴリの情報を別のテーブルに移す

データベースを以下のように構成しなおします。  
これによってカテゴリの名前を途中で変えたとしても、全部のitemsテーブルのcategoryを修正する必要がなくなります。  
`GET items`ではこれまでと同様にcategoryの名前を取得したいので、テーブルを**join**してレスポンス用のデータを作って返すように実装を更新しましょう。

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

---

### Next

[STEP5: 仮想環境でアプリを動かす](05-docker.ja.md)
