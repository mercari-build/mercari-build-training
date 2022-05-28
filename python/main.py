import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import db
import urllib.parse
import hashlib
from os.path import join, dirname, realpath
from translate import *

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


@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.get("/items")
async def read_items():
    items = db.get_items()
    all_items = []
    for item in items:
        all_items.append(
            {"id": item[0], "ja_name": item[1], "en_name": item[2], "category": item[3], "image": item[4]})
    return all_items


@app.get("/items/{item_id}")
async def read_item(item_id: int):
    item = db.get_item(item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    return {"id": item[0], "name": item[1], "category": item[2], "image": item[5]}


@app.get("/search")
async def read_items(keyword: str):
    all_items = {"items": []}
    items = db.search_items(keyword)
    for item in items:
        all_items["items"].append(
            {"id": item[0], "en_name": item[1], "ja_name":item[2], "category": item[4], "image": item[5]})
    logger.info(f"{all_items}")
    return all_items


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    category_id = int(category)
    image_hash = hash_image(image.file.read())
    
    translated_name = translate_item(name)
    detected_language = language_detection(name)
    
    if detected_language == 'ja':
        ja_name = name
        en_name = translated_name
    else:
        ja_name = translated_name
        en_name = name

    db.add_item(ja_name, en_name, category, image_hash)
    return {"id": image_hash, "Japanese name": ja_name, "English name": en_name, "category": category_id, "image_filename": image_hash}


@app.get("/image/{image_filename}")
async def get_image(image_filename):
    if not image_filename.endswith(".jpg"):
        image_filename += ".jpg"
    image_path = images / image_filename
    if not image_path.exists():
        raise HTTPException(status_code=404, detail="Image not found")
    return FileResponse(image_path)


@app.delete("/items/{item_id}")
def delete_item(item_id: int):
    item = db.get_item(item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    db.delete_item(item_id)
    logger.info(f"Delete item: {item_id}")
    return {"message": f"Item {item_id} deleted"}


def hash_image(image):
    image_hash = hashlib.sha256(image).hexdigest()
    image_filename = image_hash + ".jpg"
    image_path = images / image_filename
    with open(image_path, "wb") as f:
        f.write(image)
    return image_hash

def translate_item(item_name):
    check_ascii = item_name.isascii()
    if check_ascii == True:
        return translate_to_japanese(item_name)
    else:
        return translate_to_english(item_name)

def language_detection(item_name):
    check_ascii = item_name.isascii()
    if check_ascii == True:
        return detect_language(item_name)
    else:
        encoded_item_name = urllib.parse.quote(item_name.encode('utf-8'))
        return detect_language(encoded_item_name)