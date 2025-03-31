import os
import json
import hashlib
import logging
import pathlib
import sqlite3
from fastapi import FastAPI, Form, HTTPException, UploadFile, File, Path, Query, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

# Database Configuration
BASE_DIR = pathlib.Path(__file__).parent
DB_FILE = BASE_DIR / "db" / "mercari.sqlite3"
IMAGE_DIR = BASE_DIR / "images"
IMAGE_DIR.mkdir(exist_ok=True)

# Logging Configuration
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

# FastAPI Initialization
app = FastAPI()
origins = [os.getenv("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# Database Dependency
def get_db():
    """Returns a database connection."""
    conn = sqlite3.connect(DB_FILE, check_same_thread=False)  # Allow access from multiple threads
    conn.row_factory = sqlite3.Row
    try:
        yield conn
    finally:
        conn.close()

# Database Initialization Function
def initialize_db():
    """Ensure the database and tables exist."""
    conn = sqlite3.connect(DB_FILE)
    cursor = conn.cursor()

    # Create tables if not exists
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )
    """)
    
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER NOT NULL,
            image_name TEXT,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        )
    """)

    conn.commit()
    conn.close()

# Call the DB initialization on app startup
initialize_db()

# Item Schema
class Item(BaseModel):
    name: str
    category: str
    image_name: str = None

# Insert Item into Database
def insert_item(item: Item, image_name: str, db: sqlite3.Connection):
    cursor = db.cursor()

    # Insert category if not exists
    cursor.execute("INSERT OR IGNORE INTO categories (name) VALUES (?)", (item.category,))
    cursor.execute("SELECT id FROM categories WHERE name = ?", (item.category,))
    category_id = cursor.fetchone()["id"]

    # Insert item
    cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (item.name, category_id, image_name),
    )
    db.commit()

# Validate Image Type
def validate_image_type(image: UploadFile):
    if not image.filename.lower().endswith((".jpg", ".jpeg", ".png")):
        raise HTTPException(status_code=400, detail="Only JPG, JPEG, or PNG images are allowed.")

# Add Item Endpoint
@app.post("/items")
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(None),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name.strip():
        raise HTTPException(status_code=400, detail="Name cannot be empty")

    image_name = None
    if image:
        validate_image_type(image)
        image_bytes = await image.read()
        image_hash = hashlib.sha256(image_bytes).hexdigest()
        image_name = f"{image_hash}.jpg"
        with open(IMAGE_DIR / image_name, "wb") as f:
            f.write(image_bytes)

    insert_item(Item(name=name, category=category), image_name, db)
    return {"message": f"Item added: {name}"}

# Get All Items
@app.get("/items")
async def get_items(db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute(
        """
        SELECT items.id, items.name, categories.name AS category, items.image_name
        FROM items 
        JOIN categories ON items.category_id = categories.id
        """
    )
    items = cursor.fetchall()
    logger.debug(f"Fetched items: {items}")
    return {"items": [dict(item) for item in items]}

# Get Specific Item
@app.get("/items/{item_id}")
async def get_item(item_id: int, db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute(
        """
        SELECT items.id, items.name, categories.name AS category, items.image_name
        FROM items 
        JOIN categories ON items.category_id = categories.id
        WHERE items.id = ?
        """,
        (item_id,),
    )
    item = cursor.fetchone()

    if item is None:
        raise HTTPException(status_code=404, detail="Item not found")
    return dict(item)

# Search Items
@app.get("/search")
async def search_items(keyword: str = Query(..., description="Keyword to search for items"), db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute(
        """
        SELECT items.id, items.name, categories.name AS category, items.image_name 
        FROM items 
        JOIN categories ON items.category_id = categories.id
        WHERE items.name LIKE ? OR categories.name LIKE ?
        """,
        (f"%{keyword}%", f"%{keyword}%"),
    )
    items = cursor.fetchall()
    logger.debug(f"Search results: {items}")
    return {"items": [dict(item) for item in items]}

# Get Image
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    image_path = IMAGE_DIR / image_name
    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        raise HTTPException(status_code=404, detail="Image not found")
    return FileResponse(image_path)

# Root Endpoint
@app.get("/")
def read_root():
    return {"message": "Hello, world!"}
