import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException, UploadFile, File
import hashlib
from typing import List, Optional
from fastapi.responses import FileResponse 
from fastapi.middleware.cors import CORSMiddleware


app = FastAPI()
UPLOAD_DIR = "images"
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


@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    try:
        # Logging the received item
        logger.info(f"Received item: {name} in category: {category}")
        image_data = image.file.read()
        image_hash = hashlib.sha256(image_data).hexdigest()

        # Save the image with the hashed filename
        image_path = os.path.join(UPLOAD_DIR, f"{image_hash}.jpg")
        with open(image_path, "wb") as file:
            file.write(image_data)
        
        # Save the item into items.json
        save_item_to_json(name, category, f"{image_hash}.jpg")

        return {"message": f"Item received: {name} in category: {category} with image uploaded"}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Internal Server Error")

@app.get("/items")
async def get_items():
    try:
        with open('items.json', 'r') as file: 
            items = json.load(file)
        return items
    except (FileNotFoundError, json.JSONDecodeError):
        return {"items": []}
    
with open('items.json', 'r') as file:
    items_data = json.load(file)
    
items = items_data['items']
@app.get("/items/{item_id}")
async def get_item(item_id: int):
    for item in items:
        if item.get('id') == item_id:
            return item

    raise HTTPException(status_code=404, detail="Item not found")

def save_item_to_json(name, category, image_filename):
    try:
        if os.path.exists('items.json'):
            with open('items.json', 'r') as file:
                existing_items = json.load(file).get('items', [])
        else:
             existing_items = []
    except (FileNotFoundError, json.JSONDecodeError):
        existing_items = []

    new_item = {"name": name, "category": category, "image": image_filename}
    existing_items.append(new_item)

    with open('items.json', 'w') as file:
        json.dump({"items": existing_items}, file, indent=4)

@app.get("/image/{image_name}")
async def get_image(image_name):
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
   

