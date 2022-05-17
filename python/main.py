import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import json
import sqlite3

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)
# items = []

@app.get("/")
def root():
    # print("root is called")
    # print(json.dumps({"items": items}))

    # connect
    db_connect = sqlite3.connect('../db/mercari.sqlite3')

    # cursor
    db_cursor = db_connect.cursor()
 
    # insert new data
    sql = 'SELECT name, category FROM items'

    # execute
    db_cursor.execute(sql)

    # get the list of tuples
    items = db_cursor.fetchall()
    # print(items)

    # close the connection
    db_connect.close()

    return format_items(items)

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Receive item: name = {name}, category = {category}")

    # item = {"name": name, "category": category}
    # items.append(item)

    # with open("items.json", "w") as f:
    #     #print(json.dumps(items, indent=4))
    #     json.dump({"items": items}, f)

    # connect
    db_connect = sqlite3.connect('../db/mercari.sqlite3')

    # cursor
    db_cursor = db_connect.cursor()

    # create a table
    # sql = 'CREATE TABLE items(id INTEGER PRIMARY KEY AUTOINCREMENT, name STRING, category STRING)'
    
    # insert new data
    sql = 'INSERT INTO items(name, category) values (?, ?)'
    data = [name, category]

    # execute
    db_cursor.execute(sql, data)

    # commit
    db_connect.commit()

    # close the connection
    db_connect.close()

    return {"message": f"item received: {name}"}

@app.get("/search")
async def search_items(keyword: str):
    logger.info(f"Receive search_keyword: keyword = {keyword}")

    # connect
    db_connect = sqlite3.connect('../db/mercari.sqlite3')

    # cursor
    db_cursor = db_connect.cursor()

    # search item where the name contains the given keyword
    sql = 'SELECT name, category FROM items WHERE name LIKE ?'
    data = ('%' + keyword + '%',)

    # execute
    db_cursor.execute(sql, data)
    items = db_cursor.fetchall()
    # print(items)

    # close the connection
    db_connect.close()

    return format_items(items)

@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

"""
Formats the given list of tuples for printing
"""
def format_items(items):
    # create a list to set each item in a format
    items_format = []
    for item in items:
        item_format = {"name": item[0], "category": item[1]}
        items_format.append(item_format)

    return {"items": f"{items_format}"}
