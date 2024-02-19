# STEP3: 出品APIを作る

## 1. APIをたたいてみる

**:book: Reference**

* (JA) [Udemy Business - REST WebAPI サービス 設計](https://mercari.udemy.com/course/rest-webapi-development/)
* (JA) [HTTP レスポンスステータスコード](https://developer.mozilla.org/ja/docs/Web/HTTP/Status)
* (JA) [HTTP リクエストメソッド](https://developer.mozilla.org/ja/docs/Web/HTTP/Methods)
* (JA) [APIとは？意味やメリット、使い方を世界一わかりやすく解説](https://www.sejuku.net/blog/7087)

* (EN) [Udemy Business - API and Web Service Introduction](https://mercari.udemy.com/course/api-and-web-service-introduction/)
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

**< note 開始>**

<これはcategoryの情報も受け取るためのコードをmain.pyの@app.post("/items")に書きましょうという意味>

main.py
```
# アイテム追加エンドポイント
@app.post("/items")
async def add_item(item: Item = Body(...)):
    logger.info(f"Received item: {item.name}, Category: {item.category}")  # ログ出力
    # アイテムをJSONファイルに保存
    try:
        with open("items.json", "r+") as file:
            data = json.load(file)
            data["items"].append(item.dict())
            file.seek(0)
            json.dump(data, file, indent=4)
            file.truncate()
    except FileNotFoundError:
        with open("items.json", "w") as file:
            json.dump({"items": [item.dict()]}, file, indent=4)
    
    logger.info(f"Item added: {item.name}, Category: {item.category}")
    return {"message": f"item received: {item.name}, Category: {item.category}"}
```
<ターミナルでのPOST>
```
curl -X POST \
  --url 'http://localhost:9000/items' \
  -H 'Content-Type: application/json' \
  -d '{"name": "jacket", "category": "fashion"}'
```
*しかし，この後のコードの改変により，上記のコマンドではエラーが返ってくるようになった．
代わりにPOSTは以下のコマンドに対応
(なぜ??)
```
curl -X POST \
 --url 'http://localhost:9000/items' \
 -F 'name=jacket' \
 -F 'category=fashion' \
```


**< note 終了>**



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

**< note 開始>**

**これは録された商品一覧を取得するためのコードをmain.pyの@app.get("/items")に書きましょうという意味**

main.py
```
# アイテムゲット
@app.get("/items")
async def get_items():
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        return {"detail": "Items not found."}
```
**<ターミナルでのGET>**
```
curl -X GET 'http://127.0.0.1:9000/items'
```
**< note 終了>**

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


## 4. 画像を登録する

商品情報に画像(image)を登録できるように、`GET /items`と`POST /items`のエンドポイントを変更します。

* 画像は `images` というフォルダを作成し保存します
* ポストされた画像のファイルを sha256 で hash化し、`<hash>.jpg`という名前で保存します
* itemsに画像のファイル名をstringで保存できるように変更を加えます

```shell
# ローカルから.jpgをポストする
curl -X POST \
  --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
```


```json
{"items": [{"name": "jacket", "category": "fashion", "image_name": "510824dfd4caed183a7a7cc2be80f24a5f5048e15b3b5338556d5bbd3f7bc267.jpg"}, ...]}
```
**<note 開始>**

main.pyのinportの修正
```
import os
import logging
import pathlib
from fastapi import FastAPI, HTTPException, Body, File, UploadFile, Form
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional
import json
import hashlib
```


main.pyのpostの修正

```
@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = None):
    # アイテム情報のログ出力
    logger.info(f"Received item: {name}, Category: {category}")

    # 画像ファイルがある場合は処理
    image_name = None
    if image:
        # 画像の内容を読み取り
        contents = await image.read()
        # 画像のハッシュ値を計算してファイル名を生成
        hash_name = hashlib.sha256(contents).hexdigest()
        image_name = f"{hash_name}.jpg"
        image_path = os.path.join(images_dir, image_name)
        # 画像をファイルに保存
        with open(image_path, "wb") as file:
            file.write(contents)
        logger.info(f"Image saved: {image_name}")

    # アイテムをJSONファイルに保存
    item_data = {"name": name, "category": category, "image_name": image_name}
    try:
        with open("items.json", "r+") as file:
            data = json.load(file)
            data["items"].append(item_data)
            file.seek(0)
            json.dump(data, file, indent=4)
            file.truncate()
    except FileNotFoundError:
        with open("items.json", "w") as file:
            json.dump({"items": [item_data]}, file, indent=4)

    logger.info(f"Item added: {name}, Category: {category}, Image Name: {image_name}")
    return {"message": f"Item received: {name}, Category: {category}, Image Name: {image_name}"}
```

 それ以外にも直接的には関わらない部分を少し変えている

 ターミナル ディレクトリにある画像を指定
 ```
curl -X POST \
  --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/default.jpg'
```
**<note 終了>**


**:beginner: Point**

* Hash化とはなにか？
* sha256以外にどんなハッシュ関数があるか調べてみましょう

**<note 開始>**

Hash化とは，文字列や画像などの入力情報を別の文字列に置き換える作業．この時，同じ入力情報は同じ文字列に対応する．
MD5、SHA1、SHA256など．

**<note 終了>**

## 5. 商品の詳細を返す

商品の詳細情報を取得する  `GET /items/<item_id>` というエンドポイントを作成します。

```shell
$ curl -X GET 'http://127.0.0.1:9000/items/1'
{"name": "jacket", "category": "fashion", "image_name": "..."}
```

**<note 開始>**

ver1 uuid4を使用したとき >　IDが長いのであとで変更
main.py 

```
from uuid import uuid4
```


main.py　POSTで追加したアイテムにIDを割り当て
```
# 新しいアイテムに一意のIDを割り当て
    item_id = str(uuid4())
    item_data = {"item_id": item_id, "name": name, "category": category, "image_name": image_name}
```

main.pyに新しいend point
```
@app.get("/items/{item_id}")
async def get_item(item_id: str):
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            # item_idに一致する商品を検索
            item = next((item for item in data["items"] if item.get("item_id") == item_id), None)
            if item:
                return item
            else:
                raise HTTPException(status_code=404, detail="Item not found")
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="Items file not found")
```

ターミナル上で画像つきでリクエスト
```
curl -X POST \
 --url 'http://localhost:9000/items' \
 -F 'name=jacket' \
 -F 'category=fashion' \
 -F 'image=@images/default.jpg'
{"message":"Item added successfully","item_id":"078e7086-4019-4d90-a75c-5bac55f99b1d"}%      
```
ターミナル上で画像なしでリクエスト
```
curl -X POST \
 --url 'http://localhost:9000/items' \
 -F 'name=jacket' \
 -F 'category=fashion' \                         
{"message":"Item added successfully","item_id":"ad51a3bd-6642-4302-a60b-5d8886886e56"}%
```
json file に以下のように追加されていくことが確認された
```
        {
            "item_id": "078e7086-4019-4d90-a75c-5bac55f99b1d",
            "name": "jacket",
            "category": "fashion",
            "image_name": "ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg"
        },
        {
            "item_id": "ad51a3bd-6642-4302-a60b-5d8886886e56",
            "name": "jacket",
            "category": "fashion",
            "image_name": null
        }
```

**id を1,2, 3,..に変更**

main.pyの主な変更点

```

# アイテム追加エンドポイント
@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = None):
    # アイテム情報のログ出力
    logger.info(f"Received item: {name}, Category: {category}")

    # 画像ファイルがある場合は処理
    image_name = ""
    if image:
        # 画像の内容を読み取り
        contents = await image.read()
        # 画像のハッシュ値を計算してファイル名を生成
        hash_name = hashlib.sha256(contents).hexdigest()
        image_name = f"{hash_name}.jpg"
        image_path = os.path.join(images_dir, image_name)
        # 画像をファイルに保存
        with open(image_path, "wb") as file:
            file.write(contents)
        logger.info(f"Image saved: {image_name}")

    # 新しいアイテムIDの決定
    new_item_id = 1
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            if data["items"]:
                new_item_id = max(item["item_id"] for item in data["items"]) + 1
    except FileNotFoundError:
        data = {"items": []}

    # アイテムデータの作成
    item_data = {"item_id": new_item_id, "name": name, "category": category, "image_name": image_name}
    data["items"].append(item_data)

    with open("items.json", "w") as file:
        json.dump(data, file, indent=4)

    logger.info(f"Item added: {name}, Category: {category}, Image Name: {image_name}, Item ID: {new_item_id}")
    return {"message": "Item added successfully", "item_id": new_item_id}
```


なぜかよくわかないけどうまくいった方法
items.json をいかに書き換える
```
{
  "items": []
}
```

```
curl -X POST \ --url 'http://localhost:9000/items' \
 -F 'name=jacket' \
 -F 'category=fashion' \
> 
curl: (3) URL rejected: Malformed input to a URL function
{"message":"Item added successfully","item_id":1}%                        
```

```
curl -X POST \ --url 'http://localhost:9000/items' \
 -F 'name=jacket' \
 -F 'category=fashion' \
 -F 'image=@images/default.jpg'
curl: (3) URL rejected: Malformed input to a URL function
{"message":"Item added successfully","item_id":2}%
```
2つ実行後のjson
```
{
    "items": [
        {
            "item_id": 1,
            "name": "jacket",
            "category": "fashion",
            "image_name": ""
        },
        {
            "item_id": 2,
            "name": "jacket",
            "category": "fashion",
            "image_name": "ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg"
        }
    ]
}
```
**<note 終了>**

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

[STEP5: データベース](04-database.ja.md)
