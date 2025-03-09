import json
import os
import logging
import pathlib
import logging
import hashlib
import sqlite3 
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from contextlib import asynccontextmanager
from typing import List, Optional
from starlette.responses import FileResponse
from flask import Flask, request, jsonify
from flask_cors import CORS

app = FastAPI()
logging.basicConfig(level=logging.DEBUG)
DB_PATH = "/Users/takaho.ysz/db/mercari.sqlite3"

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"

# Ensure images directory exists
images_dir = pathlib.Path("images")
images.mkdir(exist_ok=True)

def get_db():
    conn = sqlite3.connect(DB_PATH, check_same_thread = False)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    db_path = str(db)
    print(f"Using database file: {db_path}") 
    
    db.parent.mkdir(parents = True, exist_ok= True)#dbがあることを確認
    conn = sqlite3.connect(db_path, check_same_thread = False)
    cursor = conn.cursor()
    
    #カテゴリーのテーブル
    
    cursor.execute("""
                   CREATE TABLE IF NOT EXISTS categories(
                       id INTEGER PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL
                       )
                       """)
    
    cursor.execute("""
                   CREATE TABLE IF NOT EXISTS items(
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       name TEXT NOT NULL,
                       category_id TEXT NOT NULL,
                       image_name TEXT NOT NULL,
                       FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
                       )
                       """)
    conn.commit()
    conn.close()


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

def insert_item(db: sqlite3.Connection, item:Item):
    # STEP 4-1: add an implementation to store an item
    cursor=db.cursor()
    cursor.execute(item.name,item.category,item.image_name)
    db.commit()
    
    """
    try:
        with open("items.json", "r") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}

        # Append new item
    data["items"].append(item.model_dump())
        # Write back to the file
    with open("items.json", "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    """

# add_item is a handler to add a new item for POST /items .
@app.post("/items",response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: Optional[UploadFile] = File(None),
    db: sqlite3.Connection = Depends(get_db),
):    
    image_name = ""  # Default to empty string if no image is uploaded
        
    if image is not None:
        file_bytes = await image.read()
        image_hash = hashlib.sha256(file_bytes).hexdigest()
        image_name = f"{image_hash}.jpg"
        image_path = images_dir / image_name
        with open(image_path, "wb") as f:
            f.write(file_bytes)#

    if not name or not category:
        raise HTTPException(status_code=400, detail="name is required")
        
    cursor = db.cursor()
    
    #カテゴリーが存在しなかった場合
    cursor.execute("INSERT OR IGNORE INTO categories (name) VALUES (?)",(category,))
    db.commit()
    
    #カテゴリーid
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
    result = cursor.fetchone()
    
    if result is None:
        raise HTTPException(status_code = 400, detail="Category not found")
    category_id = result[0]

    
    cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (name, category_id, image_name)
        )
    db.commit()
    
    return AddItemResponse(**{"message": f"item received: {name}"})


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name:str):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
        
@app.get("/items")
def get_all_items(db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute("""
        SELECT items.id, items.name, categories.name as category, items.image_name
        FROM items
        JOIN categories ON items.category_id = categories.id
    """)
    items = cursor.fetchall()
    return{"items":[dict(item)for item in items]}
    
@app.get("/search")
def search_items(keyword: str = Query(...),db:sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    search_query = """
    SELECT items.name, categories.name AS category, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id
    WHERE items.name LIKE ?
    """
    cursor.execute(search_query, ('%' + keyword + '%',))
    results = cursor.fetchall()
    
    items = [{"name": row["name"], "category": row["category"], "image_name":row["image_name"]} for row in results]
    return {"items":items}
    
    
@app.get("/items/{item_id}")
def get_items(item_id: int, db : sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute(item.id)
    item = cursor.fetchall()

    """
    try:
        with open(items_file, "r") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}
    
    items = data.get("items", [])
    
    return items[item_id - 1]
    """

        
