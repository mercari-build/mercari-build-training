import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse, JSONResponse
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
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    new_item = {"name": name, "category": category}
    if items_file.exists():
        with open(items_file, "r+", encoding="utf-8") as file:
            data = json.load(file)
            data['items'].append(new_item)
            file.seek(0)
            json.dump(data, file, indent=4)
    else:
        with open(items_file, "w", encoding="utf-8") as file:
            json.dump({"items": [new_item]}, file, indent=4)
    logger.info(f"Item added: {name}")
    return {"message": f"item received: {name}"}

@app.get("/items")
def get_items():
    if items_file.exists():
        with open(items_file, "r", encoding="utf-8") as file:
            data = json.load(file)
            return JSONResponse(content=data)
    else:
        return {"items": []}

@app.get("/image/{image_name}")
async def get_image(image_name: str):
    image_path = images / image_name
    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image must be a .jpg file")
    if not image_path.exists():
        logger.info(f"Image not found: {image_name}")
        return HTTPException(status_code=404, detail="Image not found")
    return FileResponse(image_path)
