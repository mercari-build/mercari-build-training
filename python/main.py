import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
# items.jsonに新しいアイテムを追加した時のデータを追加するためにjsonファイルのパスを指定
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"


def get_db():
    if not db.exists():
        yield

    conn = sqlite3.connect(db)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    pass


@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

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


class HelloResponse(BaseModel):
    message: str


# APIサーバが正しく動作しているかの簡単なテストとして利用
@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})


class AddItemResponse(BaseModel):
    message: str


# POST-/item リクエストで呼び出され、アイテム情報の追加を行う
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    
    if not category:
        raise HTTPException(status_code=400, detail="category is required")

    new_item = Item(name=name, category=category)
    insert_item(new_item)
    return AddItemResponse(**{"message": f"item received: {name} category received: {category}"})


# get_image is a handler to return an image for GET /images/{filename} .
# GET-/image/{image_name} リクエストで呼び出され、指定された画像を返す
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

# GET-/items リクエストで呼び出され、items.jsonファイルの内容(今まで保存された全てのitemの情報)を返す
@app.get("/items")
def get_items():
    if not items_file.exists():
        return {"items": []}
    
    try:
        with open(items_file, "r", encoding="utf-8") as f:
            data = json.load(f)
    except json.JSONDecodeError:
        data = {"items": []}
        
    return data

class Item(BaseModel):
    name: str
    category: str

# app.post("/items" ... のハンドラ内で用いられる。items.jsonファイルへ新しい要素の追加を行う。
def insert_item(item: Item):
    # STEP 4-1: add an implementation to store an item
    global items_file
    
    # ファイルが存在しない場合初期状態で作成する
    if not items_file.exists():
        with open(items_file, "w", encoding="utf-8") as f:
            json.dump({"items": []}, f, ensure_ascii=False, indent=2)
            
    # 既存のファイルがあった場合読み込み
    try:
        # すでにデータがある場合
        with open(items_file, "r", encoding="utf-8") as f:
            data = json.load(f)
        # ファイルはあるがデータが空の場合
    except json.JSONDecodeError:
        data = {"items": []}
        
    # 新しいアイテムを追加する
    data["items"].append({"name": item.name, "category": item.category})
    
    #更新したデータを書き戻す
    with open(items_file, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii="False", indent=2)
        