import os
import logging
import pathlib
import json
import logging
from fastapi import FastAPI, Form, HTTPException, Depends
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


# STEP 5-1: set up the database connection
def setup_database():
    pass


@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield

logging.basicConfig(level=logging.INFO)  # INFO レベルのログを有効化
logger = logging.getLogger(__name__)  # ロガーを作成

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
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    insert_item(Item(name=name))
    return AddItemResponse(**{"message": f"item received: {name}"})


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


def insert_item(item: Item):
    # STEP 4-1: add an implementation to store an item
    class Item(BaseModel):
        name:str

    file_path = "item.json"

    try:
        with open(file_path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}
    
    data["items"].append({"name": item.name})

    with open(file_path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=4)

    return {"message": "Item added successfully"}

import json
import hashlib
import shutil
from fastapi import FastAPI, UploadFile, File

app = FastAPI()


from pydantic import BaseModel

class Item(BaseModel):
    name: str
    category: str

@app.post("/items")
async def create_item(
    name: str = Form(...), 
    category: str = Form(...), 
    image: UploadFile = File(...)
):

    image_bytes = await image.read()
    hashed_filename = hashlib.sha256(image_bytes).hexdigest() + ".jpg"
    
    image_path = f"images/{hashed_filename}"
    with open(image_path, "wb") as buffer:
        buffer.write(image_bytes)

    item = {
        "name": name,
        "category": category,
        "image_name": hashed_filename
    }

    file_path = "item.json"

    with open(file_path, "r+", encoding="utf-8") as f:
        f.seek(0)  # ファイルの先頭に移動
        try:
            data = json.load(f)
        except json.JSONDecodeError:
            data = {"items": []}  # もし空ならデフォルトをセット

        except (FileNotFoundError, json.JSONDecodeError):
            data = {"items": []}

    data["items"].append({
    "name": name,
    "category": category,
    "image_name": hashed_filename
})

    with open(file_path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=4)

    return {"message": "Item added successfully", "image_name": hashed_filename}



@app.get("/items", response_model=dict)
def get_items():
    file_path = "item.json"  

    try:
        with open(file_path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        data = {"items": []}  

    print(data)
    return data

@app.get("/items/{item_id}")
def get_item(item_id: int):
    file_path = "item.json"

    try:
        with open(file_path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return {"error": "No items found"}

    # item_id が範囲内かチェック
    if item_id < 0 or item_id >= len(data["items"]):
        return {"error": "Item not found"}

    return data["items"][item_id]

@app.get("/images/{image_name}")
async def get_image(image_name: str):
    image_path = f"images/{image_name}"
    
    if not os.path.exists(image_path):
        logger.warning(f"Image not found: {image_path}")  # ログ出力！
        return {"error": f"Image not found: {image_path}"}
    
    return FileResponse(image_path)


