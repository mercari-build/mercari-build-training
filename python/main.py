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


@app.post("/files/")
async def create_file(file: bytes = File(...)):
    return {"file_size": len(file)}


@app.post("/uploadfile/")
async def create_upload_file(file: UploadFile):
    return {"filename": file.filename}


#----DB-------------------------------
# # open DB
# conn = sqlite3.connect("../db/item.db", check_same_thread=False)
# c = conn.cursor()

# # make table
# c.execute("DROP TABLE 'items'")
# c.execute("CREATE TABLE `items` (id int, name string,category string, image string);")

# # commit changes
# conn.commit()


#----config----------------------------

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

# ----methods----------------------------

def db_toList(items):
    objects_list = []
    for row in items:
        d = collections.OrderedDict()
        # d['id'] = row[0]
        d['name'] = row[1]
        d['category'] = row[2]
        d['image'] = row[3]
        objects_list.append(d)   
    # return json.dumps(items)
    return objects_list 

def image_toHash(image):
    with open(image, 'rb') as f:
        # f.seek(0)
        sha256 = hashlib.sha256(f.read()).hexdigest()
        print('SHA256ハッシュ値：\n {0}'.format(sha256))
        return sha256
    

def add_sql(id,name,category, image_hashname):
    conn = sqlite3.connect("../db/item.db", check_same_thread=False)
    c = conn.cursor()
    c.execute("INSERT INTO items(id,name,category,image) VALUES(?,?,?, ?);", (id,name,category,image_hashname))
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
def add_item(name: str = Form(...), category: str = Form(...), image: bytes = File(...)):
    image_hashname = image_toHash(image) + ".jpg"
    file_location = f"image/{image_hashname}"
    with open(file_location, mode="wb") as f:
        f.write(image)
        f.close()
    logger.info(f"Receive item: {name}")
    logger.info(f"Receive item: {category}")
    logger.info(f"Receive item: {image_hashname}") 
    id = random.randint(1,1000)
    add_sql(id,name,category,image_hashname)
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

# {"items":[{"id":89,"name":"jacket","category":"fashion","image":"ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg"}]}
# curl -X GET 'http://127.0.0.1:9000/items/89'
@app.get("/items/", response_class=ORJSONResponse)
def show_detailById(item_id: int):
    conn = sqlite3.connect('../db/item.db')
    c = conn.cursor()
    items = c.execute('SELECT * FROM items WHERE id = ? ;', (item_id)).fetchall()
    content = db_toList(items)
    conn.close()
    return {"items": content}




@app.get("/image/{image_filename_hash}")
async def get_image(image_filename_hash):

    # Create image path
    image = images / image_filename_hash

    if not image_filename_hash.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image_filename_hash}")
        image = images / "default.jpg"

    return FileResponse(image)
