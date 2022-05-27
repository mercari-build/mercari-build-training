import collections
import os
import logging
import pathlib
import shutil
from fastapi import FastAPI, Form, HTTPException
from fastapi import FastAPI, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import random
import sqlite3
import json
from fastapi.responses import ORJSONResponse
import hashlib
from .openCV import condition 

#----config----------------------------

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
image = pathlib.Path(__file__).parent.resolve() / "image"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
url = 'http://localhost:9000'
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
        d['image_filename'] = row[3]
        objects_list.append(d)   
    return objects_list 

def image_toHash(image_filename):
    image_name, image_fmt = map(str, image_filename.split('.'))
    image_hashname = hashlib.sha256(image_name.encode()).hexdigest()
    return '.'.join([image_hashname, image_fmt])

def save_image(file_location, image_file):
    with open(file_location, 'w+b') as f:
        shutil.copyfileobj(image_file.file, f) 
        
def add_sql(name,category, image_name):
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    c.execute("INSERT INTO items(name,category,image_filename) VALUES( ?, ?, ?);", (name,category,image_name))
    # idはtable作成時に割り当て済み
    conn.commit()
    id = c.execute('SELECT id FROM items WHERE image_filename = ? ;', (image_name,)).fetchone()
    conn.close()
    return id
 
def add_checkDB(id, score, checked_image_name):
    conn = sqlite3.connect("../db/check.db", check_same_thread=False)
    c = conn.cursor()
    c.execute("INSERT INTO items(name,category,image_filename) VALUES( ?, ?, ?);", (id, score, checked_image_name))
    conn.commit()
    conn.close()
    
# ----endpoints--------------------------

@app.get("/")
def root():
    return {"message": "Hello, world!"}

# curl -X GET 'http://127.0.0.1:9000/items'
@app.get("/items", response_class=ORJSONResponse)
def show_item():
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    items = c.execute('SELECT * FROM items;').fetchall()
    content = db_toList(items)
    conn.close()
    return {"items": content}

# curl -X POST \
#   --url 'http://localhost:9000/items' \
#   -d 'name=jacket' \
#   -d 'category=fashion' \
#   -d 'image=images/default.jpg'
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    #OpenCV.py
    score, checked_image = condition(image)
    save_image(f"checked_image/{checked_image.filename}", checked_image)
    
    #imageNameハッシュ化
    image_hashname = image_toHash(image.filename) 
    file_location = f"image/{image_hashname}"
    save_image(file_location,image)
    logger.info(f"Receive item: {image_hashname}") 
    
    #DB
    id = add_sql(name,category,image_hashname)
    add_checkDB(id, score, checked_image)
    return {"message": f"item received: {name}"}


# curl -X GET 'http://127.0.0.1:9000/search?keyword=jacket'
@app.get("/search" , response_class=ORJSONResponse)
def search_item(keyword: str = None):
    conn = sqlite3.connect('../db/item.db')
    c = conn.cursor()
    items = c.execute('SELECT * FROM items WHERE name LIKE  ? ;', (f"%{keyword}%",)).fetchall()
    content = db_toList(items)
    conn.close()
    return {"items": content}



# curl -X GET 'http://127.0.0.1:9000/items/(id)'
# {"items":[{"id":1,"name":"jacket","category":"fashion","image":"ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg"}]}
@app.get("/items/{item_id}", response_class=ORJSONResponse)
def show_detailById(item_id: int):
    logger.info(f"Search item: {item_id}")
    conn = sqlite3.connect('../db/item.db')
    c = conn.cursor()
    items = c.execute("SELECT * from items WHERE id=(?)", (item_id,)).fetchone()
    content = db_toList(items)
    conn.close()
    return {"items": content}



@app.get("/image/{image_filename}")
def get_image(image_filename):
    # Create image path
    logger.info(f"image_file:{image_filename}")

    image_path =  image / image_filename
    #..../mercari/mercari-build-training-2022/python/image/undefinedが返ってくる
    logger.info(f"image_location::{image_path}")
    
    if not image_filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
    

    if not image_path:
        logger.debug(f"Image not found: {image_filename}")
        image_path = images / "default.jpg"
    
        
    return FileResponse(image_path)


@app.get("/check/{id}")
def get_checked(id):
    conn = sqlite3.connect('../db/check.db')
    c = conn.cursor()
    score = c.execute('SELECT score FROM score WHERE id =  ? ;', (id),).fetchone() 
    conn.close()
    return score
