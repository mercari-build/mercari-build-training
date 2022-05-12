import collections
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import random
import sqlite3
import json


#----DB-------------------------------
# # open DB
# conn = sqlite3.connect("../db/item.db", check_same_thread=False)
# c = conn.cursor()

# # make table
# c.execute("DROP TABLE 'items'")
# c.execute("CREATE TABLE `items` (id int, name string,category string);")

# # commit changes
# conn.commit()


#----config----------------------------

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

# ----methods----------------------------

def db_toList(items):
    objects_list = []
    for row in items:
        d = collections.OrderedDict()
        # d['id'] = row[0]
        d['name'] = row[1]
        d['category'] = row[2]
        objects_list.append(d)   
    # return json.dumps(items)
    return objects_list 


def add_sql(id,name,category):
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    c.execute("INSERT INTO items(id,name,category) VALUES(?,?,?);", (id,name,category))
    conn.commit()
    conn.close()

    
# ----endpoints--------------------------

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def show_item():
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    items = c.execute("SELECT * FROM items;").fetchall()
    content = db_toList(items)
    conn.close()
    return {"items": f"{json.dumps(content)}"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Receive item: {name}")
    logger.info(f"Receive item: {category}")
    id = random.randint(1,100)
    add_sql(id,name,category)
    return {"message": f"item received: {name}"}

# 'http://127.0.0.1:9000/search?keyword=jacket'
@app.get("/search")
def search_item(keyword: str = None):
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    items = c.execute("SELECT * FROM items WHERE name LIKE '%' + ? + '%';", [keyword]).fetchall()
    print("a" + "".join(items))
    content = db_toList(items)
    conn.close()
    return {"items": f"{json.dumps(content)}"}

@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path
    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)
