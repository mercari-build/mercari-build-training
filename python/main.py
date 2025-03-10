import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, Depends, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from contextlib import asynccontextmanager
import sqlite3

# Define the path to the images & item SQL file
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

def setup_database():
    
    with sqlite3.connect(db) as conn:

    # itemsテーブルの変更（category → category_id に変更）
        conn.execute(
        """CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER NOT NULL,
            image_name TEXT NOT NULL,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        );"""
        )
        # カテゴリテーブルの作成（変更点）
        conn.execute(
        """CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        );"""
    )
    
        conn.commit() #commit()を呼び出してSQLの変更をデータベースに保存

# Added new function
def upload_image(image: UploadFile):
    image_data = image.file.read()
    hashed_value = hashlib.sha256(image_data).hexdigest()
    image_name = f"{hashed_value}.jpg"
    with open(images/image_name, "wb") as image_file:
        image_file.write(image_data)
    return image_name

@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield

app = FastAPI(lifespan=lifespan)

logger = logging.getLogger("uvicorn")
logger.setLevel = (logging.DEBUG)
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

class AddItemResponse(BaseModel):
    message: str

class Item(BaseModel):
    name: str
    category: str
    image_name: str



def insert_item(item: Item):
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row
        
        conn.execute("INSERT OR IGNORE INTO categories (name) VALUES (?)", (item.category,))
        conn.commit()

        cursor = conn.execute("SELECT id FROM categories WHERE name = ?", (item.category,))
        category_id = cursor.fetchone()["id"]

        conn.execute(
            "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
            (item.name, category_id, item.image_name)
        )
        conn.commit()

def fetch_all_items():
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row
        rows = conn.execute("""
                            SELECT items.id, items.name, categories.name AS category, items.image_name FROM items
                            JOIN categories ON items.category_id = categories.id;
        """).fetchall()
        
    return [dict(row) for row in rows]


def fetch_item_by_id(item_id: int):
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row
        row = conn.execute("""
                           SELECT items.id, items.name, categories.name AS category, items.image_name FROM items
                           JOIN categories ON items.category_id = categories.id WHERE items.id = ?;
        """, (item_id,)).fetchone()
    
    if row is None:
        raise HTTPException(status_code=404, detail="Item not found")
    
    return dict(row)

def search_items_in_db(keyword: str):
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row
        rows = conn.execute("""
                            SELECT items.id, items.name, categories.name AS category, items.image_name FROM items
                            JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?;
        """, (f"%{keyword}%",)).fetchall()
        
    return [dict(r) for r in rows]

@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = Form(...),
    db: sqlite3.Connection = Depends(get_db)
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    if not image:
        raise HTTPException(status_code=400, detail="image is required")
        
    image_name = upload_image(image)
    item = Item(name=name, category=category, image_name=image_name)
    insert_item(item)
    return {"message": f"item received: {name}"}


@app.get("/items")
def get_all_items():
    items_data = fetch_all_items()  
    return {"items": items_data}

@app.get("/items/{item_id}")
def get_item_info(item_id: int):
    item_data = fetch_item_by_id(item_id)  
    return item_data

@app.get("/search")
def search_items(keyword: str):
    items_data = search_items_in_db(keyword) 
    return {"items": items_data}

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