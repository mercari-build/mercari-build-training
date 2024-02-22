import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import json
import hashlib
from DB import Get_all,Insert_item,SearchByKw, Get_all_ex
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
    # return Get_all("items")

    #STEP4-3
    return Get_all_ex()

@app.get("/items/{item_id}")
def get_itemsById(item_id: int):
    with open('items.json','r') as f:
        data = json.load(f)
    if item_id >= len(data["items"]):
        raise HTTPException(status_code=404, detail="Array out of bound")
    
    return data[item_id]

#STEP4-2
@app.get("/search")
def search_item(keyword: str):

    return SearchByKw("items", keyword)

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), file: UploadFile = File(...) ):
    logger.info(f"Receive item: {name}")
    read = await file.read()
    hash = hashlib.sha256(read).hexdigest()

    #Save image
    with open(os.path.join("images",f"{hash}.jpg"),'wb') as f:
        f.write(read)
    dict = {"name":name, "category":category, "image_filename": f"{hash}.jpg"}

    #Insert into database
    Insert_item("items", dict)

    return {"message": f"item received: {name}"}

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

if __name__ == '__main__':
    uvicorn.run("main:app",port=9000,reload=True)

