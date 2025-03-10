メモと参考リンクまとめ
<!-- - [infra.go](#infra.go)
- [server.go](#server.go) -->

## infra.go

永続化のための処理が責務
-> データを残しておくために、保存や読み込みなどの処理を担当

### NewItemRepository()

* items tableとcategories tableを作成
* 返り値: `ItemRepository` (itemRepositoryインターフェース), `error` (エラーが発生した場合)
* データベースに接続、テーブルが存在しない場合は作成する。エラーが発生した場合は、nilのItemRepositoryとエラーを返す

### Insert()

* アイテムをデータベースに挿入する処理
* 引数: `ctx context.Context` (コンテキスト), `item *Item` (挿入するアイテム)
* 返り値: `error` (エラーが発生した場合)
* カテゴリが存在しない場合は新規作成、アイテムをデータベースに挿入
* imageの引数がなかったらdefault.jpgを使用

### GetAll()

* 全てのアイテムをデータベースから取得する処理
* 引数: `ctx context.Context` (コンテキスト)
* 返り値: `[]Item` (アイテムのスライス), `error` (エラーが発生した場合)
* itemsテーブルとcategoriesテーブルをINNER JOINして全てのアイテムを取得

### GetItemById()

* IDから特定のアイテムを取得する処理
* 引数: `ctx context.Context` (コンテキスト), `item_id string` (取得するアイテムのID)
* 返り値: `Item` (アイテム), `error` (エラーが発生した場合)
* 指定されたIDのアイテムをデータベースから取得、アイテムが存在しない場合は `errItemNotFound` エラーを返す

### SearchItemsByKeyword()

* キーワードからアイテムを検索する処理
* 引数: `ctx context.Context` (コンテキスト), `keyword string` (検索キーワード)
* 返り値: `[]Item` (アイテムのスライス), `error` (エラーが発生した場合)
* 指定されたキーワードを名前に含むアイテムをデータベースから検索する。

### StoreImage()

* 画像を保存する処理
* 引数: `fileName string` (ファイル名), `image []byte` (画像データ)
* 返り値: `error` (エラーが発生した場合)
* 指定されたファイル名(SHA-256)で画像を保存



## server.go

HTTPリクエスト/レスポンス等のハンドリング、ハンドラのロジック管理が責務
-> HTTPリクエストを受け取り、それに対する応答HTTPレスポンスを生成するまでの一連の処理を担当

### Run()

* HTTPサーバーを起動し、リクエストの受付を開始する処理
* 引数: `s Server` (サーバー設定)
* 返り値: `int` (サーバーの起動結果。0は成功、1は失敗)

### Hello()

* GET `/` に対するハンドラ。Hello, world! メッセージを返す。
* 引数: `w http.ResponseWriter` (HTTPレスポンスライター), `r *http.Request` (HTTPリクエスト)
* 返り値: `なし`

### GetItems()

* GET `/items` に対するハンドラ。全てのアイテムをJSON形式で返す。
* 引数: `w http.ResponseWriter` (HTTPレスポンスライター), `r *http.Request` (HTTPリクエスト)
* 返り値: `なし`

### AddItem()

* POST `/items` に対するハンドラ。新しいアイテムをリクエストから受け取り、データベースに挿入する。
* 引数: `w http.ResponseWriter` (HTTPレスポンスライター), `r *http.Request` (HTTPリクエスト)
* 返り値: `なし`

### GetImage()

* GET `/images/{filename}` に対するハンドラ。指定された画像を返す。画像が見つからない場合はデフォルト画像を返す。
* 引数: `w http.ResponseWriter` (HTTPレスポンスライター), `r *http.Request` (HTTPリクエスト)
* 返り値: `なし`

### GetItemById()

* GET `/items/{item_id}` に対するハンドラ。IDから特定のアイテムをJSON形式で返す。
* 引数: `w http.ResponseWriter` (HTTPレスポンスライター), `r *http.Request` (HTTPリクエスト)
* 返り値: `なし`

### SearchItemsByKeyword()

* GET `/search?keyword={keyword}` に対するハンドラ。キーワードからアイテムを検索し、JSON形式で返す。
* 引数: `w http.ResponseWriter` (HTTPレスポンスライター), `r *http.Request` (HTTPリクエスト)
* 返り値: `なし`

### storeImage()

* 画像を受けとってハッシュ値をファイル名として保存する。
* 引数: `image []byte` (画像データ)
* 返り値: `(string, error)` (保存された画像のパス, エラーが発生した場合)

### parseAddItemRequest()

* POST `/items` のリクエストを解析し、バリデーションを行う。
* 引数: `r *http.Request` (HTTPリクエスト)
* 返り値: `(*AddItemRequest, error)` (解析されたリクエスト, エラーが発生した場合)

### parseGetImageRequest()

* GET `/images/{filename}` のリクエストを解析し、バリデーションを行う。
* 引数: `r *http.Request` (HTTPリクエスト)
* 返り値: `(*GetImageRequest, error)` (解析されたリクエスト, エラーが発生した場合)

### buildImagePath()

* 画像のファイルパスを構築し、バリデーションを行う。
* 引数: `imageFileName string` (画像ファイル名)
* 返り値: `(string, error)` (構築された画像パス, エラーが発生した場合)

### parseGetItemByIdRequest()

* GET `/items/{item_id}` のリクエストを解析し、バリデーションを行う。
* 引数: `r *http.Request` (HTTPリクエスト)
* 返り値: `(*GetItemByIdRequest, error)` (解析されたリクエスト, エラーが発生した場合)

### parseGetItemByKeywordRequest()

* GET `/search?keyword={keyword}` のリクエストを解析し、バリデーションを行う。
* 引数: `r *http.Request` (HTTPリクエスト)
* 返り値: `(*GetItemByKeywordRequest, error)` (解析されたリクエスト, エラーが発生した場合)



## 🔰
### STEP4: 出品APIを作る

**❓GETとPOSTの違いについて調べてみましょう**
* GET: サーバーにリクエストを送信、リソースを取得
* POST: サーバーにデータを送信、リソースの更新など

**❓ブラウザで `http://127.0.0.1:9000/items` にアクセスしても `{"message": "item received: <name>"}` が返ってこないのはなぜでしょうか？**
* server.go の route に GET /items がないから？

**❓アクセスしたときに返ってくるHTTPステータスコードはいくつですか？**
* 200 OK

**❓それはどんな意味をもつステータスコードですか？**
* リクエストが正常に処理された

**❓ハッシュ化とはなにか？**
* 特定のルール(ハッシュ関数)に基づいて値を変換すること

**❓SHA-256 以外にどんなハッシュ関数があるか調べてみましょう**
* SHA-3, MD5など
    * アルゴリズムの設計、セキュリティ強度、速度、用途が違う らしい

**❓Log levelとは？**
* ソフトウェアが記録するログ(どんな動作が行われたかの記録)の詳細度と重要度を調整するための仕組み

**❓webサーバーでは、本番はどのログレベルまで表示する？**
* INFO以上が一般的 開発環境だとDEBUG

**❓port (ポート番号)**
* コンピュータが通信に使用するプログラムを識別するための番号 HTTP:80 etc.

**❓localhost, 127.0.0.1**
* localhost: コンピューター自身を指し示すためのホスト名
* 127.0.0.1: IPv4における特別なIPアドレス

**❓HTTPリクエストメソッド**
* Webサーバーにどのような処理をするかを伝える役割
* GET/POST/PUT(更新)/PATCH(一部更新)/DELETE(削除)

**❓HTTPステータスコード (1XX, 2XX, 3XX, 4XX, 5XXはそれぞれどんな意味を持ちますか？)**
* 1XX: リクエストが受け付けられて処理が続いている(Informational)
* 2XX: リクエストが正常に完了(Success)
* 3XX: リクエストを完了するために追加のアクションが必要(Redirection)
* 4XX: リクエストに問題あり(Client Error)
* 5XX: サーバーがリクエストを処理できなかった(Server Error)

### STEP5: データベース

**❓jsonファイルではなくデータベース(SQLite)にデータを保存する利点は何がありますか？**
* dbだとデータの整合性がとりやすい、データ操作・検索が効率的(jsonだとファイル全体を読む込む必要がある)

**❓データベースの正規化とは何でしょうか？**
* データの重複を排除し、データの整合性を保つためのプロセス
* これは第二正規形?

### STEP6: テストを用いてAPIの挙動を確認する

**❓このテストは何を検証しているでしょうか？**
* AddItemRequestへのリクエストが期待する形式かを確認

**❓`t.Error()` と `t.Fatal()` には、どのような違いがあるでしょうか？**
* t.Error() -> テスト失敗を記録、テスト関数の実行は継続
* t.Fatal() -> テスト失敗を記録、テスト関数の実行を終了(中断を示す)

**❓モックを満たすためにinterfaceを用いていますが、interfaceのメリットについて考えてみましょう**
* モックはインターフェースを実装した代替オブジェクト -> 実際の依存関係を模倣できる

**❓モックを利用するメリットとデメリットについて考えてみましょう**
* 単体テストが簡単


## なるほど
`go run cmd/api/main.go` でサーバーを起動するならmain.go の実行ディレクトリは `go/`

そもそもハンドラーとは -> HTTPリクエストを受け取り、適切なレスポンスを返す関数

    step4-4
    	# ローカルから.jpgをポストする
    	$ curl \
    	-X POST \
    	--url 'http://localhost:9000/items' \
    	-F 'name=jacket' \
    	-F 'category=fashion' \
    	-F 'image=@images/local_image.jpg' <-ローカルのuploadしたいiamgeのパス "image=go/images/default.jpg"とか
 
 r.PathValue と r.URL.Query().Get
	-> http:/@/127.0.0.1:9000/Path?<query parameter>
	-> r.PathValue: Pathを取得　/items/{item_id}だと{item_id}を取得する
	-> r.URL.Query().Get: クエリパラメータを取得　/search?keyword=jacketだとjacketを取得


## Link
🔗[GoでAPIから取得したJSONを5分でパースする - ぺい](https://tikasan.hatenablog.com/entry/2017/04/26/110854)  
🔗[go言語 ファイルの拡張子のみを取得する | mebee](https://mebee.info/2021/05/28/post-23288/)  
🔗[今goのエラーハンドリングを無難にしておく方法（2021.09現在）](https://zenn.dev/nekoshita/articles/097e00c6d3d1c9)  
🔗[Go言語でSQLite3を使う](https://zenn.dev/teasy/articles/go-sqlite3-sample)  
🔗[【Go言語】database/sqlパッケージによるデータベース操作入門 - sqlite3 - Ike Tech Blog](https://iketechblog.com/database-sql-go-sqlite3/) 
🔗[Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work.](https://github.com/mattn/go-sqlite3/issues/855)    
🔗[os package - os - Go Packages](https://pkg.go.dev/os#pkg-variables)  
🔗[[入門]GoでSQLite3を使いデータベース操作を行ってみる](https://zenn.dev/tara_is_ok/articles/15b04694466bec)  
🔗[http package - net/http - Go Packages](https://pkg.go.dev/net/http)  
🔗[go-sqlite3/_example/simple/simple.go at master · mattn/go-sqlite3](https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go)  
🔗[Go database/sql の操作ガイドあったんかい](https://sourjp.github.io/posts/go-db/)  
🔗[mockgenが2023年６月２８日で読み取り専用になった](https://zenn.dev/135yshr/articles/6fa5ccc644ba29)    
🔗[GolangでDBアクセスがあるユニットテストのやり方を考える](https://qiita.com/seya/items/582c2bdcca4ad50b03b7)   



