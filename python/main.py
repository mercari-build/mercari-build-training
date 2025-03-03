import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends,File,UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json
from typing import List
import hashlib
from pathlib import Path

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"


def get_db():
    # if not db.exists():
    #     yield

    conn = sqlite3.connect(db)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    
    print("set up db!!")
    # connect database
    db_gen = get_db()  # get_db()を呼び出してジェネレータを取得
    conn = next(db_gen)  # ジェネレータを消費してデータベース接続を取得
    
    cursor = conn.cursor()

    # open sql schema　file
    with open('db/items.sql', 'r') as f:
        sql_script = f.read()
        cursor.executescript(sql_script)
    
    conn.commit()
    db_gen.close()


@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

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


class HelloResponse(BaseModel):
    message: str


@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})


# add category
class Item(BaseModel):
    name: str
    category:str = "default_category"


# changed to return item list
class GetItemResponse(BaseModel):
    items: List[Item]

""""
# get endpoint for 4-3
@app.get("/items", response_model=GetItemResponse)
def get_item():
    with open("item.json","r") as f:
        data = json.load(f)

    resdata = []

    for i in data["items"]:
        ele = Item(name = i["name"],category = i["category"])
        resdata.append(ele)

    resjson = GetItemResponse(items = resdata)
    return  resjson
"""

# add image_name
class GetItemimage(BaseModel):
    name: str
    category: str
    image_name: str

# changed to return item list
class GetimageItemResponse(BaseModel):
    items: List[GetItemimage]

"""
# get endpoint for 4-4
@app.get("/items",response_model = GetimageItemResponse)
def get_items():

    # open json file
    with open("item.json","r") as f:
        data = json.load(f)
    
    resdata = []
    for i in data["items"]:

        # image_name exist ver.
        if "image_name" in i:
            ele = GetItemimage(name=i["name"],category = i["category"],image_name= i["image_name"])
        # if element is Item class, image_name = nan
        else:
            ele = GetItemimage(name=i["name"],category = i["category"],image_name = "Nan")
        resdata.append(ele)

    # print("[CHECK!!!!!]",resdata)
    
    resjson = GetimageItemResponse(items = resdata)

    return resjson
"""

# add index for 4-5
@app.get("/items/{ind}",response_model = GetItemimage)
def get_item(ind: int):
    print("this is index",ind)

    # open json file
    with open("item.json","r") as f:
        data = json.load(f)
    
    
    # image_name exist ver.
    if "image_name" in data["items"][ind-1]:
        ele = GetItemimage(name=data["items"][ind-1]["name"],category = data["items"][ind-1]["category"],image_name= data["items"][ind-1]["image_name"])
    # if element is Item class, image_name = nan
    else:
        ele = GetItemimage(name=data["items"][ind-1]["name"],category = data["items"][ind-1]["category"],image_name = "Nan")

    return ele

class AddItemResponse(BaseModel):
    message: str



# add_item is a handler to add a new item for POST /items . (for 4-3)
"""
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")

    insert_item(Item(name=name))
    return AddItemResponse(**{"message": f"item received: {name}"})
"""


# change post for 4-4
"""
@app.post("/items",response_model = AddItemResponse)
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
# async def get_image(image_name):

    # print("kakuninn!!!",image)
    file_name = image.filename
    # Create image path
    image_path = images / file_name


    if not image.filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_path.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    # hashlibの引数はバイナリを読み込むので一度データを読み込む必要がある
    file_hash = hashlib.sha256()
    contents = await image.read()  # ファイルの内容を読み込む
    file_hash.update(contents)
    file_hash_name= file_hash.hexdigest()+".jpg"
    
    insert_data = GetItemimage(name=name,category=category,image_name=file_hash_name)
    
    insert_imageitem(insert_data)
    
    return AddItemResponse(**{"message": f"item received: {file_name}"})
"""

# 4-6
@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name
    print("kakuninn!!",image)

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.info(f"Image not found(info): {image}")
        # logger.debug(f"Image not found(debug): {image}")
        # print("画像が存在しなかった時には以下の処理が実行される。")
        image = images / "default.jpg"

    return FileResponse(image)


# insert element of image class
def insert_imageitem(image_item: GetItemimage):

    # open json file
    with open("item.json","r") as f:
        data = json.load(f)
    # print(item)

    # add new item
    data["items"].append({
         "name": image_item.name,
        "category": image_item.category,
        "image_name":image_item.image_name
      })
    
    # write json file
    with open("item.json", "w") as f:
        json.dump(data,f,indent=2)
    

# insert element of item class
def insert_item(item: Item):

    # open json file
    with open("item.json","r") as f:
        data = json.load(f)
    # print(item)

    # add new item
    data["items"].append({
         "name": item.name,
        "category": item.category
      })
    
    # write json file
    with open("item.json", "w") as f:
        json.dump(data,f,indent=2)



# ==================STEP5===============================================================

# changed get endpoint for 5-1
"""
@app.get("/items",response_model = GetimageItemResponse)
def get_items():

    # connect database
    db_gen = get_db()  # get_db()を呼び出してジェネレータを取得
    conn = next(db_gen)  # ジェネレータを消費してデータベース接続を取得
    cursor = conn.cursor()

    # fetch item table from database
    cursor.execute("SELECT * FROM items")
    items = cursor.fetchall()

    db_gen.close()


    # print(items)
    # items = [(1, 'jacket', 'fashion', 'ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg'), 
    # (2, 'pen', 'stationery', 'ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg'), 
    # (3, 'bottoms', 'fashion', 'ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg')]
    
    lst = []
    for i in items:
        if len(i) == 4:
            ele = GetItemimage(name=i[1],category = i[2],image_name= i[3])
        else:
            ele = GetItemimage(name=i[1],category = i[2],image_name= "Nan")
        lst.append(ele)

    resjson = GetimageItemResponse(items = lst)

    return resjson
"""

"""
# change post end point for 5-1
@app.post("/items",response_model = AddItemResponse)
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):

    # print("kakuninn!!!",image)
    file_name = image.filename
    # Create image path
    image_path = images / file_name


    if not image.filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_path.exists():
        logger.info(f"Image not found: {image}")
        image = images / "default.jpg"

    # hashlibの引数はバイナリを読み込むので一度データを読み込む必要がある
    file_hash = hashlib.sha256()
    contents = await image.read()  # ファイルの内容を読み込む
    file_hash.update(contents)
    file_hash_name= file_hash.hexdigest()+".jpg"

    # connect database
    db_gen = get_db()  # get_db()を呼び出してジェネレータを取得
    conn = next(db_gen)  # ジェネレータを消費してデータベース接続を取得
    cursor = conn.cursor()

    print("kakuninnshite!!!!")
    # insert to database
"""
    # cursor.execute("""INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)""", (name, category, file_hash_name))
"""
    # saved change data
    conn.commit()
    db_gen.close()
    
    return AddItemResponse(**{"message": f"item received: {name,category,file_name}"})
"""


# add search end point for 5-2
@app.get("/search",response_model = GetimageItemResponse)
def search_items(keyword:str):

    # connect database
    db_gen = get_db()  # get_db()を呼び出してジェネレータを取得
    conn = next(db_gen)  # ジェネレータを消費してデータベース接続を取得
    cursor = conn.cursor()

    # fetch item table from database
    cursor.execute("SELECT * FROM items WHERE name = ?",(keyword,))
    items = cursor.fetchall()

    db_gen.close()
    
    lst = []
    for i in items:
        if len(i) == 4:
            ele = GetItemimage(name=i[1],category = i[2],image_name= i[3])
        else:
            ele = GetItemimage(name=i[1],category = i[2],image_name= "Nan")
        lst.append(ele)

    resjson = GetimageItemResponse(items = lst)

    return resjson

# changed get endpoint for 5-3
@app.get("/items",response_model = GetimageItemResponse)
def get_items():

    # connect database
    db_gen = get_db()  # get_db()を呼び出してジェネレータを取得
    conn = next(db_gen)  # ジェネレータを消費してデータベース接続を取得
    cursor = conn.cursor()

    # fetch item table from database
    cursor.execute("SELECT * FROM items")
    items = cursor.fetchall()

    cursor.execute("SELECT * FROM categories")
    category = cursor.fetchall()

    db_gen.close()
    
    # print(items)
    # print(category)
    # [(1, 'jacket', 1, 'ad55d25f2c10c56522147b214aeed7ad13319808d7ce999787ac8c239b24f71d.jpg')]
    # [(1, 'fashion')]

    lst = []
    for i in items:
        if len(i) == 4:
            ele = GetItemimage(name=i[1],category = category[int(i[2])-1][1],image_name= i[3])
        else:
            ele = GetItemimage(name=i[1],category = category[int(i[2])-1][1],image_name= "Nan")
        lst.append(ele)

    resjson = GetimageItemResponse(items = lst)

    return resjson


@app.post("/items",response_model = AddItemResponse)
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):

    # print("kakuninn!!!",image)
    file_name = image.filename
    # Create image path
    image_path = images / file_name


    if not image.filename.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_path.exists():
        logger.info(f"Image not found: {image}")
        image = images / "default.jpg"

    # hashlibの引数はバイナリを読み込むので一度データを読み込む必要がある
    file_hash = hashlib.sha256()
    contents = await image.read()  # ファイルの内容を読み込む
    file_hash.update(contents)
    file_hash_name= file_hash.hexdigest()+".jpg"

    # connect database
    db_gen = get_db()  # get_db()を呼び出してジェネレータを取得
    conn = next(db_gen)  # ジェネレータを消費してデータベース接続を取得
    cursor = conn.cursor()

    # print("kakuninnshite!!!!")
    cursor.execute("SELECT id FROM categories WHERE name = ?",(category,))
    result = cursor.fetchall()

    if result:
        id = result[0][0]
    else:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (category,))
        id = cursor.lastrowid  # 新しく挿入したカテゴリの id を取得


    # insert to database
    cursor.execute("""INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)""", (name, id, file_hash_name))

    # saved change data
    conn.commit()
    db_gen.close()
    
    return AddItemResponse(**{"message": f"item received: {name,category,file_name}"})



    
    
