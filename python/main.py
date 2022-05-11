import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import random
# import database
import sqlite3

# データベースを開く
conn = sqlite3.connect("../db/item.db", check_same_thread=False)
c = conn.cursor()

#テーブルを作成
# c.execute("CREATE TABLE `items` (`id` int, `name` string,`category` string);")

# 変更を確定
conn.commit()



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


def show_all():
    c.execute("SELECT * FROM items;")
    items = c.fetchall()
    return items
    # for item in items:
    #     print(item)
    
def add_sql(id,name,category):
    c.execute("INSERT INTO items(id,name,category) VALUES(?,?,?)", (id,name,category))
    conn.commit()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def show_item():
    content = show_all()
    # return '{"items":' + content + '}'
    return {"items": {content}}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Receive item: {name}")
    logger.info(f"Receive item: {category}")
    id = random.randint(1,100)
    add_sql(id,name,category)
    return {"message": f"item received: {name}"}

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
