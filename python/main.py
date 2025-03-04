import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import hashlib
import shutil
from fastapi import FastAPI, Form, File, UploadFile
import json
from typing import Union


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"


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


@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})


class AddItemResponse(BaseModel):
    message: str

class Item(BaseModel):
    name: str
    category: str
    image_name: str


IMAGES_DIR ="images"
os.makedirs(IMAGES_DIR, exist_ok =True)
# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile =File(...), 
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    
    #画像データを読み込んでSHA-256ハッシュを作成
    image_bytes =image.file.read()
    image.file.seek(0) #ファイルポインタをリセット　#これいる？なんのため
    hashed_filename = hashlib.sha256(image_bytes).hexdigest() +".jpg"

    #画像を保存
    image_path = os.path.join(IMAGES_DIR, hashed_filename)
    with open(image_path, "wb") as buffer:
        buffer.write(image_bytes)
    
    #データを返す（データベースには未保存）
    #return{
       # "name": name,
       # "category": category,
        #"image_name": hashed_filename
    #}

    insert_item(Item(name=name, category=category, image_name=hashed_filename))
    return AddItemResponse(**{"message": f"item received: {name},{category}, {hashed_filename}"})


# JSON ファイルのパス
file_path = 'items.json'

@app.get("/items")
def get_items():
    """登録された商品一覧を取得するエンドポイント"""
    if not os.path.exists(file_path):
        return {"items": []}

    with open(file_path, 'r', encoding='utf-8') as file:
        try:
            items_data = json.load(file)
        except json.JSONDecodeError:
            return {"items": []}

    return items_data  # {"items": [...]} の形式で返す

@app.get("/items/{item_id}")
def get_item(item_id: int) -> Union[Item, dict]:
    """
    1商品の詳細情報を取得するエンドポイント
    - `item_id`: 何個目に登録された商品かを示すID（0から始まるインデックス）
    """
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail="No items found")

    with open(file_path, 'r', encoding='utf-8') as file:
        try:
            items_data = json.load(file)
        except json.JSONDecodeError:
            raise HTTPException(status_code=500, detail="JSON file is corrupted")

    items = items_data.get("items", [])

    if item_id < 0 or item_id >= len(items):
        raise HTTPException(status_code=404, detail="Item not found")

    return items[item_id]



# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)


class Item(BaseModel):
    name: str
    category: str
    image_name: str

def insert_item(item: Item):
    # 新しい商品情報の辞書を作成
    new_item = {
        "name": item.name,
        "category": item.category, 
        "image_name": item.image_name}
    
     

    # ファイルが存在するかチェック
    if not os.path.exists(file_path):
        # ファイルが存在しない場合は空のデータを用意
        items_data = {"items": []}
    else:
        # JSON ファイルを読み込む
        with open(file_path, 'r', encoding='utf-8') as file:
            try:
                items_data = json.load(file)
            except json.JSONDecodeError:
                # JSON のフォーマットが壊れていた場合は初期化
                items_data = {"items": []}



    # items キーが存在しない場合は新しく作成
    if 'items' not in items_data:
        items_data['items'] = []

    # 新しい商品を追加
    items_data['items'].append(new_item)

    # JSON ファイルを更新
    with open('items.json', 'w', encoding='utf-8') as file:
        json.dump(items_data, file, ensure_ascii=False, indent=4)

    # STEP 4-2: add an implementation to store an item
    pass









