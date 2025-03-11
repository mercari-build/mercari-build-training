# STEP6: テストを用いてAPIの挙動を確認する

このステップではテストに関する内容を学びます。

**:book: Reference**

* (JA)[テスト駆動開発](https://www.amazon.co.jp/dp/4274217884)
* (EN)[Test Driven Development: By Example](https://www.amazon.co.jp/dp/0321146530)

## テストとは
テストとは、システムやコンポーネントの挙動や性能を評価し、それらが仕様や要件を満たしているか確認するプロセスのことです。例えば、次のような `sayHello` というGoの関数について考えてみましょう。

```go
func sayHello(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
```

この関数は見れば分かる通り、引数で渡される `name` 変数を用いて `Hello, ${name}!` のような文字列を組み立てる関数です。この関数は正しく振る舞うでしょうか？それを確認することが出来るのがテストです。

Goでは以下のようにテストを書くことが出来ます。詳しい書き方は後ほど記載するので、今は流し読みしてください。

```go
func TestSayHello(t *testing.T) {
    // Alice というテスト名でテストを実施
    t.Run("Alice", func(t *testing.T) {
        // 期待する返り値は Hello, Alice!
        want := "Hello, Alice!"

        // 引数は "Alice"
        arg := "Alice"
        // 実際に sayHello を呼び出す
        got := sayHello(arg)

        // 期待する返り値と実際に得た値が同じか確認
        if want != got {
            // 期待する返り値と実際に得た値が異なる場合は、エラーを表示
            t.Errorf("unexpected result of sayHello: want=%v, got=%v", want, got)
        }
    })
}
```

これを実行すると、下記の通り成功します。

```bash
=== RUN   TestSayHello
=== RUN   TestSayHello/Alice
--- PASS: TestSayHello (0.00s)
    --- PASS: TestSayHello/Alice (0.00s)
PASS
```

このようにして、関数などの機能をテストすることが出来ます。

## テストの目的
このテストには、下記のような目的があります。

- 欠陥の発見
- 要件の適合性の検証
- 性能評価
- 信頼性の評価
- セキュリティの評価
- ユーザビリティの評価
- 保守性の評価
など

特に、想定されている挙動をすることを保証してくれるのは大きなメリットです。例えば、今回のソースコードに、気づかないうちに変な文字列( `#` )を下記のように紛れ込ませてしまったとします。

```go
func sayHello(name string) string {
    return fmt.Sprintf("Hello, %s!#", name)
}
```

この時、目視では見落としてしまう可能性があります。しかし、テストを書いておくことで、このようなミスに気づくことが可能です。実際にテストを実行すると以下の通り失敗してエラーメッセージが表示されます。

```bash
=== RUN   TestSayHello
=== RUN   TestSayHello/Alice
    prog_test.go:20: unexpected result of sayHello: want=Hello, Alice!, got=Hello, Alice!#
--- FAIL: TestSayHello (0.00s)
    --- FAIL: TestSayHello/Alice (0.00s)
FAIL
```

このようにして、振る舞いをテストで保証することでソースコードの品質を担保することが出来ます。更に、複雑な機能を実装する際に、小さな機能ごとにテストを書きながら実装を進めることで、確実に動く部分を保証しながら開発を進めることが出来ます。これにより、想定外のバグが発生した時でも、原因となった箇所をある程度絞り込んで調査できるので、テストを書かないときと比較して迅速に対応することが可能です。

## テストの種類
このテストには用途に応じて様々な種類があります。

今回は簡単のために、上記のようなコンポーネントレベルで実施されるのが単体テスト(Unit Tests)、システム全体を統合した上で、ユーザの操作をシミュレーションしてテストするエンドツーエンドテスト(End-to-End Test/E2E Tests)の2種類を紹介します。興味のある方は各自で調べてください。

ここで、具体例に沿って考えてみましょう。例えば、画像投稿サイトで画像を投稿する機能のためのAPIの機能をテストする場合を想像してください。この場合、画像を投稿するAPIは、画像データを受け取って、結果を返却する関数/メソッドで実装されているはずです。そのため、想定される入力と出力を用いてテストすることが出来そうです。

しかし、テストのために毎回データベースを用意したり、サーバを起動したりするのは骨が折れます。そこで、画像をデータベースに保存する処理を実際に行わず、保存処理をするための関数/メソッドを、固定値を返す別の実装に置き換えてテストを行うことが出来ます。このような、テストのために固定値を返すようなものをモックと呼びます。

このようなモックを用いて、データベースに対する保存処理が失敗した時の挙動や成功した時の挙動を、実際にデータベースを用意せずに事細かに保証することが出来ます。しかし、このモックはあくまで僕らが勝手に指定した値なので、実際の挙動と異なるテストをしている可能性もあります。

このように、小さな機能のみのテストやモック等を用いた偽データを用いるテストを単体テスト(Unit Tests)と呼び、実際のデータベースやデータを用いて全体の機能をテストするテストをエンドツーエンドテスト(End-to-End tests: E2E tests)と呼びます。

基本的に単体テストの数がE2Eテストの数よりも多くなることが推奨されます。なぜなら、単体テストは高速かつ少ないリソースで実行することが出来ます。E2Eテストは遅く多くのリソースを必要とするためです。例えば、実データを利用するテストの場合、先ほどの例で考えると、テストデータを複数用意して、保存や削除を複数回行う必要があります。大規模データを扱う場合は実行時間が長くなったり利用リソースが増えたりするため、E2Eテストを少なめにして、単体テストで小さな機能を数でカバーするのが定石です。とはいえ、単体テストのみでは実際の環境固有で起きる問題に気付けなくなるという問題があるため、バランスが大事です。

## テスト戦略
テストの方針は言語、フレームワークによって異なります。本節では、GoとPythonにおけるテスト戦略について説明し、実際にテストを書く方法について説明します。

### Go

**:book: Reference**

- (EN)[testing package - testing - Go Packages](https://pkg.go.dev/testing)
- (EN)[Add a test - The Go Programming Language](https://go.dev/doc/tutorial/add-a-test)
- (EN)[Go Wiki: Go Test Comments - The Go Programming Language](https://go.dev/wiki/TestComments)
- (EN)[Go Wiki: TableDrivenTests - The Go Programming Language](https://go.dev/wiki/TableDrivenTests)

Goはテストに関連する機能を提供する `testing` と呼ばれる標準パッケージを有しており、 `$ go test` コマンドによってテストを行うことが可能です。Goが提示しているテストの方針については、[Go Wiki: Go Test Comments](https://go.dev/wiki/TestComments)を参照してください。言語としての一般的な方針が書かれています。これらの方針は必須という訳ではないので、問題のない範囲で倣うのが良いと思います。

では、実際に先ほどのコードの単体テストから書いてみましょう。Goではテストしたいケースを最初に列挙して、テーブルのように順番にテストするテーブルテスト(Table-Driven Test)を推奨しています。テストケースは基本的にスライスかmapで宣言することが多々ありますが、順序性が必要とされるケースでなければ、基本的にmapを利用すると良いと思います。実行順序に依存しないテストケースを書くことで、テスト対象の機能の振る舞いを、より強固に保証することが可能になるためです。

```go
func TestSayHello(t *testing.T) {
    cases := map[string]struct{
        name string
        want string
    }{
        "Alice": {
            name: "Alice",
            want: "Hello, Alice!"
        }
        "empty": {
            name: "",
            want: "Hello!"
        }
    }

    for name, tt := range cases {
        t.Run(name, func(t *testing.T) {
            got := sayHello(tt.name)

            // 期待する返り値と実際に得た値が同じか確認
            if tt.want != got {
                // 期待する返り値と実際に得た値が異なる場合は、エラーを表示
                t.Errorf("unexpected result of sayHello: want=%v, got=%v", tt.want, got)
            }
        })
    }
}
```

このように、テストケースをまとめて書くことで、一目で入力と想定される出力を確認することが出来ます。仮に、対象の関数/メソッドの振る舞いを全く知らないでコードリーディングする必要がある場合、テストコードを参考にして振る舞いを理解するヒントとして利用することもできます。

また、このようなテストも想定して、引数の設計を考えることも大事です。例えば、次のように、時間に応じて挨拶を変えるようにしたとします。

```go
func sayHello(name string) string {
    now := time.Now()
    currentHour := now.Hour()

    if 6 <= currentHour && currentHour < 10 {
        return fmt.Sprintf("Good morning, %s!", name)
    }
    if 10 <= currentHour && currentHour < 18 {
        return fmt.Sprintf("Hello, %s!", name)
    }
    return fmt.Sprintf("Good evening, %s!", name)
}
```

この場合、各時間帯の全てでテストをするためには、それぞれの時間にテストを実施しなければなりません。これはテスト的に適していない設計と言えます。テストできるようにするために、以下のように関数を書き換えることが出来ます。

```go
func sayHello(name string, now time.Time) string {
    currentHour := now.Hour()

    if 6 <= currentHour && currentHour < 10 {
        return fmt.Sprintf("Good morning, %s!", name)
    }
    if 10 <= currentHour && currentHour < 18 {
        return fmt.Sprintf("Hello, %s!", name)
    }
    return fmt.Sprintf("Good evening, %s!", name)
}
```

これにより、現在時刻を自由に設定できるようになったため、以下のように各時間帯の振る舞いをテストできるようになります。

```go
func TestSayHelloWithTime(t *testing.T) {
    type args struct {
        name string
        now time.Time
    }
    cases := map[string]struct{
        args
        want string
    }{
        "Morning Alice": {
            args: args{
                name: "Alice",
                now: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
            },
            want: "Good morning, Alice!",
        },
        "Hello Bob": {
            args: args{
                name: "Bob",
                now: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
            },
            want: "Hello, Bob!",
        },
        "Night Charie": {
            args: args{
                name: "Charie",
                now: time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
            },
            want: "Good evening, Charie!",
        },
    }

    for name, tt := range cases {
        t.Run(name, func(t *testing.T) {
            got := sayHello(tt.name, tt.now)

            // 期待する返り値と実際に得た値が同じか確認
            if tt.want != got {
                // 期待する返り値と実際に得た値が異なる場合は、エラーを表示
                t.Errorf("unexpected result of sayHello: want=%v, got=%v", tt.want, got)
            }
        })
    }
}
```

このようにして、テストのことも意識したコードをかけると良いですね。

### Python

Pythonにおけるテスト戦略

**:book: Reference**

- (EN)[pytest: helps you write better programs — pytest documentation](https://docs.pytest.org/en/stable/)
- (EN)[pytest fixtures: explicit, modular, scalable — pytest documentation](https://docs.pytest.org/en/6.2.x/fixture.html)
- (EN)[Parametrizing fixtures and test functions — pytest documentation](https://docs.pytest.org/en/stable/how-to/parametrize.html)

Pythonにはtest用のライブラリとして標準搭載されているunittestがありますが、より柔軟で可読性の高いテストを書くために**pytest**というライブラリが広く利用されています。pytestはシンプルなAPIと強力な機能を備えており、`pip install pytest`で簡単にインストールできます。`$ pytest`コマンドでテストを実行することができます。


Pythonでは`pytest.mark.parametrize`デコレータを使って複数のテストケースをまとめて記述できます。say_hello関数のテストを書いてみましょう。

```python
# hello.py
def say_hello(name=""):
    if name:
        return f"Hello, {name}!"
    return "Hello!"

# test_hello.py
import pytest
from hello import say_hello

@pytest.mark.parametrize("name, expected",[
    ("Alice", "Hello, Alice!"),
    ("", "Hello!"),
]
)
def test_say_hello(name, expected):
    got = say_hello(name)

    # 期待する返り値と実際に得た値が同じか確認した上で, 期待する返り値と実際に得た値が異なる場合は、エラーを表示
    assert got == expected, f"unexpected result of say_hello: want={expected}, got={got}"
```

テストを想定して、引数の設計を考える必要があるという点はPythonもGoと共通です。`say_hello`の実装を時間に応じて挨拶を変えるように変更することを想定します。

```python
from datetime import datetime

def say_hello(name):
    now = datetime.now() # 現在時刻に直接依存しているため、テストしにくい
    current_hour = now.hour

    if 6 <= current_hour < 10:
        return f"Good morning, {name}!"
    if 10 <= current_hour < 18:
        return f"Hello, {name}!"
    return f"Good evening, {name}!"
```

この関数は現在時刻に直接依存しているため、テストが難しい設計です。各時間帯をテストするためには、実際にその時間にテストを実行する必要があります。

テストしやすくするために、関数を次のように書き換えます：

```python
# 改善されたコード（テストしやすい設計）
from datetime import datetime

def say_hello(name, now=None):
    if now is None:
        now = datetime.now()

    current_hour = now.hour

    if 6 <= current_hour < 10:
        return f"Good morning, {name}!"
    if 10 <= current_hour < 18:
        return f"Hello, {name}!"
    return f"Good evening, {name}!"
```

これで現在時刻を引数として指定できるようになりました。デフォルト値としてNoneを設定することで、通常の使用ではnowを省略することもできます。

```python
import pytest
from datetime import datetime
from greetings import say_hello

@pytest.mark.parametrize("name, now, expected", [
    ("Alice", datetime(2024, 1, 1, 9, 0, 0), "Good morning, Alice!"),
    ("Bob", datetime(2024, 1, 1, 12, 0, 0), "Hello, Bob!"),
    ("Charlie", datetime(2024, 1, 1, 20, 0, 0), "Good evening, Charlie!"),
])
def test_say_hello_simple(name, now, expected):
    got = say_hello(name, now)
    assert got == expected, f"unexpected result of say_hello: want={expected}, got={got}"
```

## 1. 出品APIのテストを書く
基礎的な機能のテストである、アイテム登録のためのリクエストのテストを書いてみましょう。

想定されるリクエストは、 `name` および `category` を必要とするはずです。
そのため、そのデータが欠けている時にエラーを返すべきです。これをテストしてみましょう。

### Go
`server_test.go` を見てみましょう。

現在、AddItemのリクエストが来た時に全ての値が含まれている場合はOK、欠けている値がある場合はNGとしたいです。
そのようなテストケースを書いてみましょう。

**:beginner: Point**

- このテストは何を検証しているでしょうか？
- `t.Error()` と `t.Fatal()` には、どのような違いがあるでしょうか？

### Python(Read Only)

Pythonのテストは[`main_test.py`](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py)に実装されています。

GoのAPI実装と異なり、FastAPIというフレームワークを活用したPythonのAPI実装ではHTTP RequestをParseする処理を開発者で実装する必要がありません。そのため、本章で追加の実装は必要ありませんが、テストコードに目を通し、理解を深めておきましょう。


## 2. Hello Handlerのテストを書く
ハンドラのテストを書いてみましょう。

ハンドラのテストを書く際は、STEP 6-1と同様に、想定される値と引数を比較すれば良さそうです。

### Go

**:book: Reference**

- (EN)[httptest package - net/http/httptest - Go Packages](https://pkg.go.dev/net/http/httptest)
- (JA)[Goのtestを理解する - httptestサブパッケージ編 - My External Storage](https://budougumi0617.github.io/2020/05/29/go-testing-httptest/)

Goでは、 `httptest` と呼ばれるハンドラをテストするためのライブラリを用いてみましょう。

今回は、STEP6-1の時と異なり、比較する部分のコードが書かれていません。

- このハンドラでテストしたいのは何でしょうか？
- それが正しい振る舞いをしていることはどのようにして確認できるでしょうか？

ロジックが思いついたら実装してみましょう。

**:beginner: Point**

- 他の方が書いたテストコードを確認してみましょう
- httptestパッケージの既存コードで何をしているか確認してみましょう

### Python

- (JA)[FastAPI > チュートリアル > ユーザーガイド / テスト](https://fastapi.tiangolo.com/ja/tutorial/testing/)

Pythonでは、FastAPIが提供する`testclient.TestClient`を用いて、ハンドラとなる`hello`が正しく動作するかどうかを検証します。すでに用意されているテスト用の関数の[test_hello](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py#L53)を編集して、テストを書いてみましょう。

Goと同じように以下のことを意識しながら、テストコードを実装してみましょう。

- このハンドラでテストしたいのは何でしょうか？
- それが正しい振る舞いをしていることはどのようにして確認できるでしょうか？

テストの実装には、[FastAPIの公式document](https://fastapi.tiangolo.com/ja/tutorial/testing/#testclient)を参考にしてみてください。

## 3. モックを用いたテストを書く
モックを用いたテストを書いてみましょう。

モックは、先述の通り、実際のロジックを用いるのではなく、想定されたデータを返すような便利関数と実際の関数を置き換えるためのものです。このモックは様々な部分で利用できます。

例えば、今回のデータベースへのアイテム登録の部分を考えてみましょう。テストでは、データベースへのアイテム登録に成功する時と失敗する時を両方テストしたいはずです。しかし、これらのケースを意図的に引き起こすことは少々手間がかかります。また、実際のデータベースを利用すると、データベース側の問題でテストがflakyになる可能性もあります。

そこで、実際にデータベースのロジックを用いるのではなく、想定された返り値を返すようなモックを用いることで、あらゆるケースをテストすることが可能です。

### Go

**:book: Reference**

- (EN) [mock module - go.uber.org/mock - Go Packages](https://pkg.go.dev/go.uber.org/mock)

Goには様々なモックライブラリがありますが、今回は `gomock` を利用します。
`gomock` の簡単な利用方法はドキュメントや先駆者のブログを参照してください。

このモックを用いて、永続化の処理が成功するパターンと失敗するパターンの両方をテストしてみましょう。

**:beginner: Point**

- モックを満たすためにinterfaceを用いていますが、interfaceのメリットについて考えてみましょう
- モックを利用するメリットとデメリットについて考えてみましょう

### Python (Read Only)

**:book: Reference**

- (EN) [pytest-mock](https://github.com/pytest-dev/pytest-mock)
- (JA) [unittest.mock --- 入門](https://docs.python.org/ja/3.13/library/unittest.mock-examples.html#)

Pythonのmock用のライブラリとしては標準搭載された`unittest.mock`やpytestが提供する`pytest-mock`等の選択肢が存在します。モックが必要になるケースとしては、テスト対象となる処理が外部的なツールやオブジェクトに依存する場合、例えば以下のようなケースが挙げられます。

- データベース接続をモックして、実際のDBに接続せずにユーザー認証ロジックをテストする。
- HTTP APIクライアントをモックして、実際のネットワーク通信なしで天気予報取得関数をテストする。
- ファイルシステムをモックして、実際のファイル操作なしでログ出力機能をテストする。

今回の場合、例の一番最初にあげた「データベース接続をモックする」というテストの実装が考えられます。しかし、BuildのPythonによるAPI実装は非常にシンプルなもので、モックしたテストを書くために、ItemRepositoryのようなクラスを設けることは必要以上に実装を複雑にしてしまいます。

「4. 実際のデータベースを用いたテストを書く」という章で実装するテストコードで十分な検証ができる上に、Pythonの「シンプルさ」と「明示的であること」を重視する言語哲学にも反すると考え、**今回の教材からはmockを用いたpythonの実装を省いています**

ただし、実際の開発現場のように、アプリケーションが複雑化した場合はmockを用いたテストの実装をPythonでも実施するケースが多いです。興味があるという方は、Goの方のmockを用いたテストの説明に目を通してみたり、インターネット上で紹介されているmockを用いたpythonのテスト実装に目を通してみましょう。

## 4. 実際のデータベースを用いたテストを書く
STEP 6-3におけるモックを実際のデータベースに置き換えたテストを書いてみましょう。

モックは先述の通りあらゆるケースをテストすることが可能ですが、実際の環境で動かしている訳ではありません。そのため、実際のデータベース上では動かない、ということもしばしばあります。そこで、テスト用にデータベースを用意して、そのデータベースを利用してテストを実施しましょう。

### Go
Goでは、テスト用にデータベース用のファイルを作成して、そこに処理を足していく方針を取ります。

実際のデータベースで処理を行った後、データベース内のデータが想定通り変更されていることを確かめる必要があります。

- アイテム登録後のデータベースの状態はどうなっているはずでしょうか？
- それが正しい振る舞いをしていることはどのようにして確認できるでしょうか？

### Python

Pythonで、テスト用のデータベース(sqlite3)を用いたテストを書いていきましょう。[main_test.py]()の「`STEP 6-4: uncomment this test setup`」という記載がある二カ所をコメントアウトしてください。([一カ所目](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py#L9-L42)/[二カ所目](https://github.com/mercari-build/mercari-build-training/blob/main/python/main_test.py#L60-L84))


`db_connection`関数で、テスト前にはsqlite3を用いたテスト用のdbの新規作成とセットアップを行い、テストの終了後にはテスト用のdbを削除する処理を行なっています。

`test_add_item_e2e`では、APIのエンドポイント（`/items/`）に対してPOSTリクエストを送信し、アイテム追加機能をテストしています。この関数は複数のテストケース（有効なデータと無効なデータ）をパラメータ化して実行します。テストでは以下を検証します：

1. レスポンスのステータスコードが期待値と一致するか
2. エラーでない場合は、レスポンスボディに「message」が含まれているか
3. データベースに正しくデータが保存されたか（名前とカテゴリが一致するか）

特に重要なのは、モックではなく実際のデータベース（テスト用）を使用してエンドツーエンドでテストすることで、実際の環境に近い形で機能を検証している点です。

## Next

[STEP7: 仮想環境でアプリを動かす](./07-docker.ja.md)
