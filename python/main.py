
import os
import logging
import pathlib

#step3-2
import json

#STEP3-4
import hashlib

from fastapi import FastAPI, Form, File, HTTPException, UploadFile
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

@app.get("/")
def root():
    return {"message": "Hello, world!"}

#STEP3-3
@app.get("/items")
def get_item():
    with open("items.json", "r") as file:
        item_data = json.load(file)
    logger.info(f"Receive items: {item_data}")
    return item_data

    
#STEP3-2,3-4
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    image_filename = get_image_filename(image)
    new_item = {"name": name, "category": category, "image": image_filename}
    items_file = "items.json"

    #open the json file, create a new file if it doen't exist
    if os.path.exists(items_file):
        with open("items.json", "r") as file:
            item_data = json.load(file)
            item_list = item_data.get("items",[])

    else:
        item_list = []

    #add information to the list
    if "items" in item_data.keys():
        item_list.append(new_item)
    else:
        item_list.append([])

    #write to the json file
    with open(items_file, "w") as file:
        json.dump({"items": item_list}, file)

    logger.info(f"Receive item: {name}, {category}")
    
    return {"message": f"Item received: {name}, {category},{image_filename}"}

#STEP3-4
def get_image_filename(image):
    image_contents = image.file.read()
    image_hash = hashlib.sha256(image_contents).hexdigest()

    #Create a file path
    image_filename = f"{image_hash}.jpeg"
    save_path = os.path.join("images", image_filename)

    #Save a image
    with open(save_path, "wb") as f:
        f.write(image_contents)

    logger.info(f"Saved image to: {save_path}")

    return image_filename

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

#STEP3-5
@app.get("/items/{item_id}")
def get_item_information(item_id: int):
    with open("items.json", "r") as file:
        item_data = json.load(file)

    if 1<= item_id <= len(item_data[("items")]):
        item = item_data["items"][item_id - 1]
        logger.info(f"Receive item: {item}")
        return item
    else:
        raise HTTPException(status_code=404, detail="Item not found")