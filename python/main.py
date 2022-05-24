from operator import itemgetter
import json
import sqlite3
import pathlib
import hashlib
import os
import logging
import pathlib
from unicodedata import category, name
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from numpy import record

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "image"
db_path = str(pathlib.Path(os.path.dirname(__file__)
                           ).parent.resolve() / "db" / "mercari.sqlite3")
origins = [os.environ.get('FRONT_URL', 'http://localhost:3000')]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)


def write_to_json(data):
    with open("items.json", "r+") as file:
        file_data = json.load(file)
        file_data["items"].append(data)
        file.seek(0)
        json.dump(file_data, file, indent=4)


def get_from_json():
    with open("items.json", "r") as file:
        file_data = json.load(file)
        return file_data


def hash_image(image):
    filename = pathlib.Path(__file__).parent.resolve() / "image" / image
    with open(filename, "rb") as f:
        encoded_image = f.read()
        hashed_image = hashlib.sha256(encoded_image).hexdigest()
    return hashed_image + ".jpg"


@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.get("/items")
def getItems():
    conn = sqlite3.connect(db_path)
    curr = conn.cursor()
    curr.execute(
        "SELECT items.name,category.name,items.image FROM items INNER JOIN category ON items.category_id = category.id")
    data = curr.fetchall()
    data_list = list()
    for i in range(len(data)):
        temp = dict(zip(["name", "category", "image"], [
                    data[i][0], data[i][1], data[i][2]]))
        data_list.append(temp)
    conn.commit()
    conn.close()
    return {"items": data_list}


@app.get("/items/{id}")
def get_items_id(id):
    conn = sqlite3.connect(db_path)
    curr = conn.cursor()
    curr.execute(
        "SELECT items.name,category.name AS category,items.image FROM items INNER JOIN category ON items.category_id = category.id WHERE items.id = ?", (id,))
    data = curr.fetchall()
    result = dict(zip(["name", "category", "image"], [
                  data[0][0], data[0][1], data[0][2]]))
    return result


@app.get("/search")
def search_item(keyword: str):
    conn = sqlite3.connect(db_path)
    curr = conn.cursor()
    query = '%' + keyword + '%',
    curr.execute(
        "SELECT items.name,category.name AS category,items.image FROM items INNER JOIN category ON items.category_id = category.id WHERE items.name LIKE ?", (query))
    data = curr.fetchall()
    data_list = list()
    for i in range(len(data)):
        temp = dict(zip(["name", "category", "image"], [
                    data[i][0], data[i][1], data[i][2]]))
        data_list.append(temp)
    conn.commit()
    conn.close()
    return {"items": data_list}


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):
    logger.info(f"Receive item: {name}")
    conn = sqlite3.connect(db_path)
    curr = conn.cursor()
    hashed_image = hash_image(image)
    curr.execute(
        "SELECT id FROM category WHERE name = ?", (category, ))
    category_id = curr.fetchall()
    if len(category_id) != 0:
        curr.execute("INSERT INTO items(name,category_id,image) VALUES(?,?,?)",
                     (name, category_id[0][0], hashed_image))
        conn.commit()
    else:
        curr.execute("INSERT INTO category(name) VALUES(?)", (category, ))
        conn.commit()
        curr.execute("SELECT id FROM category WHERE name = ?", (category, ))
        category_id = curr.fetchall()
        curr.execute("INSERT INTO items(name, category_id, image) VALUES(?,?,?)",
                     (name, category_id[0][0], hashed_image))
        conn.commit()
    conn.close()
    return {"message": f"item received: {name}"}


@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path
    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(
            status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
