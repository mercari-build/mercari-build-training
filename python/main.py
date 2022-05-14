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

file_json = 'items.json'
if not os.path.isfile(file_json):
    initial_items_data = {"items": []}
    with open(file_json, 'w') as outfile:
        json.dump(initial_items_data, outfile)

def add_item_to_json(item, file_json):
    try:
        with open(file_json, 'r') as f:
            items_data = json.load(f)
    except json.JSONDecodeError as e:
        print('JSONDecodeError', e)
    
    items_data["items"].append(item)

    with open('items.json', 'w') as outfile:
        json.dump(items_data, outfile)

    print(f'added {item} to {file_json}')


@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Receive item: {name}, {category}")
    item = {"name": name, "category": category}
    add_item_to_json(item, file_json)
    return {"name": name, "category": category}

@app.get("/items")
def get_items():
    if not os.path.isfile(file_json):
        print('')
    try:
        with open('items.json', 'r') as f:
            data = json.load(f)
    except json.JSONDecodeError as e:
        print('JSONDecodeError', e)
    print(data)
    return data

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
