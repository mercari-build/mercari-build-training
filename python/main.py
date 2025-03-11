import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import hashlib


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"


def get_db():
    if not db.exists():
        yield

    conn = sqlite3.connect(db, check_same_thread=False)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    # items.sqlのファイルパスを定義
    sql_file_path = pathlib.Path(__file__).parent.resolve() / "db" / "items.sql"
    
    # items.sqlファイルが存在するかチェック
    if not sql_file_path.exists():
        raise FileNotFoundError(f"SQL file not found: {sql_file_path}")
    
    conn = sqlite3.connect(db)
    try:
        cursor = conn.cursor()
        with open(pathlib.Path(__file__).parent.resolve() / "db" / "items.sql", "r", encoding="utf-8") as sql_file:
            sql_script = sql_file.read()
        cursor.executescript(sql_script)
        conn.commit()
    finally:
        conn.close()
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


# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),  # category_id ではなく category を受け取る
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    file_data = image.file.read()
    hashed_value = hashlib.sha256(file_data).hexdigest()

    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    # カテゴリIDを取得（なければ作成）
    cursor = db.cursor()
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_row = cursor.fetchone()

    if category_row:
        category_id = category_row["id"]
    else:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
        category_id = cursor.lastrowid
        db.commit()

    # 画像を保存
    image_path = images / f"{hashed_value}.jpg"
    with open(image_path, "wb") as f:
        f.write(file_data)

    # アイテムを追加
    insert_item(Item(name=name, category_id=category_id, image=hashed_value), db)

    return AddItemResponse(**{"message": f"item received: {name}"})


# get_item is a handler to get all items for GET /items .
@app.get("/items")
async def get_items(db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    query = """
    SELECT items.id, items.name, categories.name as category, items.image
    FROM items
    JOIN categories ON items.category_id = categories.id
    """
    cursor.execute(query)
    items = cursor.fetchall()
    return [dict(item) for item in items]



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

# step5-2
@app.get("/search")
async def search_items(keyword: str, db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    query = """
    SELECT items.id, items.name, categories.name as category, items.image
    FROM items
    JOIN categories ON items.category_id = categories.id
    WHERE items.name LIKE ? OR categories.name LIKE ?
    """
    cursor.execute(query, (f"%{keyword}%", f"%{keyword}%"))
    items = cursor.fetchall()
    return [dict(item) for item in items]


# step5-3
class AddCategoryResponse(BaseModel):
    message: str


@app.post("/categories", response_model=AddCategoryResponse)
def add_category(
    name: str = Form(...),
    db: sqlite3.Connection = Depends(get_db),
):
    cursor = db.cursor()
    cursor.execute(
        "INSERT INTO categories (name) VALUES (?)",
        (name,)
    )
    db.commit()
    return AddCategoryResponse(**{"message": f"category received: {name}"})


class Item(BaseModel):
    name: str
    category_id: int 
    image: str  


def insert_item(item: Item, db: sqlite3.Connection):
    # STEP 5-1: add an implementation to store an item
    cursor = db.cursor()
    cursor.execute(
        "INSERT INTO items (name, category_id, image) VALUES (?, ?, ?)",
        (item.name, item.category_id, item.image)
    )
    db.commit()