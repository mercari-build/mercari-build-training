import os
import json
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
import hashlib
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import FileResponse , JSONResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "images"
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware)


@app.post("/items")
def add_item(name: str = Form(...)):
    logger.info(f"Receive item: {name}")
    return {"message": f"item received: {name}"}
async def add_item(name: str = Form(...), category: str = Form(...),image: UploadFile = File(...)):
    contents = await image.read()
    hash_sha256 = hashlib.sha256(contents).hexdigest()
    image_filename = f"{hash_sha256}.jpg"
    image_path = images / image_filename
    with open(image_path, "wb") as file:
        file.write(contents)

    item = {"name": name, "category": category,"image_name": image_filename}
    logger.info(f"Receive item: {item}")
    save_item(item)
    return {"message": f"Item received: {item}","image_name": image_filename}


def save_item(item):
    if items_file.exists():
        with open(items_file, "r+", encoding="utf-8") as file:
            data = json.load(file)
            data["items"].append(item)
            file.seek(0)
            json.dump(data, file, indent=4)
            file.truncate()
    else:
        with open(items_file, "w", encoding="utf-8") as file:
            json.dump({"items": [item]}, file, indent=4)


@app.get("/items")

def get_list():
    items = load_items()
    return {"items": items}

@app.get("/items/{item_id}")

def get_items(item_id: int):
    if item_id < 0 or item_id >= len(items):
        raise HTTPException(status_code=404, detail="Item not found")
    return items[item_id]


def load_items():
    if items_file.exists():
        with open(items_file, "r", encoding="utf-8") as file:
            try:
                data = json.load(file)
                return data["items"]
            except json.JSONDecodeError:
                print("Error decoding JSON")
                return []
    else:
        return []



@app.get("/image/{image_name}")

async def get_image(image_name):

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")


    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

        return FileResponse(image)

    return FileResponse(image)

#uvicorn main:app --reload --log-level debug --port NUMBEROFPORT