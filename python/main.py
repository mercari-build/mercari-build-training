import os
import logging
import pathlib
import json
import hashlib
import sqlite3
from fastapi import FastAPI, Form, UploadFile, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

conn=sqlite3.connect(dbname)
cur=conn.cursor()


cur.execute(CREATE TABLE items (
    id INTEGER PRIMARY KEY,
    name TEXT,
    category TEXT,
    image_name TEXT
);)

DB_FILE = "db/mercari.sqlite3"


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


def save_image(file,filename):
    with open(images / filename, "wb") as image:
        image.write(file)

def save_item_db(name, category, image_name):
    cur.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)", (name, category, image_name))
    conn.commit()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

items_list=[]
@app.post("/items")
def add_item(name: str = Form(...), category:str=Form(...), image:UpladFile = File(...)):
    logger.info(f"Receive item: {name}, category: {category}, image: {image}")
    
    file_content = image.file.read()
    hash_value = hashlib.sha256(file_content).hexdigest()
    image_filename = f"{hash_value}.jpg"
    save_image(file_content, image_filename)

    save_item_db(name, category, image_filename)
    return {"message": f"item received: {name},category:{category}","image_name": image_filename}

cur.execute("CREATE TABLE category (id INTEGER PRIMARY KEY, name TEXT)")

@app.get("/items")
def get_items():
    cur.execute("SELECT name, category, image_name FROM items")
    items=cur.fetchall()
    cur.execute("SELECT category_id FROM items INNER JOIN category ON items.category_id=category.id")
    return {"items":items}


@app.get("/image/{image_name}")
async def get_image(image_name):
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/items/{item_id}")
def get_item(item_id: int= Path(..., title="The ID of the item to get")):
    items_data=load_item()
     existing_items = items_data.get("items", [])

    if item_id < len(existing_items):
        item = existing_items[item_id-1]
        return item

@app.get("/search/{search_item}")
def search_item(search_item:str):
    cur.execute("SELECT name, category, image_name FROM items WHERE name LIKE ?",("%"+search_item+"%",))
    items= cur.fetchall()
    return {"items":items}

cur.close()
conn.close()