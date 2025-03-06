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

# TODO: Create items.json if it does not exist

app = FastAPI(lifespan=lifespan)

logger = logging.getLogger("uvicorn")
# change logger level to display varying levels of logs displayed on console
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
    # Form(None) for non-required fields
    image: str | None = Form(None),
    
    db: sqlite3.Connection = Depends(get_db),
):
    print(name, category, image)
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    if not category:
        raise HTTPException(status_code=400, detail="category is required")

    if image != None:
        # print(image)
        image_hash = hash_image(image)

    insert_item(Item(name=name, category=category, image=image))
    return AddItemResponse(**{"message": f"item received: name={name}, category={category}, image={image_hash}"})

@app.get("/items")
def get_items():
    with open('items.json', 'r') as json_file:
        try:
            data = json.load(json_file)
        except json.JSONDecodeError:
            data = {}
    return data

@app.get("/items/{item_id}")
def get_item(item_id):
    # check if item_id is a valid integer
    try:
        item_id = int(item_id)
    except:
        raise HTTPException(status_code=400, detail="Item ID is not a valid number")
    
    with open('items.json', 'r') as json_file:
        try:
            data = json.load(json_file)
        except json.JSONDecodeError:
            data = {}


    #TODO: error handling - items might be empty
    if not data:
        raise HTTPException(Status_code=400, detail="Item ID does not exist")
    items = data['items']
    item = items[item_id - 1]
    return item


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        # changed logger level from debug to higher log level (info, error etc.)
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)


class Item(BaseModel):
    name: str
    category: str
    image: str | None


def insert_item(item: Item):
    # STEP 4-2: add an implementation to store an item
    new_data = {
        "name": item.name,
        "category": item.category,
        "image": item.image
    }

    with open('items.json', 'r') as json_file:
        try:
            data = json.load(json_file)
        except json.JSONDecodeError:
            data = {}

        # print(data)
    
    if 'items' not in data:
        data['items'] = []

    data['items'].append(new_data)

    # print(data)

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