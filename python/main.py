import os
import json
import hashlib
import logging
import sqlite3
from pathlib import Path
from fastapi import FastAPI, Form, HTTPException, UploadFile, File, Path, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from pydantic import BaseModel

# Constants
DB_FILE = "mercari.sqlite3"
IMAGE_DIR = "images"
os.makedirs(IMAGE_DIR, exist_ok=True)  # Ensure image directory exists

# Function to connect to database
def get_db():
    conn = sqlite3.connect(DB_FILE)
    conn.row_factory = sqlite3.Row
    return conn

# Define Item schema
class Item(BaseModel):
    name: str
    category: str
    image_name: str = None

# Initialize FastAPI
app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# ✅ List a new item (Insert into database)
@app.post("/items")
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(None),
):
    conn = get_db()
    cursor = conn.cursor()

    # Check if category exists
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_data = cursor.fetchone()

    # If category doesn't exist, insert it
    if category_data is None:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
        category_id = cursor.lastrowid
    else:
        category_id = category_data["id"]

    # Process image
    image_name = None
    if image:
        image_bytes = image.file.read()
        image_hash = hashlib.sha256(image_bytes).hexdigest()
        image_name = f"{image_hash}.jpg"
        with open(f"{IMAGE_DIR}/{image_name}", "wb") as f:
            f.write(image_bytes)

    # Insert item into database
    cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (name, category_id, image_name),
    )
    conn.commit()
    conn.close()

    return {"message": f"Item added: {name}"}

# ✅ Get all items (from database)
@app.get("/items")
def get_items():
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("""
        SELECT items.id, items.name, categories.name AS category, items.image_name 
        FROM items 
        JOIN categories ON items.category_id = categories.id
    """)
    items = cursor.fetchall()
    conn.close()
    return {"items": [dict(item) for item in items]}

# ✅ Get specific item details (from database)
@app.get("/items/{item_id}")
def get_item(item_id: int = Path(..., description="The ID of the item to retrieve")):
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("""
        SELECT items.id, items.name, categories.name AS category, items.image_name 
        FROM items 
        JOIN categories ON items.category_id = categories.id
        WHERE items.id = ?
    """, (item_id,))
    item = cursor.fetchone()
    conn.close()

    if item is None:
        raise HTTPException(status_code=404, detail="Item not found")

    return dict(item)

# ✅ Search for an item (by name or category)
@app.get("/search")
def search_items(keyword: str = Query(..., description="Keyword to search for items")):
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("""
        SELECT items.id, items.name, categories.name AS category, items.image_name 
        FROM items 
        JOIN categories ON items.category_id = categories.id
        WHERE items.name LIKE ? OR categories.name LIKE ?
    """, (f"%{keyword}%", f"%{keyword}%"))
    
    items = cursor.fetchall()
    conn.close()
    return {"items": [dict(item) for item in items]}

# ✅ Fetch an image by filename
@app.get("/images/{image_name}")
async def get_image(image_name: str):
    image_path = Path(f"{IMAGE_DIR}/{image_name}")
    if image_path.exists():
        return FileResponse(image_path)
    raise HTTPException(status_code=404, detail="Image not found")

# ✅ Root endpoint
@app.get("/")
def read_root():
    return {"message": "Hello, world!"}

# ✅ Logger Configuration
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)
