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

class Item(BaseModel):
    name: str
    category: str

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
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    
    # insert items to items.json
    insert_item(Item(name=name, category=category))

    return AddItemResponse(**{"message": f"item received: {name}"})


class GetItemsResponse(BaseModel):
    items: List[Item]

# STEP 4-3: Retrieve product list
@app.get("/items", response_model=GetItemsResponse)
def get_items():
    with open(ITEMS_FILE_PATH, "r") as f:
        data = json.load(f)
    
    items = []
    for i in data["items"]:
        items.append(i)
    
    return GetItemsResponse(items=items)


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
    # open items.json
    with open(ITEMS_FILE_PATH, "r") as f:
        data = json.load(f)

    # add new item
    new_item = {"name": item.name, "category": item.category}
    data["items"].append(new_item)

    # write to items.json
    with open(ITEMS_FILE_PATH, "w") as f:
        json.dump(data, f, ensure_ascii=False, indent=4)
