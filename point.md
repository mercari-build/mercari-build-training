## Step1 : Git

* PRは個人のレポへのものになってるか確認
* LGTM (Looks Good To Me)

以下のキーワードについて理解できているか確認しましょう。

- [x] branch
- [x] commit
- [x] add
- [x] pull, push
- [x] Pull Request

## Step2 : 環境構築

* .gitignoreを設定

### 「コマンドにパスを通す」とは
    - コマンドの実行フルパスを、実行フルパス検索用のリストに追加する
    - コマンドは実行ファイル名を指定している　

     ex) パスが通っている場合 `ls`
     ex) パスが通っていない場合 `/bin/ls`

* 実行ファイル　＝ コマンドの実体となるファイル
* 検索パス　= 実行ファイルがあるか探しにいくパス
* 環境変数PATH = システム全体で使う変数

[PATHを通すとは？- 初心者でも分かる解説](https://hara-chan.com/it/programming/environment-variable-path/)


### go.mod 

* `go.mod`はGoモジュールのパスを書いておくファイル
* `go.sum`は依存モジュールのチェックサム(あるファイルやデータの一意性を示すハッシュ)の記録

### 使いそうなGoコマンド
*  `go get [import path]` : モジュールのダウンロード及びgo.mod / go.sumの修正
* `go install [import path]` : インターネット上のツールを$GOPATH/binにインストールする機能　
* `go mod tidy` : 
    - コード内でimportしているがgo getされていないモジュールをダウンロード
    - ダウンロードされているがコード内でimportされていないモジュールを削除
    - 上記2つを実施したあとにgo.modとgo.sumを修正 または 削除


[GO 公式ドキュメント](https://pkg.go.dev/cmd/go#hdr-The_go_mod_file)


## Step3: API

1. API
### POSTとGETのリクエストの違い
| メソッド | 意味                                               | 
| -------- | -------------------------------------------------- | 
| GET      | 指定したターゲットをサーバから取り出す             | 
| POST     | 指定したターゲット（プログラム）にデータを送る     | 
| HEAD     | 指定したターゲットに関連するヘッダー情報を取り出す | 
| PUT      | サーバ内のファイルを書き込む                       | 



* `http://127.0.0.1:9000/items`にアクセスしてみると
    > "status":405,"error":"code=405, message=Method Not Allowed"

    が返ってくる

* 405 : 「送信するクライアント側のメソッドが許可されていない」
    許可されていないメソッドでアクセスをした場合に出現するエラー

[他のステータスコード](https://www.itmanage.co.jp/column/http-www-request-response-statuscode/)


2. エンドポイントの作成
### Restful API
[参考](https://docs.microsoft.com/ja-jp/azure/architecture/best-practices/api-design)

* 目的: 
    1. プラットフォームの独立
    2. サービスの進化

*  Representational State Transfer (REST)とは
    - ハイパーメディアに基づき分散システムを構築するアーキテクチャ スタイル
    - 最も一般的な REST API 実装では、アプリケーション プロトコルとして HTTP を使用

### HTTP を使用した RESTful API の主な設計原則
* REST API は "リソース" を中心に設計

* リソースには "識別子" がある(URL)

* リソースの "表現" を交換することでサービスと対話。多くのWeb APIでは、交換形式としてJSONを使用

* クライアントとサービスの実装の分離に役立つ統一インターフェイスを使用

* REST API は、表現に含まれているハイパーメディア リンクによって動作



> jsonファイルではなくデータベース(SQLite)にデータを保存する利点は何がありますか？
- JSON is data markup format. You use it to define what the data is and means
- SQL is a data manipulation language. You use it to define the operations you want to perform on the data

[stackoverflow](https://stackoverflow.com/questions/22071735/difference-between-json-and-sql#:~:text=JSON%20is%20the%20data%20format,store%20or%20retrieve%20the%20data.)

###  step3-5 search

URIでドメインの後、？の前に来るものがパスパラメータ、?の後に来るのがクエリパラメータ

https://example.com/{pathparameter}?queryparameter1=hogehoge&queryparameter2=fugafuga


- 出力がJson形式で返ってこない問題

[ORJSONResponce](https://fastapi.tiangolo.com/ja/advanced/custom-response/)で解決


### step3-6 image

> sha256以外にどんなハッシュ関数があるか調べてみましょう
file upload
[Fast Api　re](https://fastapi.tiangolo.com/tutorial/request-files/)


sha256 hash
- hashlib
[Pythonでハッシュ値を生成するには？](https://create-it-myself.com/know-how/generate-hash-from-image-in-python/)

- バイナリファイルの読み書き
[Pythonでファイルの読み込み、書き込み](https://note.nkmk.me/python-file-io-open-with/)
