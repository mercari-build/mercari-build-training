import os
import json
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
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
items_json = pathlib.Path(__file__).parent.resolve() / "items.json"


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
    # Ensure the images directory exists
    images.mkdir(exist_ok=True)
    
    # Ensure the items.json file exists with initial structure
    if not items_json.exists():
        with open(items_json, 'w') as f:
            json.dump({"items": []}, f)


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
    name: str
    category: str
    image_name: Optional[str] = None


class ItemResponse(BaseModel):
    items: List[Item]


class AddItemResponse(BaseModel):
    message: str


# STEP 4-2: Implementation to store an item
def insert_item(item: Item):
    try:
        # Read existing items
        if items_json.exists():
            with open(items_json, 'r') as f:
                data = json.load(f)
        else:
            data = {"items": []}
        
        # Add the new item
        data["items"].append(item.dict())
        
        # Write back to the file
        with open(items_json, 'w') as f:
            json.dump(data, f, indent=2)
            
    except Exception as e:
        logger.error(f"Error inserting item: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to save item: {str(e)}")


# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: Optional[UploadFile] = File(None),
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
    insert_item(item)
    
    return AddItemResponse(**{"message": f"item received: {name}"})


# STEP 3: Implement GET /items to get the list of items
@app.get("/items", response_model=ItemResponse)
def get_items():
    try:
        # Read items from the JSON file
        if items_json.exists():
            with open(items_json, 'r') as f:
                data = json.load(f)
            return data
        else:
            return {"items": []}
    except Exception as e:
        logger.error(f"Error retrieving items: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to retrieve items: {str(e)}")


# STEP 5: Implement GET /items/{item_id} to get a specific item
@app.get("/items/{item_id}")
def get_item(item_id: int):
    try:
        # Read items from the JSON file
        if items_json.exists():
            with open(items_json, 'r') as f:
                data = json.load(f)
            
            # Check if item_id is valid
            if 0 <= item_id < len(data["items"]):
                return data["items"][item_id]
            else:
                raise HTTPException(status_code=404, detail=f"Item with ID {item_id} not found")
        else:
            raise HTTPException(status_code=404, detail="No items found")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error retrieving item: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to retrieve item: {str(e)}")


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