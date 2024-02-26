import os
import logging
import pathlib
import hashlib
import json
from fastapi import FastAPI, Form, HTTPException, UploadFile
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

@app.post("/items")
def add_item(name: str = Form(...), category: str=Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}")
    logger.info(f"Receive item: {category}")
    logger.info(f"Receive item: {image}")

    image_name = image.filename
    hashed_image_name = get_hash(image_name)
    save_image(image, hashed_image_name)

    message = {
        "items":[
            {
                "name": name,
                "category": category,
                "image_name": hashed_image_name
            }
        ]
    }

    return message

def get_hash(image):
    hash = hashlib.sha256(image.encode()).hexdigest()
    return hash+".jpg"

def save_image(image,jpg_hashed_image_name):
    imagefile = image.file.read()
    image = images / jpg_hashed_image_name
    with open(image, 'wb') as f:
        f.write(imagefile)
    return


@app.get("/items")
def get_item():
    f = open("items.json")
    data = json.load(f)

    return data


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
