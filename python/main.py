import os
import logging
import pathlib
import json
import hashlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
from typing import List

FILENAME = pathlib.Path(__file__).parent.resolve() / "items.json"

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"

images.mkdir(exist_ok=True)

def get_db():
   conn = sqlite3.connect(db)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    conn = sqlite3.connect(db)
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL COLLATE NOCASE,
            category TEXT NOT NULL,
            image_name TEXT NOT NULL
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
logger.level = logging.DEBUG
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
    items: list

# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name or not category: 
        raise HTTPException(status_code=400, detail="name and category are required")

    if not image.filename.endswith(".jpg") or image.content_type != "image/jpeg":
        raise HTTPException(status_code=400, detail="only image files with .jpg are allowed")

    image_bytes = await image.read()
    hashed_filename = hashlib.sha256(image_bytes).hexdigest() + ".jpg"
    image_path = images / hashed_filename
    
    with open(image_path, "wb") as f:
        f.write(image_bytes) # save the image

    cursor = db.cursor()
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_id = cursor.fetchone()

    if category_id is None:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
        category_id = cursor.lastrowid
    else:
        category_id = category_id[0]
        
    new_item = Item(name=name, category=category, image_name=hashed_filename)
    insert_item(new_item, db)
    logger.debug("Inserting item: %s", new_item)
    
    return AddItemResponse(
        message=f"item received: {name}",
        items=[{"name": name, "category": category, "image_name": hashed_filename}]
        )


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # Create image path
    image_path = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = images / "default.jpg"

    return FileResponse(image_path)


class Item(BaseModel):
    name: str
    category: str
    image_name: str

def insert_item(item: Item):
    cursor = db.cursor()
   cursor.execute(
        "INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
        (item.name, item.category, item.image_name)
    )
   db.commit()

def read_items():
    cursor = db.cursor()
    cursor.execute("""
    SELECT items.name, categories.name AS category, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id
""")
    rows = cursor.fetchall()
    items = [{"name": row[0], "category": row[1], "image_name": row[2]} for row in rows]
    return items

@app.get("/items/{item_id}", response_model=Item)
def get_items(item_id: int, db: sqlite3.Connection = Depends(det_db)):
    cursor = db.cursor()
    cursor.execute("""
        SELECT items.name, categories.name AS category, items.image_name
        FROM items
        JOIN categories ON items.category_id = categories.id
        WHERE items.id = ?
    """, (item_id,))
    row = cursor.fetchone()

    if row is None:
        raise HTTPException(status_code=404, detail="Item not found")

    return Item(name=row[0], category=row[1], image_name=row[2])

@app.get("/search")
def search_items(keyword: str, db: sqlite3.Connection = Depends(get_db)):
    keyword = keyword.strip()
    
    cursor = db.cursor()
    cursor.execute("""
        SELECT items.name, categories.name AS category, items.image_name
        FROM items
        JOIN categories ON items.category_id = categories.id
        WHERE items.name LIKE ?
    """, (f"%{keyword}%",))
    
    rows = cursor.fetchall()
    
    if not rows:
        logger.debug("No items found matching the search criteria.")
        
    items = [{"name": row[0], "category": row[1], "image_name": row[2]} for row in rows]
    
    return {"items": items}

@app.on_event("startup")
def startup_db():
    print("Attempting to connect to the database:", db)
    try:
        conn = sqlite3.connect(db)
        print("Database connection successful!")
        conn.close()
    except sqlite3.Error as e:
        print(f"Error connecting to the database: {e}")
