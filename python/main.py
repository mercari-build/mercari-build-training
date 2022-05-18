import os
import logging
import pathlib
import sqlite3
import hashlib
import shutil

from sqlite3 import Error
from fastapi import FastAPI, Form, UploadFile, HTTPException
from fastapi.params import File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

DB_PATH = '../db/mercari.sqlite3'

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get('FRONT_URL', 'http://localhost:3000')]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# connect to db
conn = sqlite3.connect(DB_PATH)
c = conn.cursor()

# create table
c.execute("""CREATE TABLE IF NOT EXISTS items (
                id INTEGER PRIMARY KEY,
                name STRING,
                category_id INTEGER,
                image_filename STRING,
                FOREIGN KEY(category_id) REFERENCES category(category_id)
          )""")

c.execute("""CREATE TABLE IF NOT EXISTS category (
               category_id INTEGER PRIMARY KEY,
               name STRING UNIQUE
          )""")

conn.commit()
conn.close()


# Hash the image
def hash_image(filename):
    hashed_str = hashlib.sha256(str(filename).replace('.jpg', '').encode('utf-8')).hexdigest() + '.jpg'
    return hashed_str


# function to search items by name/id
def get_specific_items(name=None, id=None):
    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        sql = """SELECT items.id,
        items.name,
        category.name as category,
        items.image_filename
        FROM items INNER JOIN category
        ON items.category_id =category.category_id"""
        conn.row_factory = sqlite3.Row
        if name:
            suffix = " WHERE items.name=(?)"
            c.execute(sql + suffix, (name.lower(),))

        if id:
            suffix = " WHERE items.id=(?)"
            c.execute(sql + suffix, (id,))
        r = [dict((c.description[i][0], value)
                  for i, value in enumerate(row)) for row in c.fetchall()]
        conn.commit()
        conn.close()
        return {'items': r} if r else {'message': 'Item not found m(_ _)m'}
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}


# API PART

@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.post("/items")
async def add_one_item(name: str = Form(...),
                       category: str = Form(...),
                       image_filename: UploadFile = File(...)):

    #save the uploaded file
    filename = image_filename.filename
    hashed_filename = hash_image(filename)
    save_path = images / hashed_filename

    with open(save_path, 'wb') as buffer:
        shutil.copyfileobj(image_filename.file, buffer)

    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute("INSERT OR IGNORE INTO category (name) VALUES (?)", (category,))
        c.execute("SELECT category_id FROM category WHERE name=(?)", (category,))
        category_id = c.fetchone()[0]
        c.execute("INSERT INTO items(name,category_id, image_filename) VALUES (?,?,?)",
                  (name, category_id, hashed_filename))
        conn.commit()
        conn.close()
        logger.info('add successfully!')
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}

    result = {"name": name, "category": category, "image_filename": hashed_filename}
    logger.info(f"Receive item: {result}")
    return {"message": f"item received: {name}"}


@app.get("/items")
async def get_all_items():
    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        conn.row_factory = sqlite3.Row
        sql = """SELECT items.id,
        items.name,
        category.name as category,
        items.image_filename
        FROM items INNER JOIN category
        ON items.category_id =category.category_id"""
        c.execute(sql)
        r = [dict((c.description[i][0], value)
                  for i, value in enumerate(row)) for row in c.fetchall()]
        conn.commit()
        conn.close()

        return {'items': r} if r else {'message': 'No data found m(_ _)m'}
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}


@app.get("/search/")
async def read_item(keyword: str):
    return get_specific_items(name=keyword)


@app.get("/items/{item_id}")
async def read_item(item_id: int):
    return get_specific_items(id=item_id)


@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.info(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
