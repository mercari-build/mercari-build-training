import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import hashlib
import shutil
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


# STEP 5-1: set up the database connection (テーブル作成)
def setup_database():
    conn = sqlite3.connect(db) #SQLiteのデータベースに接続
    cursor = conn.cursor() #cursorオブジェクトを作成。cursorはデータベースに対してSQLコマンドを実行するために使われる

    # カテゴリテーブルの作成（変更点）
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL
        );"""
    )

    # itemsテーブルの変更（category → category_id に変更）
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER NOT NULL,
            image_name TEXT NOT NULL,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        );"""
    )
    
    conn.commit() #commit()を呼び出してSQLの変更をデータベースに保存
    conn.close() #データベースとの接続を閉じる。開いたままにするとリソース無駄に消費


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
    id: int
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
    image.file.seek(0) #ファイルポインタをリセット　
    hashed_filename = hashlib.sha256(image_bytes).hexdigest() +".jpg"

    #画像を保存
    image_path = os.path.join(IMAGES_DIR, hashed_filename)
    with open(image_path, "wb") as buffer:
        buffer.write(image_bytes)

    cursor =db.cursor() 
    
    # categories テーブルにカテゴリが存在するか確認
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_row = cursor.fetchone()

    if category_row:
        category_id = category_row["id"]
    else:
        # カテゴリが存在しない場合、新しく追加
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
        category_id = cursor.lastrowid  # 追加したカテゴリの ID を取得

    # items テーブルにデータを保存
    cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (name, category_id, hashed_filename),
    )
    db.commit()
    return AddItemResponse(**{"message": f"item received: {name}, {category}, {hashed_filename}"})

    #データをデータベースに保存
    cursor = db.cursor()
    cursor.execute(
        "INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
        (name, category, hashed_filename),
    )
    db.commit()

    return AddItemResponse(**{"message": f"item received: {name},{category}, {hashed_filename}"})




@app.get("/items")
def get_items(db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute("SELECT * FROM items")
    # JOIN を使ってカテゴリ名を取得（変更点）
    cursor.execute(
        """SELECT items.id, items.name, categories.name as category, items.image_name
           FROM items
           JOIN categories ON items.category_id = categories.id"""
    )
    rows = cursor.fetchall()
    items_list = [{"name": name, "category": category, "image_name": image_name} for name, category, image_name in rows]
    
    
    return {"items": items}
    
    
@app.get("/search")
def search_items(query: str, db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    # JOIN を使ってカテゴリ名での検索も可能に（変更点）
    cursor.execute(
        """SELECT items.id, items.name, categories.name as category, items.image_name
           FROM items
           JOIN categories ON items.category_id = categories.id
           WHERE items.name LIKE ? OR categories.name LIKE ?""",
        (f"%{query}%", f"%{query}%"),
    )

    items = [
        {"id": row["id"], "name": row["name"], "category": row["category"], "image_name": row["image_name"]}
        for row in cursor.fetchall()
    ]

    if not items:
        raise HTTPException(status_code=404, detail="No items found with the given query")

    return {"items": items}
    
    
    
    

#GET/items/{items_id} (データベースから一つの商品を取得)
@app.get("/items/{item_id}")

def get_item(item_id: int, db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    # JOIN を使ってカテゴリ名を取得（変更点）
    cursor.execute(
        """SELECT items.id, items.name, categories.name as category, items.image_name
           FROM items
           JOIN categories ON items.category_id = categories.id
           WHERE items.id = ?""",
        (item_id,),
    )

    row = cursor.fetchone()

    if row is None:
        raise HTTPException(status_code=404, detail="Item not found")

    return {"id": row["id"], "name": row["name"], "category": row["category"], "image_name": row["image_name"]}
    


# get_image is a handler to return an image for GET /images/{filename} .　画像取得エンドポイント
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