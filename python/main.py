import os
import logging
import pathlib
import json
import hashlib
import shutil
import sqlite3
from fastapi import FastAPI, Form, HTTPException, UploadFile, File, Query
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
       CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    images_folder = "images"
    image_content = image.file.read()
    image_hash = hashlib.sha256(image_content).hexdigest() + ".jpg"
    image_path = os.path.join(images_folder, image_hash)
    with open(image_path, "wb") as image_file:
        image_file.write(image_content)
    logger.info(f"Receive item: {name} {category} {image_hash}")

    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute('SELECT id FROM categories WHERE name = ?', (category,))
    existing_category = cursor.fetchone()
    if existing_category:
        category_id = existing_category[0]
    else:
        cursor.execute('INSERT INTO categories (name) VALUES (?)', (category,))
        conn.commit()
        category_id = cursor.lastrowid
    cursor.execute('INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)', (name, category_id, image_hash))
    conn.commit()
    conn.close()
    return {"message": f"item received: {name}"}

@app.get("/items")
def get_items():
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute('SELECT name, category_id, image_name FROM items')
    items_data = [{"name": name, "category": get_category_name(category_id), "image_name": image_name} for name, category_id, image_name in cursor.fetchall()]
    conn.close()
    return {"items": items_data}

def get_category_name(category_id):
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute('SELECT name FROM categories WHERE id = ?', (category_id,))
    category_name = cursor.fetchone()[0]
    conn.close()
    return category_name

@app.get("/items/{item_id}")
def get_item_id(item_id: int):
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute('SELECT name, category_id, image_name FROM items WHERE id = ?', (item_id,))
    item_data = cursor.fetchone()
    name, category_id, image_name = item_data
    category_name = get_category_name(category_id)
    item = {"name": name, "category": category_name, "image_name": image_name}
    return item
    
@app.get("/search")
def search_items(keyword: str = Query(...)):
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute('SELECT items.name, categories.name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?', ('%' + keyword + '%',))
    filtered_items = [{"name": name, "category": category} for name, category in cursor.fetchall()]
    
    return {"items": filtered_items}

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