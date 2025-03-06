import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, Depends, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json
from pathlib import Path
from typing import List


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"


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

class Item(BaseModel):
    name: str
    category: str
    image_name: str


class HelloResponse(BaseModel):
    message: str

class AddItemResponse(BaseModel):
    message: str

class GetItemsResponse(BaseModel):
    items: List[Item]

class GetItemResponse(BaseModel):
    item: Item


@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})


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

    #load image
    image_contents = await image.read() 

    # Hashing images with SHA-256
    sha256 = hashlib.sha256(image_contents).hexdigest()
    image_name = f"{sha256}.jpg"  

    # Save images to images directory
    image_path = images / image_name
    with open(image_path, "wb") as f:
        f.write(image_contents)
    
    # insert items to items.json
    insert_item(Item(name=name, category=category, image_name=image_name))

    return AddItemResponse(**{"message": f"item received: {name}"})


# STEP 4-3: Retrieve product list
@app.get("/items", response_model=GetItemsResponse)
def get_items():
    # if no json data
    if not ITEMS_FILE_PATH.exists():
        return GetItemsResponse(items=[])
    
    with open(ITEMS_FILE_PATH, "r") as f:
        data = json.load(f)
    
    items = []
    for item in data["items"]:
        items.append(item)
    
    return GetItemsResponse(items=items)

#STEP 4-5: Return product details
@app.get("/items/{item_id}", response_model=GetItemResponse)
def get_item(item_id: int):
    # if no json data
    if not ITEMS_FILE_PATH.exists():
        raise HTTPException(status_code=404, detail="item is not found")
    
    with open(ITEMS_FILE_PATH, "r") as f:
        data = json.load(f)

    # check if item_id exists
    if item_id > len(data["items"]) and item_id > 0:
        raise HTTPException(status_code=404, detail="Item not found")

    # get item with item_id 
    item = data["items"][item_id - 1]
    
    return GetItemResponse(item=item)
    

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


# items.json's pash
ITEMS_FILE_PATH = Path("items.json")

def insert_item(item: Item):
    # STEP 4-2: add an implementation to store an item
    # if no json data
    if not ITEMS_FILE_PATH.exists():
        with open(ITEMS_FILE_PATH, "w") as f:
            json.dump({"items": []}, f)

    # open items.json
    with open(ITEMS_FILE_PATH, "r", encoding="utf-8") as f:
        data = json.load(f)  

    # add new item
    new_item = item.model_dump()
    data["items"].append(new_item)

    # write to items.json
    with open(ITEMS_FILE_PATH, "w") as f:
        json.dump(data, f, ensure_ascii=False, indent=4)
