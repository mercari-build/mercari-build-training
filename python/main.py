import json
import os
import logging
import pathlib
import logging
import hashlib
import sqlite3
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from contextlib import asynccontextmanager
from typing import List, Optional


app = FastAPI()
items_file = "items.json"

logging.basicConfig(level=logging.DEBUG)

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"

# Ensure images directory exists
images_dir = pathlib.Path("images")
images.mkdir(exist_ok=True)

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
    pass


@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

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


class HelloResponse(BaseModel):
    message: str


@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})


class AddItemResponse(BaseModel):
    message: str

class Item(BaseModel):
    name: str
    category: str
    image_name: str

def insert_item(item: Item):
    # STEP 4-1: add an implementation to store an item
    try:
        with open("items.json", "r") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}

        # Append new item
    data["items"].append({"name": item.name, "category": item.category, "image_name": item.image_name})

        # Write back to the file
    with open("items.json", "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

# add_item is a handler to add a new item for POST /items .
@app.post("/items")
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: Optional[UploadFile] = File(None),
    db: sqlite3.Connection = Depends(get_db),
):    
    image_name = ""  # Default to empty string if no image is uploaded
    if image is not None:
        file_bytes = await image.read()
        image_hash = hashlib.sha256(file_bytes).hexdigest()
        image_name = f"{image_hash}.jpg"
        image_path = images_dir / image_name
        with open(image_path, "wb") as f:
            f.write(file_bytes)#

    if not name or not category:
        raise HTTPException(status_code=400, detail="name is required")

    insert_item(Item(name=name,category=category,image_name=image_name))
    return AddItemResponse(**{"message": f"item received: {name}"})


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name:str):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
        
@app.get("/items")
def get_all_items():
    try:
        with open(items_file, "r") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}
    
    return data


@app.get("/items/{item_id}")
def get_items(item_id: int):
    try:
        with open(items_file, "r") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}
    
    items = data.get("items", [])
    
    return items[item_id - 1]