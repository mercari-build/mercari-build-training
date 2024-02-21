import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, UploadFile
from pathlib import Path
import json
import hashlib
import base64
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

images_dir = Path(__file__).parent.resolve() / "images"
images_dir.mkdir(parents=True, exist_ok=True)
images_dir = Path("C:/Users/ctech/mercari-build-training/items.json")
images_dir = Path("images")
items_file = Path("items.json")

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}")
    try:
        with open(items_file, 'r') as file:
            items_data = json.load(file)
    except (FileNotFoundError, json.JSONDecodeError):
        items_data = {"items": []}

    image_bytes = await image.read()
    image_hash = hashlib.sha256(image_bytes).hexdigest()
    image_filename = f"{image_hash}.jpg"
    image_path = images_dir / image_filename
    with open(image_path, "wb") as f:
        f.write(image_bytes)

    new_item = {"name": name, "category": category, "image_name": image_filename}
    items_data["items"].append(new_item)

    with open(items_file, 'w') as file:
        json.dump(items_data, file, indent=2)

    return {"message": f"item received: {name}", "items": items_data["items"]}
@app.get("/items")
def get_items():
    logger.info("Retrieving items")
    try:
        with open(items_file, 'r') as file:
            items_data = json.load(file)
    except (FileNotFoundError, json.JSONDecodeError):
        items_data = {"items": []}
    return items_data

@app.get("/items/{item_id}")
def get_item(item_id: int):
    logger.info(f"Retrieving item with ID: {item_id}")
    try:
        with open(items_file, 'r') as file:
            items_data = json.load(file)
    except (FileNotFoundError, json.JSONDecodeError):
        items_data = {"items": []}
    if  0 <= item_id < len(items_data["items"]):
        return items_data["items"][item_id]
    else:
        raise HTTPException(status_code=404, detail="Item not found")
    

@app.get("/image/{image_name}")
async def get_image(image_name):
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
