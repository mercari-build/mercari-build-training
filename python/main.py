import os
import logging
import pathlib
import json
import hashlib
from fastapi import FastAPI, Form, UploadFile, HTTPException
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

items_json = pathlib.Path(__file__).parent.resolve() / "items.json"

def save_item(item):
    with open(items_json, "r+") as f:
        json.dump(item, f, indent=4)

def load_item():
    if items_json.exists():
        with open(items.json,"r") as f:
            return json.load(f)
    return{"item": []}

def save_image(file,filename):
    with open(images / filename, "wb") as image:
        image.write(file)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

items_list=[]
@app.post("/items")
def add_item(name: str = Form(...), category:str=Form(...), image:UpladFile = File(...)):
    logger.info(f"Receive item: {name}, category: {category}, image: {image}")
    item={"name": name, "category": category,"image_name": image_filename}
    save_item(item)

    
    file_content = image.file.read()
    hash_value = hashlib.sha256(file_content).hexdigest()

    
    image_filename = f"{hash_value}.jpg"
    save_image(file_content, image_filename)

    
    new_item = {"name": name, "category": category, "image": image_filename}

    
    items_data = load_items_from_json()
    existing_items = items_data.get("items", [])

    
    existing_items.append(new_item)
    items_data["items"] = existing_items

   
    save_items_to_json(items_data)

    return {"message": f"item received: {name},category:{category}","image_name": image_filename}



@app.get("/items")
def get_items():
    return FileResponse(items_json)

@app.get("/image/{image_name}")
async def get_image(image_name):
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/items/{item_id}")
def get_item(item_id: int= Path(..., title="The ID of the item to get")):
    items_data=load_item()
     existing_items = items_data.get("items", [])

    if item_id < len(existing_items):
        item = existing_items[item_id-1]
        return item