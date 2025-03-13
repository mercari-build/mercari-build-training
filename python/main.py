import os
import logging
import pathlib
import json
import hashlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File, Body
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
from typing import List, Optional

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
FILENAME = pathlib.Path(__file__).parent.resolve() / "items.json"

images.mkdir(exist_ok=True)

def get_db():
   if not db.exists():
        setup_database()
      
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

   # Create categories table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL COLLATE NOCASE UNIQUE
        )
    """)

  # Create items table with category_id instead of category
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL COLLATE NOCASE,
            category_id INTEGER NOT NULL,
            image_name TEXT NOT NULL,
            FOREIGN KEY (category_id) REFERENCES categories (id)
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

class ItemCreate(BaseModel):
    name : str
    category: str

@app.post("/items", response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name or not category: 
        raise HTTPException(status_code=400, detail="name and category are required")

   # Image validation 
   if not image.filename.endswith(".jpg") or image.content_type != "image/jpeg":
        raise HTTPException(status_code=400, detail="only image files with .jpg are allowed")

   # Save image
    image_bytes = await image.read()
    hashed_filename = hashlib.sha256(image_bytes).hexdigest() + ".jpg"
    image_path = images / hashed_filename
   
    with open(image_path, "wb") as f:
        f.write(image_bytes)  # Save the image

    cursor = db.cursor()
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_id = cursor.fetchone()

    if category_id is None:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
        category_id = cursor.lastrowid
    else:
        category_id = category_id["id"]
        
    # Insert the item into the items table
    cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (name, category_id, hashed_filename)
    )
    db.commit()

    # Response with item information
    return AddItemResponse(
        message=f"Item '{name}' added successfully!",
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

# New endpoint to get image by item ID
@app.get("/image/id/{item_id}")
async def get_image_by_id(item_id: int, db: sqlite3.Connection = Depends(get_db)):
    # Query the database to get the image_name for this item_id
    cursor = db.cursor()
    cursor.execute("SELECT image_name FROM items WHERE id = ?", (item_id,))
    result = cursor.fetchone()
    
    if not result:
        logger.warning(f"Item not found with ID: {item_id}")
        # Fall back to default image
        image_path = images / "default.jpg"
        if not image_path.exists():
            logger.error("Default image not found!")
            raise HTTPException(status_code=404, detail="Image not found")
        return FileResponse(image_path)
    
    image_name = result["image_name"]
    image_path = images / image_name
    
    if not image_path.exists():
        logger.warning(f"Image not found for item ID {item_id}, using default")
        image_path = images / "default.jpg"
        if not image_path.exists():
            logger.error("Default image not found!")
            raise HTTPException(status_code=404, detail="Image not found")
    else:
        logger.info(f"Serving image for item ID {item_id}: {image_name}")
    
    return FileResponse(image_path)

class Item(BaseModel):
    name: str
    category: str
    image_name: str

def insert_item(item: Item, category_id, db: sqlite3.Connection):
    cursor = db.cursor()
   cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (item.name, category_id, item.image_name)
    )
   db.commit()

def read_items(db: sqlite3.Connection):
    cursor = db.cursor()
    cursor.execute("""
    SELECT items.name, categories.name AS category, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id
""")
    rows = cursor.fetchall()
     items = [{"id": row["id"], "name": row["name"], "category": row["category"], "image_name": row["image_name"]} for row in rows]
    return items

@app.get("/items/{item_id}", response_model=Item)
def get_items(item_id: int, db: sqlite3.Connection = Depends(det_db)):
    cursor = db.cursor()
    cursor.execute("""
    SELECT items.id, items.name, categories.name AS category, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id
    """)
    rows = cursor.fetchall()
    items = [{"id": row["id"], "name": row["name"], "category": row["category"], "image_name": row["image_name"]} for row in rows]
    return {"items": items}

@app.get("/search")
def search_items(keyword: str, db: sqlite3.Connection = Depends(get_db)):
    keyword = keyword.strip()
    
    cursor = db.cursor()
    cursor.execute("""
        SELECT items.id, items.name, categories.name AS category, items.image_name
        FROM items
        JOIN categories ON items.category_id = categories.id
        WHERE items.name LIKE ?
    """, (f"%{keyword}%",))
    
    rows = cursor.fetchall()
    
    if not rows:
        logger.debug("No items found matching the search criteria.")
        
    items = [{"id": row["id"], "name": row["name"], "category": row["category"], "image_name": row["image_name"]} for row in rows]
    
    return {"items": items}

@app.delete("/items/{item_id}")
async def delete_item(item_id: int, db: sqlite3.Connection = Depends(get_db)):
# Find the item in the database
    cursor = db.cursor()
    cursor.execute("SELECT id FROM items WHERE id = ?", (item_id,))
    item = cursor.fetchone()

    if not item:
        raise HTTPException(status_code=404, detail="Item not found")

    # Delete the item from the database
    cursor.execute("DELETE FROM items WHERE id = ?", (item_id,))
    db.commit()

    return {"message": f"Item with ID {item_id} deleted successfully"}
