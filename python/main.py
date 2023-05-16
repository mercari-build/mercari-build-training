import json
import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3


app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get('FRONT_URL', 'http://localhost:3000')]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)
sqlpath = pathlib.Path(__file__).parent.parent.resolve() / "db" / "mercari.sqlite3"
db_setup_path = pathlib.Path(__file__).parent.parent.resolve() / "db" / "items.db"
conn = sqlite3.connect(sqlpath)
with open(db_setup_path, 'r') as f:
    db = f.read()
cursor = conn.cursor()
cursor.executescript(db)
conn.commit()
conn.close()
@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def get_item():
    conn = sqlite3.connect(sqlpath)
    cursor = conn.cursor()
    cursor.execute(("""
        SELECT items.id, items.name, items.category_id, items.image_name, category.name 
        FROM items
        INNER JOIN category ON items.category_id = category.id"""))
    items = cursor.fetchall()
    result = []
    for i in items:
        item = {
            "id": i[0],
            "name": i[1],
            "category": i[4],
            "image_name": i[3]
        }
        result.append(item)
    return {"items": result}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Received item: {id}, Receive item: {name}, Receive image:{image.filename}")

    # Hash the image using sha256, and save it with the name <hash>.jpg
    file = image.file.read()
    image_hash = hashlib.sha256(file).hexdigest()
    filename = image_hash + ".jpg"
    path = images / filename
    with open(path, "wb") as f:
        f.write(file)

    # Add new items into items.db
    conn = sqlite3.connect(sqlpath)
    cursor = conn.cursor()
    cursor.execute("SELECT id FROM category WHERE name = ? ", (category,))
    category_id = cursor.fetchone()
    if category_id is None:
        cursor.execute("INSERT INTO category (name) VALUES (?)", (category,))
        cursor.execute("SELECT id FROM category WHERE name = ?", (category,))
        category_id = cursor.fetchone()[0]
    else:
        category_id = category_id[0]
    cursor.execute("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
                   (name, category_id, filename,))
    conn.commit()
    conn.close()
    return {"message": f"items received: {name}, items received: {category_id}, items received: {filename}, category received{category}"}

@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.get("/items/{item_id}")
def get_itemsid(item_id:int):
    try:
        with open("items.json", "r") as f:
            mydata = json.load(f)
            return mydata["items"][item_id]
    except IndexError:
        raise HTTPException(
            status_code=404, detail=f"item_id {item_id} not exist"
        )
@app.get("/search")
def get_keyword(keyword: str):
    conn = sqlite3.connect(sqlpath)
    cursor = conn.cursor()
    cursor.execute("""SELECT items.name, category.name 
                    FROM items
                    INNER JOIN category ON items.category_id = category.id
                    WHERE items.name = ?""", (keyword,))
    items = cursor.fetchall()
    conn.close()
    result = []
    for i in items:
        item = {
            "name": i[0],
            "category": i[1]
        }
        result.append(item)
    return {"items": result}


