import os
import sqlite3
import json
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import FileResponse , JSONResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "images"
db_path = pathlib.Path(__file__).parent.resolve() / "/Users/yurainagaki/mercari/mercari-build-training/db/items.db"
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

def create_table():
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER NOT NULL,
            image_name TEXT NOT NULL,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        )
    """)
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE
        )
    """)
    conn.commit()
    conn.close()
create_table()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
async def add_item(name: str = Form(...), category_name: str = Form(...),image: UploadFile = File(...)):
    contents = await image.read()
    hash_sha256 = hashlib.sha256(contents).hexdigest()
    image_filename = f"{hash_sha256}.jpg"
    image_path = images / image_filename
    with open(image_path, "wb") as file:
        file.write(contents)
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category_name,))
    category = cursor.fetchone()

    if not category:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category_name,))
        conn.commit()
        category_id = cursor.lastrowid
    else:
        category_id = category[0]
    
    cursor.execute("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)", (name, category_id, image_filename))
    conn.commit()
    conn.close()

    item = {"name": name, "category_id": category_id,"image_name": image_filename}
    logger.info(f"Receive item: {item}")

    # save_item(item)
    return {"message": f"Item received: {item}","image_name": image_filename}

@app.get("/categories")
def get_categories():
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("SELECT id, name FROM categories")
    categories = [{"id": row[0], "name": row[1]} for row in cursor.fetchall()]
    conn.close()
    return {"categories": categories}

@app.get("/items")
def get_items():
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("""
    SELECT items.id, items.name, categories.name AS category_name, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id
    """)
    items = [{"id": row[0], "name": row[1], "category_name": row[2], "image_name": row[3]} for row in cursor.fetchall()]
    conn.close()
    return {"items": items}


@app.get("/items/{item_id}")
def get_items(item_id: int):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("""
    SELECT items.id, items.name, categories.name AS category_name, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id
    WHERE items.id = ?
    """, (item_id,))

    item = cursor.fetchone()
    conn.close()

    if item is None:
        raise HTTPException(status_code=404, detail="Item not found")
    return {"id": item_id, "name": item[1], "category_name": item[2], "image_name": item[3]}

@app.get("/search")
async def search_items(keyword: str):
    if not keyword:
        raise HTTPException(status_code=400, detail="Keyword must not be empty")
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    search_query = f"%{keyword}%"
    cursor.execute("SELECT id, name, category_id, image_name FROM items WHERE name LIKE ?", (search_query,))
    items = [{"id": row[0], "name": row[1], "category_name": row[2], "image_name": row[3]} for row in cursor.fetchall()]
    conn.close()
    return JSONResponse(content={"items": items})

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
    
    return FileResponse(image)


#uvicorn main:app --reload --log-level debug --port NUMBEROFPORT