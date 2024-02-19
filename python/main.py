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


app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO

# 画像を保存するディレクトリへのパス
# if not os.path.exists('images'):
#     os.makedirs('images')

# images = pathlib.Path(__file__).parent.resolve() / "images"

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

# Itemモデルの定義
class Item(BaseModel):
    name: str
    category: str
    image_name: str = None  # 画像はオプショナルとする

# ルートエンドポイント
@app.get("/")
def root():
    return {"message": "Hello, world!"}

# アイテム追加エンドポイント
# @app.post("/items")
# async def add_item(item: Item = Body(...)):
#     logger.info(f"Received item: {item.name}, Category: {item.category}")  # ログ出力
#     # アイテムをJSONファイルに保存
#     try:
#         with open("items.json", "r+") as file:
#             data = json.load(file)
#             data["items"].append(item.dict())
#             file.seek(0)
#             json.dump(data, file, indent=4)
#             file.truncate()
#     except FileNotFoundError:
#         with open("items.json", "w") as file:
#             json.dump({"items": [item.dict()]}, file, indent=4)
    
#     logger.info(f"Item added: {item.name}, Category: {item.category}")
#     return {"message": f"item received: {item.name}, Category: {item.category}"}

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
# @app.get("/image/{image_name}")
# async def get_image(image_name: str):
#     image_path = images / image_name
#     if not image_name.endswith(".jpg"):
#         raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
#     if not image_path.exists():
#         logger.debug(f"Image not found: {image_path}")
#         image_path = images / "default.jpg"
#     return FileResponse(image_path)
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    image_path = pathlib.Path(images_dir) / image_name
    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = pathlib.Path(images_dir) / "default.jpg"
    return FileResponse(image_path)