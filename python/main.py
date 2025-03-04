import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends,File, UploadFile, Form
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

@app.get("/items") #4-3getエンドポイント作成
def get_items():
    with open('items.json', 'r') as json_open:
        json_load = json.load(json_open)
    return json_load

@app.get("/items/{item_id}") #4-5
def get_items_by_id(item_id: int):
    with open('items.json', 'r') as json_open:
        json_load = json.load(json_open)
    item = json_load["items"][item_id]
    return item



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
    image_content = image.file.read() 
    image_hash = hashlib.sha256(image_content).hexdigest()
    image_path =  images / f"{image_hash}.jpg"

    with open(image_path, "wb") as img_file:
        img_file.write(image_content)

    insert_item(Item(name=name,category=category,image_name = f"{image_hash}.jpg"))
    return AddItemResponse(**{"message": f"item received: name:{name},category:{category},image_name: {image_hash}.jpg"})


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
    image_name :str

def insert_item(item: Item):
    #4-2
    # 読み込む
    with open('items.json', 'r') as json_open:
        json_load = json.load(json_open)  

    # 新しいデータを追加
    json_load["items"].append(item.dict())

    # 書き込む
    with open('items.json', 'w') as json_open:
        json.dump(json_load, json_open, indent=4)
