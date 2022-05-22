# uvicorn main:app --reload --port 9000 で起動
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

import sqlite3
import hashlib
from fastapi import UploadFile, File
import shutil

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)
api_url = os.environ.get('API_URL', 'http://localhost:9000')

sqlite_path = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"

# def get_items_json():
#     with open('items.json', mode='r', encoding='utf-8') as f:
#         items_json = json.load(f)
#     return items_json

# def post_items_json(name, category) -> None:
#     items_json = get_items_json()
#     items_json["items"].append({"name":name, "category":category})
#     with open('items.json', mode='w') as f:
#         json.dump(items_json, f)

def get_hash_name(image):
    image_name, image_exp = image.split('.')
    image_hashed = hashlib.sha256(image_name.encode()).hexdigest()
    return '.'.join([image_hashed, image_exp])

def save_image(image_name, image) -> None:
    with open(images / image_name, mode='w+b') as f:
        shutil.copyfileobj(image, f)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Receive item: {name}, {category}, {image.filename}")

    conn = sqlite3.connect(sqlite_path)
    cursor = conn.cursor()

    image_hashed = get_hash_name(image.filename)

    save_image(image_hashed, image.file)

    cursor.execute('SELECT id FROM category WHERE name=(?)', (category, ))
    id = cursor.fetchone()

    if not id:
        cursor.execute("INSERT OR IGNORE INTO category(name) VALUES (?)", (category, ))
    if id:
        cursor.execute("INSERT OR IGNORE INTO category(id, name) VALUES (?, ?)", (id[0], category))

    conn.commit()
    cursor.execute("SELECT id FROM category WHERE name=(?)", (category, ))
    category_id = cursor.fetchone()
    category_id = category_id[0]

    cursor.execute("INSERT INTO items(name, category_id, image_filename) VALUES(?, ?, ?)", (name, category_id, image_hashed))
    conn.commit()
    conn.close()

    return {"message": f"item received: {name}, {category}"}

@app.get("/items")
def get_items():
    conn = sqlite3.connect(sqlite_path)
    cursor = conn.cursor()
    cursor.execute("SELECT name, category_id, image_filename FROM items")
    items_json = cursor.fetchall()
    items_get_data = {"items": []}

    for i in range(len(items_json)):
        cursor.execute("SELECT name FROM category WHERE id=(?)", (items_json[i][1], ))
        category = cursor.fetchone()[0]
        items_get_data["items"].append({"name":items_json[i][0], "category":category, "image":os.path.join(api_url, 'images', items_json[i][2])})

    conn.close()
    return items_get_data

@app.get("/items/{item_id}")
def get_item_id(item_id):
    conn = sqlite3.connect(sqlite_path)
    cursor = conn.cursor()
    cursor.execute("SELECT name, category_id, image_filename FROM items WHERE id=(?)", (item_id))
    data = cursor.fetchone()
    cursor.execute("SELECT name FROM category WHERE id=(?)", (data[1], ))
    category = cursor.fetchone()[0]
    conn.close()
    return {"name": data[0], "category": category, "image_filename": data[2]}

@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    print(FileResponse(image))

    return FileResponse(image)

