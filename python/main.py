import os
import logging
import pathlib
from fastapi import FastAPI, Form, File, UploadFile, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager


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

import json
JSON_FILE = 'items.json'

def read_json_file():
    if not os.path.exists(JSON_FILE):
        with open(JSON_FILE, 'w') as f:
            json.dump({"items":[]}, f)
        with open(JSON_FILE, 'r') as f:
            return json.load(f)
        
def write_json_file(data):
    with open(JSON_FILE, 'w') as f:
        json.dump(data, f, indent=4)


import hashlib
def hash_image(image_file: UploadFile):
    try:
        image = image_file.file.read()
        hash_value = hashlib.sha256(image).hexdigest()
        hashed_image_name = f"{hash_value}.jpeg"
        hashed_image_path = image / hashed_image_name

        with open(hashed_image_path, 'wb') as f:
            f.write(image)
        return hashed_image_name
    except Exception as e:
        raise RuntimeError(f"An unexpected error occurred: {e}")

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
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    if not image:
        raise HTTPException(status_code=400, detail="image is required")

    hashed_image = hash_image(image)

    insert_item(Item(name=name, category=category, image=hashed_image))
    return AddItemResponse(**{"message": f"item received: {name}"})

@app.get("/items")
def get_items():
    all_data = read_json_file()
    return all_data

@app.get("items/{item_id}")
def get_item_by_id(item_id):
    item_id_int = int(item_id)
    all_data = read_json_file()
    item = all_data["items"] [item_id_int -1]
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
    image: str


def insert_item(item: Item):
    current_data = read_json_file()

    current_data["items"].append({"name": item.name, "category": item.category, "image_name": item.image})

    write_json_file(current_data)
