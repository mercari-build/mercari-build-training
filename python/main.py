import os
import logging
import pathlib
import hashlib
import json
from fastapi import FastAPI, Form, HTTPException, UploadFile, Query
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
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


@app.get("/")
def root():
    return {"message": "Hello, world!"}

# @app.post("/items")
# def add_item(name: str = Form(...), category: str=Form(...), image: UploadFile = Form(...)):
#     logger.info(f"Receive item: {name}")
#     logger.info(f"Receive item: {category}")
#     logger.info(f"Receive item: {image}")

#     image_name = image.filename
#     hashed_image_name = get_hash(image_name)
#     save_image(image, hashed_image_name)

#     message = {
#         "items":[
#             {
#                 "name": name,
#                 "category": category,
#                 "image_name": hashed_image_name
#             }
#         ]
#     }

#     return message

# def get_hash(image):
#     hash = hashlib.sha256(image.encode()).hexdigest()
#     return hash+".jpg"

# def save_image(image,jpg_hashed_image_name):
#     imagefile = image.file.read()
#     image = images / jpg_hashed_image_name
#     with open(image, 'wb') as f:
#         f.write(imagefile)
#     return


# @app.get("/items")
# def get_item():
#     f = open("items.json")
#     data = json.load(f)

#     return data    


@app.get("/image/{image_name}")
async def get_image(image_name):
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
class Item(BaseModel):
    id: int
    name: str
    category_id: int
    image_name: str

class Category(BaseModel):
    id: int
    name: str

def create_or_update_schema():
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()


    cursor.execute('''
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY,
            name TEXT,
            category_id INTEGER,
            image_name TEXT,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        )
    ''')
    

    cursor.execute('''
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY,
            name TEXT
        )
    ''')
    
    conn.commit()
    conn.close()


@app.get('/items')
def get_items():
    # conn = sqlite3.connect('mercari.sqlite3')
    # cursor = conn.cursor()
    # cursor.execute('SELECT * FROM items')
    # items = cursor.fetchall()
    # conn.close()
    # return items
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute('''
        SELECT items.id, items.name, categories.name, items.image_name
        FROM items
        INNER JOIN categories ON items.category_id = categories.id
    ''')
    items = cursor.fetchall()
    conn.close()
    return [{'id': item[0], 'name': item[1], 'category': item[2], 'image_name': item[3]} for item in items]

@app.post('/items')
def add_item(item: Item):
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()

    with open('items.json', 'r') as file:
        items = json.load(file)

    for item in items:
        cursor.execute('''
            INSERT INTO items (name, category_id, image_name) 
            VALUES (?, ?, ?)
        ''', (item['name'], item['category_id'], item['image_name']))
    # cursor.execute('INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)',
    #                (item.name, item.category, item.image_name))
    conn.commit()
    conn.close()
    return {'message': 'Item added successfully'}

@app.get('/search')
def search_items(keyword: str = Query(...)):
    conn = sqlite3.connect('mercari.sqlite3')
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM items WHERE name LIKE ? OR category LIKE ?", ('%'+keyword+'%', '%'+keyword+'%'))
    items = cursor.fetchall()
    conn.close()
    if not items:
        raise HTTPException(status_code=404, detail="No items found matching the keyword.")
    return items

if __name__ == '__main__':
    import uvicorn
    create_or_update_schema()
    uvicorn.run(app, host='127.0.0.1', port=9000)




