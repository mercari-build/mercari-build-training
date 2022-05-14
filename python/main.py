import collections
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi import FastAPI, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import random
import sqlite3
import json
from fastapi.responses import ORJSONResponse
import hashlib


app = FastAPI()


# @app.post("/files/")
# async def create_file(file: bytes = File(...)):
#     return {"file_size": len(file)}


# @app.post("/uploadfile/")
# async def create_upload_file(file: UploadFile):
#     return {"filename": file.filename}




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
        #d['id'] = row[0]
        d['name'] = row[1]
        d['category'] = row[2]
        d['image'] = row[3]
        objects_list.append(d)   
    # return json.dumps(items)
    return objects_list 

def image_toHash(image):
    # path = "../image/" + str(image)
    
    # image = image.filename()
    # print(image)
    #file_location = f'image/{image}'
    # os.path.splitext(image)
    #print(file_location)
    with open(image, 'rb') as f:
        # f.seek(0)
        sha256 = hashlib.sha256(f.read()).hexdigest()
        f.close()
        # print('SHA256ハッシュ値：\n {0}'.format(sha256))
        return sha256
    

def add_sql(name,category, image_name):
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    c.execute("INSERT INTO items(name,category,image) VALUES( ?, ?, ?);", (name,category,image_name))
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
#   -d 'image=image/default.jpg'
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: bytes = File(...)):
    image_name = image_toHash(image) + ".jpg"
    file_location = f"images/{image_name}"
    with open(file_location, mode="wb") as f:
        f.write(image)
        f.close()
    logger.info(f"Receive item: {name}")
    logger.info(f"Receive item: {category}")
    logger.info(f"Receive item: {image_name}") 
    #id = random.randint(1,1000)
    add_sql(name,category,image_name)
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
@app.get("/items/", response_class=ORJSONResponse)
def show_detailById(item_id: int):
    conn = sqlite3.connect('../db/item.db')
    c = conn.cursor()
    items = c.execute('SELECT * FROM items WHERE id = ? ;', (item_id)).fetchall()
    content = db_toList(items)
    conn.close()
    return {"items": content}


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
