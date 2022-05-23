import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

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


@app.get("/")
def root():
    return {"message": "Hello, world!"}


filename = "items.json"


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    with open(filename, 'r') as file:
        items_dict = json.load(file)
    items_dict["items"].append({"name": name, "category": category})
    with open(filename, 'w') as file:
        json.dump(items_dict, file)
    logger.info(f"Receive item: {name}")
    return {"message": f"item received: {name}"}


@app.get("/items")
def display_item():
    with open(filename, 'r') as file:
        items_dict = json.load(file)
    return items_dict


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
