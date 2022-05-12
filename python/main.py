from asyncore import file_dispatcher
from calendar import c
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

import json
from pathlib import Path


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


def add_json(data,file_json='item.json'):
    path = Path(file_json)
    if not path.is_file():
        file_data={"items": []}
    else:
        with open(file_json,'r')as f:
            file_data=json.load(f)


    with open(file_json, 'w') as f:
        file_data["items"].append(data)
        json.dump(file_data, f)


@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    add_json({"name": name, "category": category})
    logger.info(f"Receive item: {name} , {category}")
    return {"message": f"item received: {name} , {category}"}

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
