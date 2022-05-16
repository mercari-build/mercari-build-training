import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import json
import sqlite3

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
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):

    #初期化
    # items_list = {"items" : []}
    # if os.path.isfile('items.json') == True:
    #     with open("items.json", "r") as items_json_file:
    #         item_list = json.load(items_json_file)

    # with open("items.json", "w") as items_json_file:
    #     new_data = {"name" : name, "category" : category}
    #     items_list["items"].append(new_data)
    #     json.dump(items_list, items_json_file)

    dbname = '../db/mercari.sqlite3'
    # DBを作成する（既に作成されていたらこのDBに接続する）
    conn = sqlite3.connect(dbname)
    # SQLiteを操作するためのカーソルを作成
    cur = conn.cursor()
    #データ登録
    cur.execute('insert into items (name, category) values (?,?)', (name,category))
    #コミットしないと登録が反映されない
    conn.commit()
    # DBとの接続を閉じる(必須)
    conn.close()

    #3-3
    with open('items.json','r',encoding='utf-8') as f:
        if len(f.readline())==0:
            items_json={"items" : []}
        #{"items": [{"name": "jacket", "category": "fashion"}, {"name": "jacket", "category": "fashion"},...]}
        #読み込み
        else:
            f.seek(0)
            items_json=json.load(f)

    with open('items.json','w',encoding='utf-8') as f:
        item_new={"name": name, "category": category}
        items_json["items"].append(item_new)
        json.dump(items_json,f)

    logger.info(f"Receive item: {name}")
    return {"message": f"item received: {name}"}

@app.get("/items")
def get_item():

    dbname = '../db/mercari.sqlite3'
    # DBを作成する（既に作成されていたらこのDBに接続する）
    conn = sqlite3.connect(dbname)
    # SQLiteを操作するためのカーソルを作成
    cur = conn.cursor()
    #データ登録
    cur.execute('select id,name,category from items')
    #クエリの現在行から残り全行を取得し、次の行へ移動。
    items = cur.fetchall()
    #コミットしないと登録が反映されない
    conn.commit()
    # DBとの接続を閉じる(必須)
    conn.close()


    items_list = {"items" : []}
    if os.path.isfile("items.json") == True:
        with open("items.json") as items_json_file:
            items_list = json.load(items_json_file)

    return items_list

@app.get("/search")
def search_item(keyword: str):
    dbname = '../db/mercari.sqlite3'
    # DBを作成する（既に作成されていたらこのDBに接続する）
    conn = sqlite3.connect(dbname)
    cur = conn.cursor()
    
    cur.execute(f"SELECT name, category FROM items WHERE name LIKE '%{keyword}%'")
    items = cur.fetchall()
    
    result_dict = {}
    result_dict['items'] = [{'name': name, 'category': category} for name, category in items]

    conn.commit()
    conn.close()
    return result_dict

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
