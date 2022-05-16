import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import db

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
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
    all_items = {"items": []}
    items = db.get_items()
    for item in items:
        all_items["items"].append(
            {"id": item[0], "name": item[1], "category": item[2]})
    return all_items


@app.get("/items/{item_id}")
async def read_item(item_id: int):
    item = db.get_item(item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    return {"id": item[0], "name": item[1], "category": item[2]}


@app.get("/search")
async def read_items(keyword: str):
    all_items = {"items": []}
    items = db.search_items(keyword)
    for item in items:
        all_items["items"].append(
            {"id": item[0], "name": item[1], "category": item[2]})
    return all_items


@app.post("/items")
def add_item(id: int, name: str, category: str):
    db.add_item(id, name, category)
    logger.info(f"Receive item: {name}, {category}")
    return {"message": f"Item {name} added"}


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
