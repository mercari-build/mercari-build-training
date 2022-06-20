import os
import logging
import pathlib
import shutil

from sqlite3 import Error
from fastapi import FastAPI, Form, UploadFile, HTTPException, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

import database
import hashlib

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

"""
Creates tables
"""
database.create_tables()

"""
Saves the given bytes of the image file as a new file in "items" directory
Creates the hash from the given bytes and uses it as the filename
"""


def save_image(filename):
    # hash the bytes with sha256, and put '.jpg' in the end
    filename_hash = hashlib.sha256(str(filename.filename).replace('.jpg', '').encode('utf-8')).hexdigest() + '.jpg'
    save_path = images / filename_hash
    # write the given bytes to a new file
    try:
        with open(save_path, "wb") as buffer:
            shutil.copyfileobj(filename.file, buffer)
            return filename_hash
    except BufferError as e:
        logger.error(e)

"""
Main page
"""


@app.get("/")
def root():
    return {"message": "Hello, world!"}


"""
Gets the list of all items
"""


@app.get("/items")
def read_items():
    # get the list of all items in the database
    try:
        database.add_views()
        items = database.get_items(status_id=1)
        # format the list and return
        return items
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}


"""
Gets item with the given item_id
"""


@app.get("/items/{item_id}")
async def get_item_by_id(item_id):
    try:
        item = database.get_item_by_id(item_id, status_id=1)
        return item
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}

"""
Search requests by item_id
"""


@app.get("/requests/{item_id}")
async def get_request_by_id(item_id):
    try:
        item = database.get_item_by_id(item_id, status_id=2)
        return item
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}

"""
Creates a new item with the given name, cateogry, image
Accepts the arguments as File.
"""

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):

    filename_hash = save_image(image)
    logger.info(f"Created file: {filename_hash}")
    # add a new item in the database with the hashed filename
    try:
        database.add_views()
        database.add_item(name, category, filename_hash, status_id=1)
        return {"message": f"item received: {name}"}
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}


@app.get("/search")
async def search_items_by_keyword(keyword: str):
    logger.info(f"Receive search_keyword: keyword = {keyword}")

    # get the list of items with name that contains the given keyword
    try:
        items = database.search_items(keyword, status_id=1)
        return items
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}


@app.get("/search/requests")
async def search_requests_by_keyword(keyword: str):
    logger.info(f"Receive search_keyword: keyword = {keyword}")

    # get the list of requests with name that contains the given keyword
    try:
        items = database.search_items(keyword, status_id=2)
        return items
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}



@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.info(f"Image not found: {image}")
        default_image_filename = "default.jpg"
        image = images / default_image_filename

    return FileResponse(image)

"""
Gets the list of all requests
"""

@app.get("/requests")
def read_requests():
    # get the list of all items in the database
    try:
        database.add_views()
        requests = database.get_recommend_requests()
    # format the list and return
        return requests
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}


"""
Creates a new requests with the given name, cateogry, image
Accepts the arguments as File.
"""


@app.post("/requests")
def add_request(name: str = Form(...), category: str = Form(...), image: UploadFile = File(default=None)):
    if image:
        filename_hash = save_image(image)
        logger.info(f"Created file: {filename_hash}")

    else:
        filename_hash = ""

    # add a new item in the database with the hashed filename
    try:
        database.add_views()
        database.add_item(name, category, filename_hash, status_id=2)
        return {"message": f"item received: {name}"}
    except Error as e:
        logger.error(e)
        return {'message': f'{e}'}