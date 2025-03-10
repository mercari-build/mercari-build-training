import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, File, UploadFile
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
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

   # insert_item(Item(name=name))
   # return AddItemResponse(**{"message": f"item received: {name}"})

    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    if not image:
        raise HTTPException(status_code=400, detail="category is required")
    content = await image.read()
    hashedimg = hashlib.sha256(content).hexdigest()
    image_path = images/ f"{hashedimg}.png"
    with open(image_path, "wb") as a:
        a.write(content)

    insert_item(Item(name=name, category=category, image=hashedimg + ".jpg"))
    return AddItemResponse(**{"message": f"item received: {name}"}) 


# STEP 4-3: return items for GET /items .
@app.get("/items")
def get_items():
    if not os.path.exists("items.json"):
        return []

    with open("items.json", "r") as f:
        items = json.load(f)

    return items

# STEP 4-5 #new
@app.get("/items/{item_id}")
def get_item_by_id(item_id: int):
    #item_idint = int(item_id)
    if not os.path.exists("items.json"):
        return []

    with open("items.json", "r") as f:
        alldata = json.load(f)
    ###added
    if not isinstance(alldata, list):
        return {"error":"Invalid data format in items.json"}

    item= alldata["items"][item_id -1] 
    return item

# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name: str):
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
    category:str 
    #image
    image: str


def insert_item(item: Item):
    # STEP 4-2: add an implementation to store an item
   # pass
    item_dict = {"name": item.name, "category": item.category, "image": item.image}

    if os.path.exists("items.json"):
        with open("items.json", "r") as f:
            try:
                items = json.load(f)
            except json.JSONDecodeError:
                items = []
    else:
        items = []

    items.append(item_dict)

    with open("items.json", "w") as f:
        json.dump(items, f, indent=2)