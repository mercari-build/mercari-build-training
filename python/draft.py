import os
import logging
import pathlib
import json
import sqlite3
import hashlib
from sqlite3 import Error

from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

DB_PATH = './mercari.sqlite3'

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

# loading inventory
with open('inventory.json') as js:
    iv = json.load(js)

# connect to db
conn = sqlite3.connect(DB_PATH)
c = conn.cursor()

# create table
c.execute("""CREATE TABLE IF NOT EXISTS items (
                id INTEGER PRIMARY KEY,
                name STRING,
                category_id INTEGER,
                image STRING,
                FOREIGN KEY(category_id) REFERENCES category(category_id)
          )""")

c.execute("""CREATE TABLE IF NOT EXISTS category (
               category_id INTEGER PRIMARY KEY,
               name STRING
          )""")

# inv = [('jacket', 'fashion'), ('cat-cage', 'pet'), ('switch', 'entertainment')]

# operations on db
# c.execute("INSERT INTO items (name,category) VALUES ('jacket', 'fashion')") #insert one
# c.execute("INSERT INTO items (name,category_id) VALUES ('jacket', '1')") #insert one
# c.execute('DELETE FROM category;',) #drop all data in table
# c.execute('DELETE FROM items WHERE id = 6 or 7 or 8') # delete certain item
# c.executemany("INSERT INTO items(name,category) VALUES(?,?)", inv) #insert many
# c.execute("""ALTER TABLE items
# ADD COLUMN image STRING""") # add new column
# c.execute("""DROP table items""") #drop the table
# RENAME COLUMN location TO image""")# rename the column

# c.execute("INSERT INTO category (name) VALUES('game')")
# c.execute("UPDATE items SET category_id='1' WHERE category='fashion'")

conn.commit()
conn.close()


# Hash the image
def hash_image(str):
    hashed_str = hashlib.sha256(str.replace('.jpg', '').encode('utf-8')).hexdigest() + '.jpg'
    return hashed_str


def get_specific_items(**args):
    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        sql = """SELECT items.name,
             items.image, 
             category.name as category
             FROM items INNER JOIN category
             ON items.category_id =category.category_id """
        c.execute(sql)
        conn.row_factory = sqlite3.Row
        if 'name' in args:
            suffix = "WHERE items.name=?"
            print(sql + suffix)
            c.execute(sql + suffix, (args['name'].lower(),))

        if 'id' in args:
            suffix = "WHERE items.id=?"
            c.execute(sql + suffix, (args['id'],))
        r = [dict((c.description[i][0], value)
                  for i, value in enumerate(row)) for row in c.fetchall()]
        conn.commit()
        conn.close()
        return r if r else None
    except Error as e:
        print(e)
        return None


def add_one_item(name, category_id, image):
    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute("INSERT INTO items(name,category_id, image) VALUES (?,?,?)",
                  (name, category_id, hash_image(image) if image else None))
        conn.commit()
        conn.close()
        print('add succefully!')
    except Error as e:
        print(e)
        return None


# add_one_item('cate-cage','2','images/default.jpg')
# print(get_all_items())
# print(get_specific_items(id=2))


# API PART
@app.get("/")
def root():
    return {"message": "Hello, world!"}


# POST
@app.post("/items")
async def add_one_item(name: str = Form(...),
                       category: str = Form(...),
                       image: str = Form(...)):
    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute("INSERT INTO category (name) VALUES (?)", (category,))
        c.execute("SELECT category_id FROM category WHERE name=?", (category,))
        category_id = c.fetchone()[0]
        c.execute("INSERT INTO items(name,category_id, image) VALUES (?,?,?)",
                  (name, category_id, hash_image(image) if image else None))
        conn.commit()
        conn.close()
        print('add successfully!')
    except Error as e:
        print(e)
        logger.debug(f"failed to add: {name}")
        return {"message": f"Failed to add: {name}"}

    result = {"name": name, "category": category, "image": image}
    logger.info(f"Receive item: {result}")
    return {"message": f"item received: {name}"}


# async def add_item(
#         name: str,
#         category_id: int,
#         image: str = None
# ):
#     # add_one_item(name, category_id, image)
#     result = {"name": name, "category_id": category_id, "image": image}
#     logger.info(f"Receive item: {result}")
#     return {"message": f"item received: {name}"}

# GET

@app.get("/items")
async def get_all_items():
    try:
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        conn.row_factory = sqlite3.Row
        sql = """SELECT items.name,
        items.image, 
        category.name as category
        FROM items INNER JOIN category
        ON items.category_id =category.category_id"""

        c.execute(sql)
        r = [dict((c.description[i][0], value)
                  for i, value in enumerate(row)) for row in c.fetchall()]
        conn.commit()
        conn.close()
        return {'items': r} if r else None
    except Error as e:
        print(e)
        return None


@app.get("/search/")
async def read_item(keyword: str):
    return {"items": get_specific_items(name=keyword)}


@app.get("/items/{item_id}")
async def read_item(item_id: int):
    return {"items": get_specific_items(id=item_id)}


@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path

    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():

        logger.info(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
