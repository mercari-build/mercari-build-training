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
def add_item(name: str = Form(...), category: str = Form(...)):
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''insert into items(name,category) values (?, ?)''', (name,category))
    conn.commit()
    cur.close()
    conn.close()

@app.get("/items")
def get_items():
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''select id,name,category from items''')
    items = cur.fetchall()
    conn.commit()
    conn.close()
    logger.info("Get items")
    return items

@app.delete("/items")
def init_item():
    conn = sqlite3.connect(data_base_name)
    cur = conn.cursor()
    cur.execute('''drop table items;''')
    conn.commit()
    cur.execute('''create table items(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT,category TEXT)''')
    conn.commit()
    cur.close()
    conn.close()

# @app.delete("/items")
# def init_item():
#     path = Path(file_json)
#     if os.path.exists(path):
#         os.remove(path)
#     else:
#         print("Can not delete the file as it doesn't exists")


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
