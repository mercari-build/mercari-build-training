import os
import json
import logging
import pathlib
import hashlib
from fastapi import FastAPI,Path, Form,UploadFile, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"
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
    return {"message": "Hello world!"}
def load_items_from_json():
    items_path = pathlib.Path(__file__).parent.resolve() / “items.json”
    try:
        with open(items_path, “r”) as file:
            items = json.load(file)
    except FileNotFoundError:
        items = []
    return items
def save_items_to_json(items):
    items_path = pathlib.Path(__file__).parent.resolve() / “items.json”
    with open(items_path, “w”) as file:
        json.dump(items, file, indent=2)
@app.post("/items")
def add_item(name: str = Form(...),category: str =Form(...),image_path: str= Form(...)):
    logger.info(f"Receive item: {name}")
     # step1:check if the file exists and if not create an empty json list
    if not items_file.exists() or not items_file.is_file():
        items_data = {"items": []}
    else:
        # if file exists load the existing data
        try:
            with open(items_file, 'r') as file:
                items_data = json.load(file)
        except json.decoder.JSONDecodeError:
            # handling case if the file content is not valid json
            items_data = {"items": []}
    #converting the image to image_hash
    image_filename = os.path.basename(image_path)
    image_hash = hashlib.sha256(image_filename.encode()).hexdigest()
    # appending new item recieved
    new_item = {"name": name, "category": category,"image":f"{image_hash}.jpg"}
    items_data["items"].append(new_item)
    # rewriting the file with updated data
    with open(items_file, 'w') as file:
        json.dump(items_data, file, indent=2)
    return {"message": f"item received: {name}"}
@app.get("/items")
def get_items():
    if not items_file.exists() or not items_file.is_file():
        items_data = {"items": []}
    else:
        # if file exists load the existing data
        try:
            with open(items_file, 'r') as file:
                items_data = json.load(file)
        except json.decoder.JSONDecodeError:
            # handling case if the file content is not valid json
            items_data = {"items": []}
    return {"items": items_data}
@app.get("/items/{item_id}")
def get_one_item(item_id : int = Path(..., title="The ID of the item to retrieve")):
    if not items_file.exists() or not items_file.is_file():
        items_data = {"items": []}
    else:
        # if file exists load the existing data
        try:
            with open(items_file, 'r') as file:
                items_data = json.load(file)
        except json.decoder.JSONDecodeError:
            # handling case if the file content is not valid json
            items_data = {"items": []}
    if 0 <= item_id - 1 < len(items_data["items"]):
        result = items_data["items"][item_id - 1]
        return result
    else:
        return {"detail": "Item not found"}
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
