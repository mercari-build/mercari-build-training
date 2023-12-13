# STEP2: 環境構築

PythonかGoから1つの言語を選び、環境を作りましょう。

---
## Pythonの環境を作る

### 1. Pythonをインストールする
* Python3.7以上がインストールされていない場合、Python3.10をインストールします
* すでに3.7以上がインストール済みの方はskipして問題ないです

### 2. Pythonのバージョンをチェックする

* インストールしたPythonにパスが通っている(=ターミナルから使える状態)か確認します。

```shell
$ python -V
```

表示されるPythonのバージョンがインストールしたものではなければ、**パスが通っていない**状態なので確認してください。

**:book: Reference**

* [PATHを通すとは？- 初心者でも分かる解説](https://hara-chan.com/it/programming/environment-variable-path/)

### 3. 依存ライブラリをインストールする

Pythonでは、`requirements.txt`というファイルに依存しているライブラリの一覧を記載します。
以下のコマンドを実行することで、依存ライブラリをまとめてインストールすることができます。

```shell
$ cd python
$ pip install -r requirements.txt
```

追加でライブラリをインストールした場合は、requirements.txtにライブラリを追加するようにしましょう。

### 4. アプリにアクセスする

```shell
$ uvicorn main:app --reload --port 9000
```

起動に成功したら、 ブラウザで `http://127.0.0.1:9000` にアクセスして、`{"message": "Hello, world!"}`
が表示されれば成功です。

---

## Goの環境を作る
### 1. Goをインストールする
* Go1.14以上がインストールされていない場合、Go1.18をインストールします
* すでに1.14以上がインストール済みの方はskipして問題ないです

### 2. Goのバージョンをチェックする

* インストールしたGoにパスが通っている(=ターミナルから使える状態)か確認します。

```shell
$ go version
```

表示されるGoのバージョンがインストールしたものではなければ、**パスが通っていない**状態なので確認してください。

**:book: Reference**

* [PATHを通すとは？- 初心者でも分かる解説](https://hara-chan.com/it/programming/environment-variable-path/)

### 3. 依存ライブラリをインストールする

Goでは、`go.mod`というファイルで依存しているライブラリを管理しています。
以下のコマンドを実行することで、依存ライブラリをインストールすることができます。

```shell
$ cd go
$ go mod tidy
```

**:beginner: Point**

[このdocument](https://pkg.go.dev/cmd/go#hdr-The_go_mod_file)を参考に go.mod の役割や go.mod を扱うコマンドについて理解しましょう。

### 4. アプリにアクセスする

```shell
$ go run app/main.go
```

起動に成功したら、 ブラウザで `http://127.0.0.1:9000` にアクセスして、`{"message": "Hello, world!"}`
が表示されれば成功です。

---
**:beginner: Point**

* (LinuxやMacの場合) `.bash_profile` や `.bashrc` (zshを使っている場合は`.zshrc`)
  等はどのタイミングで呼ばれ、何をしているのか理解しましょう。
* **パスを通す** の意味を理解しましょう

**:book: Reference**

環境構築の仕方やlinuxについてさらにしっかり学ぶためには以下の教材がおすすめです。

* (JA)[book - [試して理解]Linuxのしくみ ~実験と図解で学ぶOSとハードウェアの基礎知識](https://www.amazon.co.jp/dp/477419607X/ref=cm_sw_r_tw_dp_178K0A3YTGA97XRH318R)
* (JA)[Udemy Business - もう絶対に忘れない Linux コマンド【Linux 100本ノック+名前の由来+丁寧な解説で、長期記憶に焼き付けろ！](https://mercari.udemy.com/course/linux100test/)
  * ↑わかりやすい講座だと思い貼ってますが、コマンドの暗記は特にしなくていいです

* (EN)[An Introduction to Linux Basics](https://www.digitalocean.com/community/tutorials/an-introduction-to-linux-basics)
* (EN)[Udemy Business - Linux Mastery: Master the Linux Command Line in 11.5 Hours](https://mercari.udemy.com/course/linux-mastery/)
  * You do NOT have to memorize the commands!

---
### Next

[STEP3: 出品APIを作る](03-api.ja.md)