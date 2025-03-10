import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, File, UploadFile, File, Query
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json
import hashlib
import sqlite3
from typing import Dict, List


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
#when executing in cmd, go to path until python folder 
#

def get_db():
    if not db.exists():
        yield

    conn = sqlite3.connect(db, check_same_thread=False)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn  
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    conn = sqlite3.connect(db)
    cursor = conn.cursor() 
    sql_file = pathlib.Path(__file__).parent.resolve() / "db" / "items.sql"
    with open(sql_file, "r") as f:
        cursor.executescript(f.read())
    conn.commit()
    conn.close()


@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG #step 4-6
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


# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    if not image:
        raise HTTPException(status_code=400, detail="category is required")
    content = await image.read()
    hashedimg = hashlib.sha256(content).hexdigest()
    image_path = images/ f"{hashedimg}.jpg"
    with open(image_path, "wb") as a:
        a.write(content)

    insert_item_db(Item(name=name, category=category, image=hashedimg), db)
    return AddItemResponse(**{"message": f"item received: {name}"})


# STEP 4-3: return items for GET /items .
@app.get("/items")
def get_items(db: sqlite3.Connection = Depends(get_db)):
    return get_items_from_database(db) 


#step 5:1 GET /items
def get_items_from_database(db: sqlite3.Connection)-> Dict[str, List[Dict[str, str]]]:
    cursor = db.cursor() 
    # Query the Items table
    # STEP 5-8 change to get category from category table not just id
    query = """
    SELECT items.name, category.name AS category, items.image_name
    FROM items
    JOIN category
    ON category_id = category.id 
"""
##view get items or insert items!!
    cursor.execute(query)
    rows = cursor.fetchall()
    items_list = [{"name": name, "category": category, "image_name": image_name} for name, category, image_name in rows]
    result = {"items": items_list}
    cursor.close()

    return result

def get_items_from_database_by_id(id: int, db: sqlite3.Connection) -> Dict[str, List[Dict[str,str]]]:
    cursor = db.cursor()
    # Query the Items table, modify for STEP5-8
    query = """" 
    "    SELECT items.name, category.name AS category, image_name 
    FROM items
    JOIN category
    ON category_id = category.id
    WHERE items.id = ?
    """

    cursor.execute(query, (id,))
    rows = cursor.fetchall()
    items_list = [{"name": name, "category": category, "image_name": image_name} for name, category, image_name in rows]
    result = {"items": items_list}
    cursor.close()

    return result

# STEP 4-5 #new
@app.get("/items/{item_id}")
def get_item_by_id(item_id: int, db: sqlite3.Connection = Depends(get_db)):
    all_items = get_items_from_database(db)["items"]  # Get all items
    
    if item_id <= 0 or item_id > len(all_items):  # Check if ID is valid
        raise HTTPException(status_code=404, detail="Item not found")

    return all_items[item_id - 1]  # Adjust to zero-based index


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/search")
def search_keyword(keyword: str = Query(...), db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    query = """
    SELECT items.name AS name, category.name AS category, image_name 
    FROM items 
    JOIN category
    ON category_id = category.id
    WHERE items.name LIKE ?
    """

    #specifying what the query will do upon searching, with LIKE

    pattern = f"%{keyword}%" #search keyword in any part of the item name
    cursor.execute(query, (pattern,))
    rows = cursor.fetchall()
    items_list = [{"name": name, "category": category, "image_name": image_name} for name, category, image_name in rows]
    result = {"items": items_list}
    cursor.close()
    return result

class Item(BaseModel):
    name: str
    category:str 
    #image
    image: str


def insert_item_db(item: Item, db: sqlite3.Connection) -> int:
    cursor = db.cursor()
    query = """
        INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?);
    """
    #query into category table
    query_category = "SELECT id FROM category WHERE name = ?"
    cursor.execute(query_category, (item.category,))
    rows = cursor.fetchone()

    if rows is None:
        insert_query_category = "INSERT INTO category (name) VALUES (?)" 
        cursor.execute(insert_query_category, (item.category,))
        category_id = cursor.lastrowid
    else:
        category_id = rows[0]
                                             
    print(f"Inserting into DB: {item.name}, {item.category}, {item.image}")  # Debugging
    cursor.execute(query, (item.name, category_id, item.image))  # Use category_id here
    db.commit()
   