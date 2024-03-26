import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException, UploadFile, File
import hashlib
from typing import List, Optional
from fastapi.responses import FileResponse 
from fastapi.middleware.cors import CORSMiddleware
import sqlite3 

app = FastAPI()
UPLOAD_DIR = "images"
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

def connect():
    con = sqlite3.connect("items.db")
    cursor = con.cursor()
    return con, cursor

con, cursor = connect()

cursor.execute("CREATE TABLE IF NOT EXISTS category (id INTEGER PRIMARY KEY, name TEXT)")
cursor.execute("CREATE TABLE IF NOT EXISTS items (name TEXT, category_id INTEGER, image TEXT)")

# cursor.execute("UPDATE items SET category_id = (SELECT id FROM categories WHERE categories.name = items.name)")

con.commit()
con.close()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    try:
        con,cursor = connect()
        logger.info(f"Received item: {name} in category: {category}")
        cursor.execute("INSERT INTO category (name) VALUES (?)", (category,))
        category_id = cursor.lastrowid
        logger.info(f"Inserted category: {category} with ID: {category_id}")
        image_data = image.file.read()
        image_hash = hashlib.sha256(image_data).hexdigest()

        image_path = os.path.join(UPLOAD_DIR, f"{image_hash}.jpg")
        with open(image_path, "wb") as file:
            file.write(image_data)
        
        cursor.execute("INSERT INTO items (name, category_id, image) VALUES (?,?,?)", (name, category_id, f"{image_hash}.jpg"))
        con.commit()
        con.close()
        return {"message": f"Item received: {name} in category: {category} with image uploaded"}
    except Exception as e:
        print(e)
        raise HTTPException(status_code=500, detail="Internal Server Error")

@app.get("/items")
async def get_items():
    try:
        con,cursor = connect()
        cursor.execute("SELECT items.name, category.name as category FROM items JOIN category ON items.category_id = category.id")
        items = cursor.fetchall()
        
        formatted_items = [{"name": item[0], "category": item[1]} for item in items]
        
        return {"items": formatted_items}
    
    except Exception as e: 
        print(e)
        raise HTTPException(status_code=500, detail="Internal Server Error")
    finally:
        if cursor:
            cursor.close()
        if con:
            con.commit()
            con.close()
@app.get("/items/{item_id}")
async def get_item(item_id: int):
    try:
        con, cursor = connect()

        cursor.execute("""
            SELECT items.name, category.name as category 
            FROM items 
            JOIN category ON items.category_id = category.id 
            WHERE items.id = ?
        """, (item_id,))

        item = cursor.fetchone()

        print(f"Item fetched from database: {item}")  

        if item is None:
            raise HTTPException(status_code=404, detail="Item not found")

        formatted_item = {"name": item[0], "category": item[1]}

        con.close()

        return formatted_item

    except Exception as e:
        print(f"Error occurred: {e}") 
        raise HTTPException(status_code=500, detail="Internal Server Error")

@app.get("/image/{image_name}")
async def get_image(image_name):
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/search")
def search_items(keyword: str):
    try:
        con, cursor = connect()

        cursor.execute("""
            SELECT items.name, category.name as category 
            FROM items 
            JOIN category ON items.category_id = category.id 
            WHERE items.name LIKE ? OR category.name LIKE ?
        """, ('%' + keyword + '%', '%' + keyword + '%'))

        items = cursor.fetchall()
        formatted_items = [{"name": item[0], "category": item[1]} for item in items]

        con.close()

        return {"search_results": formatted_items}

    except Exception as e:
        print(e)
        raise HTTPException(status_code=500, detail="Internal Server Error")

  

