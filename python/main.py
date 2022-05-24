import os
import logging
import pathlib
import sqlite3
import hashlib
from fastapi import FastAPI, Form, HTTPException, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "image"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {"message": "Welcome to Mercari's Items Database Made by Momoe"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Receive item: {name}, category: {category}, image filename: {image.filename}")

    if not image.filename.endswith(".jpg"):
        raise HTTPException(
            status_code=400, detail="Image is not in .jpg format")
        
    image_file_hash = hashlib.sha256(image.filename.split('.')[0].encode("utf-8")).hexdigest() + '.jpg'

    # connect to database
    conn = sqlite3.connect("../db/mercari.sqlite3")
    c = conn.cursor()
    
    # insert category name into category table or ignore if duplicate
    c.execute("INSERT OR IGNORE INTO category(name) VALUES (?)", (category,))

    # retrieve id from category table using category name
    category_id = c.execute("SELECT id FROM category WHERE name = (?)", (category,)
            ).fetchone()[0]

    # insert into items table
    c.execute("INSERT INTO items(name, category_id, image) VALUES (?, ?, ?)",
                (name, category_id, image_file_hash))

    conn.commit()
    conn.close()

    # --- using json ---
    # items_list = {"items": []}
    # check if json file exists
    # if os.path.isfile("items.json"):
    #     with open("items.json", "r") as items_json_f:
    #         # load existing data
    #         items_list = json.load(items_json_f)
    
    # add new item
    # new_item = {"name" : name, "category": category}
    # items_list["items"].append(new_item)
    
    # with open("items.json", "w") as items_json_f:
    #     # write new data to json file
    #     json.dump(items_list, items_json_f)

    return {"message": f"item received: {name}, category: {category}, image filename: {image_file_hash}"}

@app.get("/items")
def get_items():
    # connect to database
    conn = sqlite3.connect("../db/mercari.sqlite3")
    c = conn.cursor()

    result = c.execute(
        """
            SELECT
                items.id, items.name, category.name, items.image
            FROM 
                items
                INNER JOIN category 
                    ON items.category_id = category.id
        """
    ).fetchall()

    conn.close()

    items_list = {
        "items": [{"id": id, "name": name, "category": category, "image": image} for (id, name, category, image) in result]
    }

    # --- using json ---
    # items_list = {"items" : []}
    # if os.path.isfile("items.json"):
    #     with open("items.json", "r") as items_json_f:
    #         items_list = json.load(items_json_f)

    return items_list

@app.get("/search")
def search_item(keyword: str):
    conn = sqlite3.connect("../db/mercari.sqlite3")
    c = conn.cursor()

    result = c.execute(
        """
            SELECT 
                items.id, items.name, category.name, items.image
            FROM 
                items 
                INNER JOIN category
                    ON items.category_id = category.id
            WHERE 
                items.name LIKE (?)
        """, 
        (f"%{keyword}%",),
    ).fetchall()

    conn.close()

    if result == []:
        message = {"message": "No matching item"}
    else:
        message = {
            "items": [{"id": id, "name": name, "category": category, "image": image} for (id, name, category, image) in result]
        }

    return message

@app.get("/items/{item_id}")
def get_item(item_id: int):
    conn = sqlite3.connect("../db/mercari.sqlite3")
    c = conn.cursor()

    result = c.execute(
        f"""
            SELECT 
                items.id, items.name, category.name, items.image
            FROM 
                items 
                INNER JOIN category
                    ON items.category_id = category.id
            WHERE 
                items.id = {item_id}
        """
    ).fetchall()

    conn.close()

    if result is None:
        message = {"message": "No matching item"}
    else:
        message = {
            "items":[{"id": id, "name": name, "category": category, "image": image} for (id, name, category, image) in result]
        }

    return message

@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path
    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.info(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
