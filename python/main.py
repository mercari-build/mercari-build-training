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
