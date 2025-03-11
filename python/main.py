import os
import pathlib
import hashlib
import sqlite3
from fastapi import FastAPI, Form, UploadFile, File, HTTPException, Query
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

# Define file paths
BASE_DIR = pathlib.Path(__file__).parent.resolve()
IMAGES_DIR = BASE_DIR / "images"
DB_PATH = BASE_DIR / "db" / "mercari.sqlite3"

# Ensure necessary directories exist
IMAGES_DIR.mkdir(parents=True, exist_ok=True)

# FastAPI App
app = FastAPI()

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

# Response models
class Item(BaseModel):
    name: str
    category: str
    image_path: str  

class GetItemsResponse(BaseModel):
    items: list[Item]

# Hash function for image names
def hash_image(file_data):
    sha256 = hashlib.sha256()
    sha256.update(file_data)
    return sha256.hexdigest()

# 🟢 GET All Items (With Category Name)
@app.get("/items", response_model=GetItemsResponse)
def get_all_items():
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    try:
        cursor.execute("""
            SELECT items.name, categories.name AS category, items.image_name 
            FROM items 
            JOIN categories ON items.category_id = categories.id
        """)
        items = [{"name": row[0], "category": row[1], "image_path": row[2]} for row in cursor.fetchall()]
        return GetItemsResponse(items=items)
    finally:
        conn.close()

# 🟢 GET Search Items by Keyword
@app.get("/search", response_model=GetItemsResponse)
def search_items(keyword: str = Query(..., title="Keyword to search in item names")):
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    try:
        cursor.execute("""
            SELECT items.name, categories.name AS category, items.image_name 
            FROM items 
            JOIN categories ON items.category_id = categories.id
            WHERE items.name LIKE ?
        """, (f"%{keyword}%",))
        items = [{"name": row[0], "category": row[1], "image_path": row[2]} for row in cursor.fetchall()]

        if not items:
            raise HTTPException(status_code=404, detail="No matching items found")

        return GetItemsResponse(items=items)
    finally:
        conn.close()

# 🟢 POST Add New Item
@app.post("/items")
async def add_item(
    name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)
):
    if not name or not category or not image:
        raise HTTPException(status_code=400, detail="Name, category, and image are required")

    # Read and hash image
    image_data = await image.read()
    hashed_image_name = hash_image(image_data) + ".jpg"
    image_path = IMAGES_DIR / hashed_image_name

    # Save the image
    with open(image_path, "wb") as img_file:
        img_file.write(image_data)

    # Insert into database
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    try:
        # Get or create category
        cursor.execute("SELECT id FROM categories WHERE name = ?", (category,))
        category_id = cursor.fetchone()
        if not category_id:
            cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
            category_id = cursor.lastrowid
        else:
            category_id = category_id[0]

        # Insert item
        cursor.execute("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
                       (name, category_id, hashed_image_name))
        conn.commit()
    finally:
        conn.close()

    return {"message": f"Item '{name}' added successfully."}

# 🟢 GET Image by Name
@app.get("/images/{image_name}")
async def get_image(image_name: str):
    image_path = IMAGES_DIR / image_name

    if not image_path.exists():
        raise HTTPException(status_code=404, detail="Image not found")

    return FileResponse(image_path)

