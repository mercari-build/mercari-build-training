import os
import logging
import pathlib
import json
import hashlib  # SHA-256　ハッシュ化用
from fastapi import FastAPI, Form, File, UploadFile, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
<<<<<<< HEAD
items_json_path = pathlib.Path(__file__).parent.resolve() / "items.json"  # items.json のパス追記
=======
items_json_path = pathlib.Path(__file__).parent.resolve() / "items.json"  # items.json のパスを追記
>>>>>>> origin/main


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
    images.mkdir(exist_ok=True, parents=True)  # imagesディレクトリを作成


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


@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})

# 画像をSHA-256でハッシュ化し、保存する関数
def save_image(image: UploadFile) -> str:
    # 画像データ読み込み
    image_data = image.file.read()
    # ハッシュ値作成
    hash_value = hashlib.sha256(image_data).hexdigest()
    # ハッシュ値を使ってファイル名を作成(重複を防ぐため)
    file_name = f"{hash_value}.jpg"
    # 画像を保存
    image_path = images /file_name
    with open(image_path,"wb") as file:
        file.write(image_data)
    return file_name #画像のハッシュ名を返す


class AddItemResponse(BaseModel):
    message: str

class Item(BaseModel):
    name: str
<<<<<<< HEAD
    category: str  # カテゴリー追加
    image_name: str #　画像のファイル名追加

# items.jsonのデータ読み込み
=======
    category: str  # カテゴリーも受け取れるように追加

# items.jsonの初期データ読み込み（items.jsonが存在しない場合の対応を）
>>>>>>> origin/main
def load_items():
    if items_json_path.exists():
        with open(items_json_path, "r", encoding="utf-8") as file:
            data = json.load(file)
            return data.get("items", [])
    return []

<<<<<<< HEAD
#item.jsonに商品を追加して保存する関数
def save_item(item: Item):
    items = load_items()
    items.append({"name": item.name, "category": item.category, "image_name": item.image_name}) #商品を追加する
=======

#item.jsonに商品を追加して保存する関数
def save_item(item: Item):
    items = load_items()
    items.append({"name": item.name, "category": item.category}) #商品を追加する
>>>>>>> origin/main

    with open(items_json_path, "w", encoding="utf-8") as file:
        json.dump({"items": items}, file, indent=2, ensure_ascii=False)  # JSONを保存

@app.get("/items") #登録された商品データをJSON形式で取得
def get_items():
    return {"items": load_items()}

<<<<<<< HEAD
@app.get("/items/{item_id}")
def get_item(item_id: int):
    items = load_items()

    if item_id < 0 or item_id >= len(items):
        raise HTTPException(status_code=404, detail=f"Item ID {item_id} is not found")

    return items[item_id]  # 指定されたIDの商品を返す
=======
>>>>>>> origin/main

# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
<<<<<<< HEAD
    category: str = Form(...), #カテゴリーを受け取る
    image: UploadFile = File(...), # 画像を受け取る
    db: sqlite3.Connection = Depends(get_db),
):
    if not name or not category:
        raise HTTPException(status_code=400, detail="name and category is required")
    
    image_name = save_image(image) #画像を保存、ハッシュ名を取得
    new_item = Item(name=name, category=category, image_name=image_name)
    save_item(new_item)  # JSONファイルに保存
    insert_item(new_item)
    return AddItemResponse(**{"message": f"item received: {name}, {category}, {image_name}"})
=======
    category: str = Form(...), #カテゴリーも受け取れるように追記
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    new_item = Item(name=name, category=category)
    save_item(new_item)  # JSONファイルに保存できるようにする
    insert_item(new_item)
    return AddItemResponse(**{"message": f"item received: {name}"})
>>>>>>> origin/main


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=404, detail=f"Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)



def insert_item(new_item):
    # STEP 4-1: add an implementation to store an item
    pass