
import os
import logging
import pathlib

import hashlib

import sqlite3


DATABASE = 'mercari.sqlite3'


def create_table():
    con = sqlite3.connect(DATABASE)
    cur = con.cursor()
    cur.execute('''CREATE TABLE IF NOT EXISTS categories (
                   id INTEGER PRIMARY KEY,
                   name TEXT NOT NULL)''')
    cur.execute('''CREATE TABLE IF NOT EXISTS items (
                   id INTEGER PRIMARY KEY, 
                   name TEXT NOT NULL,
                   category_id INTEGER NOT NULL,
                   image_name TEXT NOT NULL,
                   FOREIGN KEY (category_id) REFERENCES categories(id))''' )

    con.commit()
    con.close()
    print("Tables created")

#STEP4-1
from fastapi import FastAPI, Form, File, HTTPException, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

create_table()

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

#STEP3-3, 4-1
@app.get("/items")
def get_items_from_database():

    con = sqlite3.connect(DATABASE)
    cur = con.cursor()

    cur.execute("SELECT name, category_id, image_name FROM items")
    items = cur.fetchall()

    cur.close()
    con.close()

    items_list = [{"name": name, "category": get_category_name(category_id), "image_name": image_name} for name, category_id, image_name in items]

    print("Items list", items_list)

    return {"items": items_list}

def get_category_name(category_id):
    con = sqlite3.connect('mercari.sqlite3')
    cur = con.cursor()
    cur.execute('SELECT name FROM categories WHERE id =?', (category_id,))
    category_name = cur.fetchone()[0]
    con.close()
    return category_name

    
#STEP3-2, 3-4, 4-1, 4-3
@app.post("/items")
def add_item(name: str = Form(...), category_name: str = Form(...), image: UploadFile = File(...)):

    #STEP4-1
    image_filename = get_image_filename(image)

    con = sqlite3.connect(DATABASE)
    cur = con.cursor()

    cur.execute ("SELECT id FROM categories WHERE name = ?", (category_name,)) 
    category_row = cur.fetchone()

    if category_row == None:
        cur.execute("INSERT INTO categories (name) VALUES (?)", (category_name,))
        con.commit()
        category_id = cur.lastrowid
    else:
        category_id = category_row[0]

    cur.execute ("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",(name, category_id, image_filename))

    con.commit()
    con.close()

    return {"message": f"Item added: {name}, {category_name}, {image_filename}"}

#STEP3-4
def get_image_filename(image):
    image_contents = image.file.read()
    image_hash = hashlib.sha256(image_contents).hexdigest()

    #Create a file path
    image_filename = f"{image_hash}.jpeg"
    save_path = os.path.join("images", image_filename)

    #Save a image
    with open(save_path, "wb") as f:
        f.write(image_contents)

    logger.info(f"Saved image to: {save_path}")

    return image_filename

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

#STEP3-5, 4-1, 4-3
@app.get("/items/{item_id}")
def get_item_information(item_id: int):

    con = sqlite3.connect(DATABASE)
    cur = con.cursor()

    cur.execute ("SELECT items.id, items.name, categories.name as category, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE items.id = ?", (item_id,))
    result = cur.fetchone()

    con.commit()
    con.close()
    

    if result:
        return {"id": result["id"], "name": result["name"], "category": result["category"], "image_name": result["image_name"]}
    else:
        return{"detail": "Item not found"}
    
#STEP4-2,4-3
@app.get("/search")
def search_items(keyword: str):
    print(keyword)
    con = sqlite3.connect(DATABASE)
    cur = con.cursor()

    res = cur.execute("SELECT items.id, items.name, categories.name, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?", ("%" + keyword + "%",))

    found_items = res.fetchall()
    cur.close()
    con.close()

    print("Search results", found_items)
    return found_items

 



