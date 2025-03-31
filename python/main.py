import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json
import hashlib
import shutil

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
    print("Yes")
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
    # STEP6 のためにコメントアウトしておく（念のため）
    # if not image:
    #     raise HTTPException(status_code=400, detail="image is required")
    image_bin = await image.read()
    hashed_image = hashlib.sha256(image_bin).hexdigest()
    
    insert_item(Item(name=name, category=category, image=hashed_image))

    with open('images/' + hashed_image + '.jpg', 'wb') as f:
        f.write(image_bin)

    return AddItemResponse(**{"message": f"item received: {name}"})

@app.get("/items")
def get_items():
    with open('items.json') as f:
        d_update = json.load(f)
    return d_update

@app.get("/items/{item_id}")
def get_items(item_id:int):
    with open('items.json') as f:
        d_update = json.load(f)
    return d_update['items'][item_id]

@app.get("/items/{item_id}")
def get_items(item_id:int):
    with open('items.json') as f:
        d_update = json.load(f)
    return d_update['items'][item_id]

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


class Item(BaseModel):
    name:str
    category:str
    image:str

def insert_item(item: Item):
    # STEP 4-1: add an implementation to store an item
    with open('items.json') as f:
        d_update = json.load(f)

    d = {'name' : item.name, 'category': item.category, 'image_name':item.image}
    d_update['items'].append(d)

    with open('items.json', 'w') as f:
        json.dump(d_update, f, indent=2)
