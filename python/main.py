import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException, Depends,UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import hashlib


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"

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
    image_name: str = ""

def load_items():
    if items_file.exists():
        with open(items_file, "r") as f:
            return json.load(f)
    return {"items": []}


def save_items(items):
    with open(items_file, "w") as f:
        json.dump({"items": items}, f, indent=4)

def insert_item(item: Item):
    if items_file.exists():
        with open(items_file, "r") as f:
            data = json.load(f)
            items = data.get("items", [])
    else:
        items = []
    
    items.append(item.dict())
    save_items(items)

# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(None),
    db: sqlite3.Connection = Depends(get_db)
):
    if not name or not category:
        raise HTTPException(status_code=400, detail="name and category are required")

    image_name = ""
    if image:
        contents = image.file.read()
        image_hash = hashlib.sha256(contents).hexdigest()
        image_name = f"{image_hash}.jpg"
        image_path = images / image_name
        with open(image_path, "wb") as f:
            f.write(contents)
    
    item = Item(name=name, category=category, image_name=image_name)
    insert_item(item)
    return AddItemResponse(**{"message": f"item received: {name}"})

# get_items is a handler to return the list of items for GET /items .
@app.get("/items")
def get_items():
    return load_items()

# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/images/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)