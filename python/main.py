import os
import logging
import pathlib
import json
import hashlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager

FILENAME = pathlib.Path(__file__).parent.resolve() / "items.json"

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"

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
    items: list

# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name or not category: 
        raise HTTPException(status_code=400, detail="name and category are required")

    if not image.filename.endswith(".jpg") or image.content_type != "image/jpeg":
        raise HTTPException(status_code=400, detail="only image files with .jpg are allowed")

    image_bytes = await image.read()
    hashed_filename = hashlib.sha256(image_bytes).hexdigest() + ".jpg"
    image_path = images / hashed_filename
    
    with open(image_path, "wb") as f:
        f.write(image_bytes) # save the image
        
    new_item = Item(name=name, category=category, image_name=hashed_filename)
    insert_item(new_item)
    
    return AddItemResponse(message=f"item received: {name}", items=[{"name": name, "category": category, "image_name": hashed_filename}])


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # Create image path
    image_path = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = images / "default.jpg"

    return FileResponse(image_path)


class Item(BaseModel):
    name: str
    category: str
    image_name: str

def insert_item(item: Item):
    # STEP 4-2: add an implementation to store an item
    if os.path.exists(FILENAME):  
        with open(FILENAME, "r", encoding="utf-8") as file:
            try: 
                data = json.load(file)
            except json.JSONDecodeError:
                data = {"items": []}          
    else: 
        data = {"items": []}
        
    data["items"].append(item.dict())
    
    with open(FILENAME, "w", encoding = "utf-8") as file:
        json.dump(data, file, indent = 2, ensure_ascii = False)

def read_items():
    if not FILENAME.exists():
        return []
    with open(FILENAME, "r", encoding="utf-8") as file:
        try:
            data = json.load(file)
            return data.get("items", [])
        except json.JSONDecodeError:
            return []

@app.get("/items/{item_id}", response_model=Item)
def get_items(item_id: int):
    items = read_items()

    if item_id < 0 or item_id >= len(items):
        raise HTTPException(status_code=404, detail="Item not found")
    
    return items[item_id]
    else:
        data = {"items": []}

    return data
