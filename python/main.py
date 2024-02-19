import os
import logging
import pathlib
import hashlib
import json
from fastapi import FastAPI, Form, HTTPException, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

os.chdir('/Users/xiaotongye/Programs/mercari-build-training/python')

# 3-6 Understand Loggers
logging.basicConfig(level=logging.DEBUG)


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

def load_items_json():
    items_path = pathlib.Path(__file__).parent.resolve() / "items.json"
    if os.path.exists(items_path):
        with open(items_path, "r") as file:
            try:
                data = json.load(file)
                items, cur_max_id = data['items'], data['cur_max_id']
            except:
                items, cur_max_id = {}, 0
    else:
        items, cur_max_id = {}, 0
    print(items, cur_max_id)
    return items, cur_max_id

def save_items_json(items, cur_max_id):
    items_path = pathlib.Path(__file__).parent.resolve() / "items.json"
    with open(items_path, "w") as file:
        json.dump({'items': items, 'cur_max_id': cur_max_id},file,indent=2)
    return

@app.get("/")
def root():
    return {"message": "Hello, world!"}

# 3-3 Get a list of items
@app.get("/items")
def get_item():
    items, cur_max_id = load_items_json()
    items_list =  \
        [{"name": item["name"], "category": item["category"]} for item in items.values()]
    return items_list

# 3-5 Index a item by id
@app.get("/items/{item_id}")
def get_item(item_id):
    items, cur_max_id = load_items_json()
    if item_id in items:
        return items[item_id]
    else:
        return {"message": "The id does not exist!"}

# 3-4 Add a new item
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}")
    items, cur_max_id = load_items_json()
    new_item = {"name":name, "category":category}
    
    imgBytes = image.file.read()
    hashedImgName = hashlib.sha256(imgBytes).hexdigest() + os.path.splitext(image.filename)[1]
    hashedImgPath = images / hashedImgName
    hashedImgPath.write_bytes(imgBytes)

    new_item = {"name": name, "category": category, "image": hashedImgName}
    cur_max_id += 1
    items[cur_max_id] = new_item
    save_items_json(items, cur_max_id)
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
