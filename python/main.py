import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json 
import hashlib


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
items = pathlib.Path(__file__).parent.resolve() / "items.json" 

def get_items():
    if not items.exists():
        with open(items, "w") as f:
            items.parent.mkdir(parents=True, exist_ok=True)
            json.dump({"items": []}, f)
    return items
 
def get_db():
    if not db.exists():
        yield

    conn = sqlite3.connect(db)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()
    
# STEP 5-1: set up the database connection
def setup_database():
    with sqlite3.connect(db) as conn:
        conn.execute("""
                     CREATE TABLE IF NOT EXISTS items (
                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                         name TEXT NOT NULL,
                         category_id INTEGER NOT NULL,
                         image_name TEXT NOT NULL,
                         FOREIGN KEY (category_id) REFERENCES categories(id)
                         
                     );
        """)
        
        conn.execute("""
                     CREATE TABLE IF NOT EXISTS categories(
                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                         name TEXT NOT NULL 
                     );
        """)
        conn.commit()

# Added new function 
def get_item_by_id(item_id: int):
    # with open(get_items(), "r") as f:
    #     data = json.load(f)
    # if not (0 <= item_id < len(data["items"])):
    #     raise HTTPException(status_code=404, detail="Item not found")
    # return data["items"][item_id - 1]
    
    with sqlite3.connect(db) as conn:
        conn.row_factory =sqlite3.Row
        row = conn.execute(
            "SELECT id, name, category, image_name FROM items WHERE id=?",
            (item_id,)
        ).fetchone()
    if row is None:
        raise HTTPException(status_code=404, detail="Item not found")
    
    return dict(row)

# Added new function
def upload_image(image: UploadFile):
    image_data = image.file.read()
    hashed_value = hashlib.sha256(image_data).hexdigest()
    image_name = f"{hashed_value}.jpg"
    with open(images/image_name, "wb") as image_file:
        image_file.write(image_data)
    return image_name

@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

logger = logging.getLogger("uvicorn")
# optional
logger.setLevel(logging.DEBUG)
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
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    db: sqlite3.Connection = Depends(get_db),
    image: UploadFile = Form(...)
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    if not image:
        raise HTTPException(status_code=400, detail="image is required")
    
    image_name = upload_image(image)
    item = Item(name=name, category=category, image_name=image_name)
    insert_item(item)
    return {"message": f"item received: {name}"}

@app.get("/items")
def get_all_items():
    # with open(get_items(), "r") as f:
    #     return json.load(f)
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row
        rows = conn.execute("""
            SELECT items.id, items.name, categories.name AS category, items.image_name FROM items
            JOIN categories ON items.category_id = categories.id;
        """).fetchall()
    
    items_data = [dict(row) for row in rows]
    return {"items": items_data}
    
    
# info endpoint
@app.get("/items/{item_id}")
def get_item_info(item_id: int):
    # return get_item_by_id(item_id)
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row
        row = conn.execute("""
                           SELECT items.id, items.name, categories.name AS category, items.image_name FROM items
                           JOIN categories ON items.category_id = categories.id WHERE items.id= ?;
                           """, (item_id,)).fetchone()
        if row is None:
            raise HTTPException(status_code=404, detail= "Item not found")
        return dict(row)    
    


# get_image is a handler to return an image for GET /images/{filename} .
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


class Item(BaseModel):
    name: str
    category: str
    image_name: str 


def insert_item(item: Item):
    with sqlite3.connect(db) as conn:
        conn.row_factory = sqlite3.Row 

        conn.execute("INSERT OR IGNORE INTO categories (name) VALUES (?)", (item.category,))
        conn.commit()

        cursor = conn.execute("SELECT id FROM categories WHERE name = ?", (item.category,))
        category_id = cursor.fetchone()["id"]  
        
        conn.execute(
            "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
            (item.name, category_id, item.image_name)
        )
        conn.commit()


# Step5-2 search function
@app.get("/search")
def search_items(keyword:str):
    with sqlite3.connect(db) as conn:
        conn.row_factory=sqlite3.Row
        rows = conn.execute(
            "SELECT items.id, items.name, categories.name AS category, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?",
            (f"%{keyword}%",)
        ).fetchall()
        items_data = [dict(row) for row in rows]
        return {"items": items_data}