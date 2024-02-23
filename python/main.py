import os
import logging
import pathlib
import hashlib
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

items = list()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def get_items():
    return {"items": items}

@app.get("/items/{item_id}")
def get_item(item_id):
    return items[int (item_id)]

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}")

    file_content = image.file.read()
    image.file.seek(0)

    image_hash = hashlib.sha256(file_content).hexdigest()
    save_image(file_content, f"{image_hash}.jpg")
        
    new_item = {"name": name, "category": category, "image": f"{image_hash}.jpg"}
    items.append(new_item)

    return {"items": items}

def save_image (file_content, hashed_filename):
    save_directory = "images/"
    os.makedirs(save_directory, exist_ok=True)

    with open(os.path.join(save_directory, hashed_filename), "wb") as f:
        f.write(file_content)

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
