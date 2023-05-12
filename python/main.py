import json
import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware


app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get('FRONT_URL', 'http://localhost:3000')]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def get_item():
    try:
        with open("items.json", "r") as f:
            mydata = json.load(f)
            return mydata
    except FileNotFoundError:
        items = {"items": []}
        with open("items.json", "w") as f:
            json.dump(items, f)
        return items

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Receive item: {name}, Receive category: {category}, Receive image:{image.filename}")

    # Hash the image using sha256, and save it with the name <hash>.jpg
    file = image.file.read()
    image_hash = hashlib.sha256(file).hexdigest()
    filename = image_hash + ".jpg"
    path = images / filename
    with open(path, "wb") as f:
        f.write(file)

    # Add new items into json file
    try:
        with open("items.json", "r") as f:
            items = json.load(f)
    except FileNotFoundError:
        items = {"items": []}

    items["items"].append({"name": name, "category": category, "image_filename": filename})
    with open("items.json", "w") as f:
        json.dump(items, f)

    return {"message": f"item received: {name}, category received: {category}, image received: {filename}"}

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

@app.get("/items/{item_id}")
def get_itemsid(item_id:int):
    try:
        with open("items.json", "r") as f:
            mydata = json.load(f)
            return mydata["items"][item_id]
    except IndexError:
        raise HTTPException(
            status_code=404, detail=f"item_id {item_id} not exist"
        )