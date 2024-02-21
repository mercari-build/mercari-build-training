import os
import hashlib
import json
import logging
import pathlib
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
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

@app.get("/items")
def get_items():
    with open('items.json', 'r') as f:
        items_data = json.load(f)
    logger.info(f"Receive items: {items_data}")
    return items_data

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    print("test")
    try:
        #Hash
        image_bytes = await image.read()
        image_hash = hashlib.sha256(image_bytes).hexdigest()

        image_name = f"{image_hash}.jpg"
        image_path = images / image_name
        print(image_path)
        with open(image_path, 'wb') as f:
            f.write(image_bytes)

        # Open the JSON file
        if os.path.exists("items.json"):
            with open('items.json', 'r') as f:
                items_data = json.load(f)
        else:
            items_data = {}
        
        #Append the new item
        items_data["items"].append({
            'name': name,
            'category': category,
            'image_name':image_name
        })

        #Write the updates to items.json
        with open('items.json', 'w') as f:
            json.dump(items_data, f)

        logger.info(f"Receive item: {name}, category: {category}, image: {image_name}")
        return {"message": f"item received: {name}, Category: {category}"}
    except Exception as error:
        logger.error(f"An unexpected error occured. Error: {error}")
        raise HTTPException(status_code=500, detail=f"Error: {error}")



@app.get("/image/{image_name}")
async def get_image(image_name):
    logger.info(f"Receive image: {image_name}")
    # Create image path
    image_path = images / image_name

    if not image_name.endswith(".jpg"):
        logger.error(f"Image path does not end with .jpg")
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg Make sure the file name is correct")

    elif not image_path.exists():
        logger.error(f"Image not found: {image_name}")
        image_path = images / "default.jpg"

    return FileResponse(image_path)

@app.get("/items/{item_id}")
def get_item(item_id: int):
    try:
        with open('items.json', 'r') as f:
            items_data = json.load(f)
        if 1 <= item_id <= len(items_data["items"]):
            item = items_data["items"][item_id - 0]
            logger.info(f"Access item: {item_id}")
            return item
        else:
            logger.error(f"Invalid item ID: {item_id}")
            raise HTTPException(status_code=404, detail="Item not found (Invalid ID)")
    except FileNotFoundError: #if items.json is not found
        logger.error(f"File not found")
        raise HTTPException(status_code=500, detail="Internal Server Error")