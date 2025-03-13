import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends
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


# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image_name: str = Form(None),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    
    if not category:
        raise HTTPException(status_code=400, detail="category is required")

    if not image_name or image_name.strip() == "":
        image_name = "images/default.jpg"

    
    image_hash = hash_image(image_name)+".jpg"
    insert_item(Item(name=name, category=category ,image_name=image_hash))
    return AddItemResponse(**{"message": f"item received:name= {name}, category= {category}, image={image_hash}"})


@app.get("/items")
def get_items():
    with open('items.json', 'r') as json_file:
        try:
            data = json.load(json_file)
        except json.JSONDecodeError:
            data = {}
    return data


@app.get("/item/{item_id}")
def get_item(item_id):
    with open('items.json', 'r') as json_file:
        try:
            data = json.load(json_file)
        except json.JSONDecodeError:
            data = {}

    if not data:
        raise HTTPException(Status_code=400, detail="Item does not exist")
    
    item = data[int(item_id) - 1]

    return item


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
    # STEP 4-2: add an implementation to store an item
    # pass

    with open('items.json', 'r') as json_file:
        try:
            data = json.load(json_file)
        except json.JSONDecodeError:
            data = []

    new_data = {
        "name": item.name,
        "category": item.category,
        "image_name":item.image_name
    }

    # if 'items' not in data:
    #     data = []

    data.append(new_data)

    with open('items.json', 'w') as json_file:
        json.dump(data, json_file, indent=4)


def hash_image(image):
    with open(image, "rb") as f:
        try:
            image_bytes = f.read()
        except:
            raise HTTPException(status_code=400, detail="Image not found")
        image_hash = hashlib.sha256(image_bytes).hexdigest()
    return image_hash