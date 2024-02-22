import os
import hashlib
import json
import logging
import pathlib
import sqlite3
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.parent.resolve() / "db"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

#Reset the table everytime virtual environment is re-built
con = sqlite3.connect(db / "mercari.sqlite3") #create connection object
cur = con.cursor() #create cursor
cur.execute('''DROP TABLE IF EXISTS items''')
cur.execute('''DROP TABLE IF EXISTS categories''')
con.commit()
con.close()

# Function: Create table if it doesn't exist yet
def create_tables():
    con = sqlite3.connect(db / "mercari.sqlite3") #create connection object
    cur = con.cursor() #create cursor
    cur.execute('''CREATE TABLE IF NOT EXISTS items 
             (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, category_id INTEGER, image_name TEXT)''')

    cur.execute('''CREATE TABLE IF NOT EXISTS categories 
                 (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)''')
    con.commit()
    con.close()

@app.get("/")
def root():
    logger.info("Saying hello to the world")
    return {"message": "Hello, world!"}

@app.get("/items")
def get_items():
    try:
        con = sqlite3.connect(db / "mercari.sqlite3")
        cur = con.cursor()
        cur.execute("SELECT items.id, items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id")
        items_data = []
        for row in cur.fetchall():
            items_data += [{"id": row[0], "name": row[1], "category": row[2], "image_name": row[3]}]
        con.close()
        logger.info(f"Receive items: {items_data}")
        return items_data
    except sqlite3.Error as error:
        logger.error(f"Error occured: {error}")
        raise HTTPException(status_code=500, detail="Internal Server Error")

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    try:
        #Hash
        image_bytes = await image.read()
        image_hash = hashlib.sha256(image_bytes).hexdigest()
        image_name = f"{image_hash}.jpg"
        image_path = images / image_name
        with open(image_path, 'wb') as f:
            f.write(image_bytes)
        create_tables() #create table if it deosn't exist
        con = sqlite3.connect(db / "mercari.sqlite3")
        cur = con.cursor()
        cur.execute("SELECT id FROM categories WHERE name = ?", (category,))
        category_row = cur.fetchone()

        if category_row == None:
            cur.execute("INSERT INTO categories (name) VALUES (?)", (category,))
            con.commit()
            category_id = cur.lastrowid
        else:
            category_id = category_row[0]

        cur.execute("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)", (name, category_id, image_name))
        con.commit()
        con.close()

        logger.info(f"Receive item: {name}, category_id: {category_id}, category: {category}, image: {image_name}")
        return {"message": f"item received: {name}, Category: {category}"}
    except sqlite3.Error as sqlerror:
        logger.error(f"SQLite error occurred: {sqlerror}")
        raise HTTPException(status_code=500, detail=f"SQL Error: {sqlerror}")
    except Exception as error:
        logger.error(f"An unexpected error occured. Error: {error}")
        raise HTTPException(status_code=500, detail=f"Error: {error}")



@app.get("/image/{image_name}")
async def get_image(image_name):
    logger.info(f"Receive image: {image_name}")
    # Create image path
    image_path = images / image_name

    if not image_name.endswith(".jpg"):
        logger.error(f"Image path does not end with .jpg")
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg Make sure the file name is correct")

    elif not image_path.exists():
        logger.error(f"Image not found: {image_name}")
        image_path = images / "default.jpg"

    return FileResponse(image_path)

@app.get("/items/{item_id}")
def get_item(item_id: int):
    try:
        with open('items.json', 'r') as f:
            items_data = json.load(f)
        if 1 <= item_id <= len(items_data["items"]):
            item = items_data["items"][item_id - 1]
            logger.info(f"Access item: {item_id}")
            return item
        else:
            logger.error(f"Invalid item ID: {item_id}")
            raise HTTPException(status_code=404, detail="Item not found (Invalid ID)")
    except FileNotFoundError: #if items.json is not found
        logger.error(f"File not found")
        raise HTTPException(status_code=500, detail="Internal Server Error")
    
#STEP 4.2
@app.get("/search")
def get_items(keyword: str):
    try:
        con = sqlite3.connect(db / "mercari.sqlite3")
        cur = con.cursor()
        cur.execute("SELECT name, category_id FROM items WHERE name LIKE ?", ('%' + keyword + '%',))
        items_data = {"items": []}
        for row in cur.fetchall():
            items_data["items"].append({"name": row[0], "category": row[1]})
        con.close()
        logger.info(f"Items under the name {keyword}: {items_data}")
        return items_data
    except sqlite3.Error as sqlerror:
        logger.error(f"SQLite error occurred: {sqlerror}")
        raise HTTPException(status_code=500, detail=f"SQL Error: {sqlerror}")
    except Exception as error:
        logger.error(f"Error occured: {error}")
        raise HTTPException(status_code=500, detail=f"Error: {error}")
