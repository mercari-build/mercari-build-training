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
images = pathlib.Path(__file__).parent.resolve() / "image"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
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

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Receive item: {name}")

    items_list = {"items" : []}
    if os.path.isfile('items.json') == True:
        with open('items.json') as items_json_file:
            item_list = json.load(items_json_file)

    new_data = {"name" : name, "category" : category}
    items_list["items"].append(new_data)

    with open("items.json", "w") as items_json_file:
        json.dump(items_list, items_json_file)

    return {"message": f"item received: {name}"}

@app.get("/items")
def get_item():
    items_list = {"items" : []}
    if os.path.isfile("items.json") == True:
        with open("items.json") as items_json_file:
            items_list = json.load(items_json_file)

    return items_list

@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path
    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
