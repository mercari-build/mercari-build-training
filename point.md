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
