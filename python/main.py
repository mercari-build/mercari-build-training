import os
import logging
import pathlib
from pathlib import Path
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.encoders import jsonable_encoder
import json
import sqlite3

cwd = Path.cwd()
parent = cwd.parent
database_file = parent / 'db/items.db'

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "image"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)

#For saving into JSON file
'''
def write_json(new_item, filename='items.json'):
    open(filename, 'a').close()
    if os.stat(filename).st_size == 0:
        with open(filename, 'w') as fp:
            res = {"items":[new_item]}
            json.dump(res,fp)
    else:
        with open(filename,'r+') as fp:
            file_data = json.load(fp)
            # Join new_data with file_data inside emp_details
            file_data["items"].append(new_item)
            # Sets file's current position at offset.
            fp.seek(0)
            # convert back to json.
            json.dump(file_data, fp)

'''
# For saving into the database with one table
'''
def write_json(new_item):
    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    c2 = c.execute('SELECT max(id) FROM items')
    max_id = c2.fetchone()[0]
    if max_id is None:
        max_id = 0

    params = (max_id+1, new_item["name"], new_item["category"])
    c.execute("INSERT INTO items VALUES (?, ? ,?)",params)

    conn.commit()
    conn.close()
'''
# adds onto database with two tables (items and category)
def write_json(new_item):
    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    c2 = c.execute('SELECT max(id) FROM items')
    max_id = c2.fetchone()[0]
    if max_id is None:
        max_id = 0

    c3 = c.execute('SELECT max(id) FROM category')
    max_category_id = c3.fetchone()[0]
    if max_category_id is None:
        category_id = 0
        params2 = (1, new_item["category"])
        c.execute("INSERT INTO category VALUES (?, ?)",params2)
    else:

        c4 = c.execute('SELECT id FROM category WHERE name=?', (new_item["category"],))
        category_id = c4.fetchone()
        if category_id is None:
            category_id = max_category_id+1
            params2 = (category_id, new_item["category"])
            c.execute("INSERT INTO category VALUES (?, ?)",params2)

    params = (max_id+1, new_item["name"], category_id, new_item["image"])
    c.execute("INSERT INTO items VALUES (?, ? ,?, ?)",params)


    conn.commit()
    conn.close()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):
    newItem = {
    "name": f"{name}",
    "category": f"{category}",
    "image" : f"{image}"
    }
    write_json(newItem)

    logger.info(f"Receive item: {name} {category} {image}")
    return {"message": f"item received: {name} "}

@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path
    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/items")
def get_items():
    data = {"items":[]}

    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    c.execute('SELECT items.name, category.name, image FROM items JOIN category ON items.category_id = category.id')
    items = c.fetchall()

    for item in items:
        new_item = {
        "name": f"{item[0]}",
        "category": f"{item[1]}",
        "image": f"{item[2]}"
        }
        data["items"].append(new_item)

    conn.commit()
    conn.close()

    return data

# GET request with item_id
@app.get("/items/{item_id}")
def get_item_from_id(item_id):

    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    c.execute('SELECT items.name, category.name, image FROM items JOIN category ON items.category_id = category.id WHERE items.id = ?',item_id)
    items = c.fetchone()

    res = {
    "name": f"{items[0]}",
    "category": f"{items[1]}",
    "image": f"{items[2]}"
    }

    conn.commit()
    conn.close()

    return res

# GET request for /search
@app.get("/search")
def search_items(keyword: str):
    data = {"items":[]}

    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    params = (keyword, keyword)
    c.execute('SELECT items.name, category.name, image FROM items JOIN category ON items.category_id = category.id WHERE items.name = ? OR category.name = ?', params)
    items = c.fetchall()

    for item in items:
        new_item = {
        "name": f"{item[0]}",
        "category": f"{item[1]}",
        "image": f"{item[2]}"
        }
        data["items"].append(new_item)

    conn.commit()
    conn.close()

    return data

# For GET request from JSON file
'''
def get_items():
    filename = "items.json"
    open(filename, 'a').close()
    if os.stat(filename).st_size == 0:
        return {"message": "no items"}
    else:
        with open(filename,'r+') as fp:
            items = json.loads(fp.read())
            return items
'''
# GET request (/search) to database with one table
'''
@app.get("/items")
def get_items():
    data = {"items":[]}

    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    c.execute('SELECT * FROM items')
    items = c.fetchall()

    for item in items:
        new_item = {
        "name": f"{item[1]}",
        "category": f"{item[2]}"
        }
        data["items"].append(new_item)

    conn.commit()
    conn.close()

    return data

# GET request to database with one table

@app.get("/items/{item_id}")
def get_item_from_id(item_id):

    conn = sqlite3.connect(database_file, check_same_thread=False)
    c = conn.cursor()
    c.execute('SELECT * FROM items WHERE id=?',item_id)
    items = c.fetchone()

    res = {
    "name": f"{items[1]}",
    "category": f"{items[2]}",
    "image": f"{items[3]}"
    }

    conn.commit()
    conn.close()

    return res
'''
