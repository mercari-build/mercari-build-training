## Step3 

## 基本的なAPIの使い方
1つ目のターミナルでサーバーを起動

```
(.venv) uvicorn main:app --reload --port 9000
```

停止するときは同じコマンドでcontrl + c

## APIの編集について
main.pyに書き込む
元のmain.py
```
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)


@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.post("/items")
def add_item(name: str = Form(...)):
    logger.info(f"Receive item: {name}")
    return {"message": f"item received: {name}"}



@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
```

## 編集後のAPI Get や　Postに対する操作を追加する
```
import os
import logging
import pathlib
from fastapi import FastAPI, HTTPException, Body
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import json

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO

# 画像を保存するディレクトリへのパス
images = pathlib.Path(__file__).parent.resolve() / "images"

# CORSの設定
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# Itemモデルの定義
class Item(BaseModel):
    name: str
    category: str

# ルートエンドポイント
@app.get("/")
def root():
    return {"message": "Hello, world!"}

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

# アイテムゲット
@app.get("/items")
async def get_items():
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        return {"detail": "Items not found."}



# 画像取得エンドポイント
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    image_path = images / image_name
    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = images / "default.jpg"
    return FileResponse(image_path)
```


# ターミナルからのアクセス
## ドキュメントに指定されていたコードと応答
```
(.venv) curl -X POST \
  --url 'http://localhost:9000/items' \
  -d 'name=jacket' \
  -d 'category=fashion'
{"detail":[{"type":"model_attributes_type","loc":["body"],"msg":"Input should be a valid dictionary or object to extract fields from","input":"name=jacket&category=fashion","url":"https://errors.pydantic.dev/2.6/v/model_attributes_type"}]}%    
```
## 変更した方がうまく応答した
```
(.venv) curl -X POST \
  --url 'http://localhost:9000/items' \
  -H 'Content-Type: application/json' \
  -d '{"name": "jacket", "category": "fashion"}'

{"message":"item received: jacket, Category: fashion"}%
```
## Get method 
```
curl -X GET 'http://127.0.0.1:9000/items'
{"items":[{"name":"jacket","category":"fashion"},{"name":"jacket","category":"fashion"}]}%    
```
これはmain.pyに以下の項を加えたことで可能になった
```
@app.get("/items")
async def get_items():
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        return {"detail": "Items not found."}
```


