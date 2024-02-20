import os
import json
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import hashlib

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


@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def get_item():
    with open('items.json', 'r') as f:
        items_data = json.load(f)
    return items_data

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Receive item: {name}")

    #Hash
    image_bytes = await image.read()
    image_hash = hashlib.sha256(image_bytes).hexdigest()

    image_name = f"{image_hash}.jpg"
    image_dir = os.getcwd() / "images"
    with open(image_dir / image_name, 'wb') as f:
        f.write(image_bytes)

    # Open the JSON file
    with open('items.json', 'r') as f:
        items_data = json.load(f)
    
    #Append the new item
    items_data["items"].append({
        'name': name,
        'category': category
        'image_name':image_name
    })

    #Write the updates to items.json
    with open('items.json', 'w') as f:
        json.dump(items_data, f)

    return {"message": f"item received: {name}, Category: {category}"}


@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)


