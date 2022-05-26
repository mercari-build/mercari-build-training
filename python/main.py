import os
import logging
import pathlib
import sqlite3
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

import sqlite3
import json

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


# filename = "items.json"
DatabaseName = "../db/mercari.sqlite3"


# def format_items(items):
#     items_format = []
#     for item in items:
#         item_format = {"name": item[0], "category": item[1]}
#         items_format.append(item_format)

#     return {"items": f"{items_format}"}


@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):

    # connect
    conn = sqlite3.connect(DatabaseName)
    # cursor
    cur = conn.cursor()

    cur.execute("INSERT INTO items VALUES (?,?)",
                (name, category))

    conn. commit()
    conn.close()

    logger.info(f"Receive item: {name}")
    return {"message": f"item received: {name}"}


@app.get("/items")
def display_item():
    # connect
    conn = sqlite3.connect(DatabaseName)
    # cursor
    cur = conn.cursor()

    cur.execute("SELECT * FROM items")

    item_list = cur.fetchall()

    # close connection
    conn.close()

    # return formatted list of items from db
    # return format_items(item_list)

    return item_list


@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(
            status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
