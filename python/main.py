import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import db
import hashlib

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "image"
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
            {"id": item[0], "name": item[1], "category": item[2], "image": item[3]})
    return all_items


@app.get("/items/{item_id}")
async def read_item(item_id: int):
    item = db.get_item(item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    return {"id": item[0], "name": item[1], "category": item[2], "image": item[3]}


@app.get("/search")
async def read_items(keyword: str):
    all_items = {"items": []}
    items = db.search_items(keyword)
    for item in items:
        all_items["items"].append(
            {"id": item[0], "name": item[1], "category": item[2], "image": item[3]})
    logger.info(f"{all_items}")
    return all_items


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):
    category_id = int(category)
    image_hash = hash_image(image)
    db.add_item(name, category_id, image_hash)
    logger.info(f"Receive item: {name}, {category}")
    return {"message": f"Item {name} added"}


@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = pathlib.Path(
        __file__).parent.resolve() / "images" / image_filename
    if not image_filename.endswith(".jpg"):
        raise HTTPException(
            status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)


@app.delete("/items/{item_id}")
def delete_item(item_id: int):
    item = db.get_item(item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    db.delete_item(item_id)
    logger.info(f"Delete item: {item_id}")
    return {"message": f"Item {item_id} deleted"}

# hash image and save to /images


def hash_image(image):
    filename = ""
    filename = filename + image
    readable_hash = ""
    # filename = "images/test.jpg"
    with open(filename, "rb") as f:
        bytes = f.read()
        readable_hash = readable_hash + hashlib.sha256(bytes).hexdigest()
        # print(readable_hash + ".jpg")
    with open("images/" + readable_hash + ".jpg", "wb") as f:
        f.write(bytes)
    return readable_hash
