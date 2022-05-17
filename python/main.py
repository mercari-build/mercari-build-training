import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
# import json
import database

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

"""
Gets the list of all items
"""
@app.get("/")
def root():
    # get the list of all items in the database
    items = database.get_items()
    # format the list and return
    return format_items(items)

"""
Gets item with the given item_id
"""
@app.get("/items/{item_id}")
async def get_item_by_id(item_id):

    item = database.get_id_by_id(item_id)

    if item == None:
        logger.debug(f"Item not found: {item_id}")
    
    return format_items(item)


"""
Create a new item
"""
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):

    logger.info(f"Receive item: name = {name}, category = {category}, image = {image}")

    # add a new item in the database
    database.add_item(name, category, image)

    # return message
    return {"message": f"item received: {name}"}

@app.get("/search")
async def search_items(keyword: str):
    logger.info(f"Receive search_keyword: keyword = {keyword}")

    # get the list of items with name that contains the given keyword
    items = database.search_items(keyword)

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
        item_format = {"name": item[0], "category": item[1], "image": item[2]}
        items_format.append(item_format)

    return {"items": f"{items_format}"}
