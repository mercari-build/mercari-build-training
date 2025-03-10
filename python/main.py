import os
import hashlib
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
from typing import List, Optional


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
images = pathlib.Path(__file__).parent.resolve() / "images"
db_dir = pathlib.Path(__file__).parent.resolve() / "db"
db = db_dir / "mercari.sqlite3"


def get_db():
    conn = sqlite3.connect(db, check_same_thread=False)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    # Ensure the images directory exists
    images.mkdir(exist_ok=True)
    
    # Ensure the db directory exists
    db_dir.mkdir(exist_ok=True)
    
    # Initialize the database with the schema
    conn = sqlite3.connect(db)
    cursor = conn.cursor()
    
    # Read and execute SQL schema from items.sql file
    schema_file = db_dir / "items.sql"
    if schema_file.exists():
        with open(schema_file, 'r') as f:
            schema_sql = f.read()
            cursor.executescript(schema_sql)
    
    conn.commit()
    conn.close()



@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
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


class Item(BaseModel):
    id: Optional[int] = None
    name: str
    category: str
    image_name: Optional[str] = None


class Category(BaseModel):
    id: int
    name: str


class ItemResponse(BaseModel):
    items: List[Item]


class AddItemResponse(BaseModel):
    message: str

# Find or create category ID
def get_or_create_category(category_name: str, db_conn):
    cursor = db_conn.cursor()
    
    # Look for existing category
    cursor.execute("SELECT id FROM categories WHERE name = ?", (category_name,))
    result = cursor.fetchone()
    
    if result:
        return result['id']
    
    # Create new category if it doesn't exist
    cursor.execute("INSERT INTO categories (name) VALUES (?)", (category_name,))
    db_conn.commit()
    return cursor.lastrowid


# STEP 4-2: Implementation to store an item
def insert_item(item: Item, db_conn):
    # Get or create category_id
    category_id = get_or_create_category(item.category, db_conn)
    
    cursor = db_conn.cursor()
    cursor.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (item.name, category_id, item.image_name)
    )
    db_conn.commit()
    return cursor.lastrowid

# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: Optional[UploadFile] = File(None),
    db_conn: sqlite3.Connection = Depends(get_db)
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    
    image_name = None
    
    # Handle image upload if provided
    if image:
        # Read image content
        image_content = await image.read()
        
        # Hash the image content using SHA-256
        hash_obj = hashlib.sha256(image_content)
        hashed_value = hash_obj.hexdigest()
        image_name = f"{hashed_value}.jpg"
        
        # Save the image
        image_path = images / image_name
        with open(image_path, "wb") as f:
            f.write(image_content)
    
    # Create and insert the item
    item = Item(name=name, category=category, image_name=image_name)
    
    try:
        insert_item(item, db_conn)
        return AddItemResponse(**{"message": f"item received: {name}"})
    except Exception as e:
        logger.error(f"Error inserting item: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to save item: {str(e)}")


# STEP 3: Implement GET /items to get the list of items
@app.get("/items", response_model=ItemResponse)
def get_items(db_conn: sqlite3.Connection = Depends(get_db)):
    try:
        cursor = db_conn.cursor()
        cursor.execute("""
            SELECT i.id, i.name, c.name as category, i.image_name 
            FROM items i
            JOIN categories c ON i.category_id = c.id
        """)
        items = [dict(item) for item in cursor.fetchall()]
        return {"items": items}
    except Exception as e:
        logger.error(f"Error retrieving items: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to retrieve items: {str(e)}")



# STEP 5: Implement GET /items/{item_id} to get a specific item
@app.get("/items/{item_id}")
def get_item(item_id: int, db_conn: sqlite3.Connection = Depends(get_db)):
    try:
        cursor = db_conn.cursor()
        cursor.execute("""
            SELECT i.id, i.name, c.name as category, i.image_name 
            FROM items i
            JOIN categories c ON i.category_id = c.id
            WHERE i.id = ?
        """, (item_id,))
        item = cursor.fetchone()
        
        if item:
            return dict(item)
        else:
            raise HTTPException(status_code=404, detail=f"Item with ID {item_id} not found")
    except Exception as e:
        logger.error(f"Error retrieving item: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to retrieve item: {str(e)}")


# GET /search to search for items with a keyword
@app.get("/search", response_model=ItemResponse)
def search_items(keyword: str, db_conn: sqlite3.Connection = Depends(get_db)):
    try:
        cursor = db_conn.cursor()
        cursor.execute("""
            SELECT i.id, i.name, c.name as category, i.image_name 
            FROM items i
            JOIN categories c ON i.category_id = c.id
            WHERE i.name LIKE ?
        """, (f"%{keyword}%",))
        items = [dict(item) for item in cursor.fetchall()]
        return {"items": items}
    except Exception as e:
        logger.error(f"Error searching items: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to search items: {str(e)}")

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



