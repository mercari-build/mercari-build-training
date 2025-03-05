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
items_json_path = pathlib.Path(__file__).parent.resolve() / "items.json"  # items.json のパスを追記


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
    category: str  # カテゴリーも受け取れるように追加

# items.jsonの初期データ読み込み（items.jsonが存在しない場合の対応を）
def load_items():
    if items_json_path.exists():
        with open(items_json_path, "r", encoding="utf-8") as file:
            data = json.load(file)
            return data.get("items", [])
    return []


#item.jsonに商品を追加して保存する関数
def save_item(item: Item):
    items = load_items()
    items.append({"name": item.name, "category": item.category}) #商品を追加する

    with open(items_json_path, "w", encoding="utf-8") as file:
        json.dump({"items": items}, file, indent=2, ensure_ascii=False)  # JSONを保存

@app.get("/items") #登録された商品データをJSON形式で取得
def get_items():
    return {"items": load_items()}


# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...), #カテゴリーも受け取れるように追記
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    new_item = Item(name=name, category=category)
    save_item(new_item)  # JSONファイルに保存できるようにする
    insert_item(new_item)
    return AddItemResponse(**{"message": f"item received: {name}"})


# get_image is a handler to return an image for GET /images/{filename} .
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



def insert_item(new_item):
    # STEP 4-1: add an implementation to store an item
    pass
