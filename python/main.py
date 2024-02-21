import os
import logging
import pathlib
from fastapi import FastAPI, HTTPException, Form, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import Optional
import json
import hashlib


app = FastAPI()
logger = logging.getLogger("uvicorn")

# 画像を保存するディレクトリを確認（存在しなければ作成）
images_dir = "images"
if not os.path.exists(images_dir):
    os.makedirs(images_dir)

# CORSの設定
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# カスタムエラークラスの定義
class ErrorL107(Exception):
    def __init__(self, message="Error L107: File not found"):
        self.message = message
        super().__init__(self.message)

# ルートエンドポイント
@app.get("/")
def root():
    return {"message": "Hello, world!"}

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
        return {"image_path": image_path}
        # 画像をファイルに保存
        with open(image_path, "wb") as file:
            file.write(contents)
        logger.info(f"Image saved: {image_name}")
        # return {"image_path": image_path}

    # 新しいアイテムIDの決定
    new_item_id = 1
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            if data["items"]:
                new_item_id = max(item["item_id"] for item in data["items"]) + 1
    except FileNotFoundError:
        raise ErrorL107()
    # except json.JSONDecodeError:
    #     # JSONファイルが空または不正な形式の場合のエラー処理
    #     data = {"items": []}

    # アイテムデータの作成
    item_data = {"item_id": new_item_id, "name": name, "category": category, "image_name": image_name}
    data["items"].append(item_data)

    with open("items.json", "w") as file:
        json.dump(data, file, indent=4)

    logger.info(f"Item added: {name}, Category: {category}, Image Name: {image_name}, Item ID: {new_item_id}")
    return {"message": "Item added successfully", "item_id": new_item_id}

# アイテムゲット
@app.get("/items")
async def get_items():
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        # return {"detail": "Items not found."}
        raise ErrorL107()

# 画像取得エンドポイント
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    image_path = pathlib.Path(images_dir) / image_name
    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = pathlib.Path(images_dir) / "default.jpg"
    return FileResponse(image_path)

# 特定のアイテムを取得するエンドポイント
@app.get("/items/{item_id}")
async def get_item(item_id: int):
    try:
        with open("items.json", "r") as file:
            data = json.load(file)
            # item_idに一致する商品を検索
            item = next((item for item in data["items"] if item["item_id"] == item_id), None)
            if item:
                return item
            else:
                raise HTTPException(status_code=404, detail="Item not found")
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="Items file not found")
