import os
import logging
import json
import hashlib
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"

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

#added
    
# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def insert_item(item: Item):
    logger.info(f"Inserting item: {item.dict()}")
    try:
        with open(items_file, "r") as f:
            data = json.load(f)
    except FileNotFoundError:
        data = {"items": []}
    except json.decoder.JSONDecodeError as e:
        logger.error(f"Error decoding JSON: {e}")
        return #Exit the function, or raise an exception.

    data["items"].append(item.dict())

    try:
        with open(items_file, "w") as f:
            json.dump(data, f, indent=4)
        logger.info(f"Item inserted successfully: {item.dict()}")
    except Exception as e:
        logger.error(f"Error writing to JSON: {e}")


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
    try:
        with open(items_file, "r") as f:
            data = json.load(f)
    except FileNotFoundError:
        data = {"items": []}

    data["items"].append(item.dict()) #add the item as a dictionary.

    with open(items_file, "w") as f:
        json.dump(data, f, indent=4) #write the data back to the file with indentation.

   
