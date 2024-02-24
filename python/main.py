import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3

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


'''
global variables
database_file : sqlite3 database file location.

Notes: given sqlite module version is not thread safe according to sqlite3.threadsafety.
'''

database_file = "../db/mercari.sqlite3"

@app.get("/")
def root():
    return {"message": "Hello, world!"}

def get_items_from_db():
    global database_file
    con = sqlite3.connect(database_file)
    cur = con.cursor()
    res = cur.execute("SELECT items.id, name, category, image_name FROM items inner join categories on items.category_id = categories.id")
    items = res.fetchall()
    cur.close()
    con.close()
    return items

@app.get("/items")
def get_items():
    items = get_items_from_db()
    return {"items": items}

@app.get("/items/{item_id}")
def get_item(item_id):
    global database_file
    con = sqlite3.connect(database_file)
    cur = con.cursor()
    res = cur.execute("SELECT items.id, name, category, image_name FROM items inner join categories on items.category_id WHERE items.id = ?", [int(item_id)])
    item = res.fetchone()
    cur.close()
    con.close()
    return item

@app.get("/search")
def search_items(keyword: str):
    global database_file
    con = sqlite3.connect(database_file)
    cur = con.cursor() 
    res = cur.execute("SELECT items.id, name, category, image_name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE name LIKE ? OR category LIKE ?", [keyword, keyword])
    found_items = res.fetchall()
    cur.close()
    con.close()
    return found_items
    

@app.post("/items")
def add_item(name: str = Form(...), category_id: int = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}")

    file_content = image.file.read()
    image.file.seek(0)

    image_hash = hashlib.sha256(file_content).hexdigest()
    save_image(file_content, f"{image_hash}.jpg")

    global database_file
    con = sqlite3.connect(database_file)
    cur = con.cursor()
    data = [name, category_id, f"{image_hash}.jpg"]
    res = cur.execute("INSERT INTO items VALUES(NULL,?,?,?)", data)
    con.commit()
    items = get_items_from_db()
    cur.close()
    con.close()
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
