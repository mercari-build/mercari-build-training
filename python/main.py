from asyncore import file_dispatcher
from calendar import c
from multiprocessing import allow_connection_pickling
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

import json
from pathlib import Path
import sqlite3
import hashlib


data_base_name="../db/mercari.sqlite3"

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "image"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)


def add_item_to_json(data,file_json='item.json'):
    path = Path(file_json)
    if not path.is_file():
        file_data={"items": []}
    else:
        with open(file_json,'r')as f:
            file_data=json.load(f)


    with open(file_json, 'w') as f:
        file_data["items"].append(data)
        json.dump(file_data, f)
    print (file_data)


@app.get("/")
def root():
    return {"message": "Hello, world!"}


# @app.post("/items")
# def add_item(name: str = Form(...), category: str = Form(...)):
#     add_item_to_json({"name": name, "category": category})
#     logger.info(f"Receive item: {name} , {category}")
#     return {"message": f"item received: {name} , {category}"}


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()

    cur.execute('''insert or ignore into category(name) values (?)''', (category,))
    cur.execute('''select id from category where name = (?)''', (category,))
    category_id = cur.fetchone()[0]
    logger.info(f"Receive item: {category_id}")
    hashed_filename = hashlib.sha256(image.replace(".jpg", "").encode('utf-8')).hexdigest() + ".jpg"
    cur.execute('''insert into item(name, category_id, image) values(?, ?, ?)''', (name, category_id, hashed_filename))
    conn.commit()
    conn.close()
    logger.info(f"Receive item: {name,category,hashed_filename}")

@app.get("/items")
def get_items():
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''select * from item''')
    items = cur.fetchall()
    cur.execute('''select * from category''')
    categorys=cur.fetchall()
    conn.commit()
    conn.close()
    logger.info("Get item")
    return items,categorys

@app.delete("/items")
def init_item():
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''drop table item;''')
    cur.execute('''drop table category;''')
    conn.commit()
    cur.execute('''create table item(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT,category_id INTEGER,image TEXT)''')
    cur.execute('''create table category(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE)''')
    conn.commit()
    cur.close()
    conn.close()

@app.get("/search")
def search_item(keyword:str):
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''select item.name,category.name as category,item.image from item inner join category on category.id = item.category_id where item.name like (?)''', (f"%{keyword}%", ))
    items = cur.fetchall()
    conn.close()
    logger.info(f"Get items with name containing {keyword}")
    return items

@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = image / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/items/{item_id}")
def get_items_from_id(item_id):
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''select item.name,category.name as category,item.image from item inner join category on category.id = items.category_id where items.category_id = (?)''', (item_id, ))
    items=cur.fetchall()
    conn.commit()
    conn.close()
    logger.info(f"Receive item: {items}")
    return items

