import os
import json
import logging
import pathlib
import hashlib
import sqlite3
from fastapi import FastAPI, Form, HTTPException,File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG #デバッグを可能にする

images = pathlib.Path(__file__).parent.resolve() / "images"
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"
sqlite3_file = pathlib.Path(__file__).parent.parent.resolve() / "db" / "mercari.sqlite3"


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


# 商品を登録する
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Receive item: {name}")
    logger.info(f"Receive item: {category}")
    logger.info(f"Receive item: {image}")

    image_name = image.filename
    hashed_image_name = get_hash_by_sha256(image_name) #hash.jpg 作成

    #save_items_to_file(name,category,hashed_image_name) #items.jsonに商品を保存
    save_items_to_sqlite3(name, category, hashed_image_name) #mercari.sqlite3に商品を保存

    save_image_file(image, hashed_image_name) #imagesに画像を保存

    return {"message": f"item received: {name}"}


# 新しい商品をmercari_sqlite3ファイルに保存
# {"items": [{"id": int, "name": string, "category": string, "image_name": string}, ...]}
def save_items_to_sqlite3(name, category, image_name):
    con = sqlite3.connect(sqlite3_file)
    cur = con.cursor()
    new_item = {"name": name, "category": category, "image_name": image_name}
    try:
        cur.execute("INSERT INTO items(name, category, image_name) VALUES (?, ?, ?)", (name, category, image_name))
        con.commit()
        cur.execute("DELETE FROM items WHERE id NOT IN (SELECT min_id from (SELECT MIN(id) min_id FROM items GROUP BY name,category,image_name) tmp)")
    except sqlite3.Error as e:
        con.rollback()
        logger.error(f"エラーが発生したためロールバック: {e}")
    con.close()


# 新しい商品をitem.jsonファイルに保存
# {"items": [{"name": "jacket", "category": "fashion", "image_name": "xxxxx.jpg"}, ...]}
def save_items_to_file(name, category, image_name):
    new_item = {"name": name, "category": category, "image_name": image_name}
    if os.path.exists(items_file):
        with open(items_file,'r') as f:
            now_data = json.load(f)
        if new_item in now_data["items"]:
            return 
        else:
            now_data["items"].append(new_item)
        with open(items_file, 'w') as f:
            json.dump(now_data, f, indent=2)
    else:
        first_item = {"items": [new_item]}
        with open(items_file, 'w') as f:
            json.dump(first_item, f, indent=2)

# 画像をimagesに保存 
def save_image_file(image, jpg_hashed_image_name):
    imagefile = image.file.read()
    image = images / jpg_hashed_image_name
    with open(image, 'wb') as f:
        f.write(imagefile)
    return


# sha256でハッシュ化 jpg型で返す
def get_hash_by_sha256(image):
    hs = hashlib.sha256(image.encode()).hexdigest()
    return hs+".jpg"


# 商品一覧を取得
@app.get("/items")
def get_items():
    #items = get_items_from_json()
    items = get_items_from_sqlite()
    return items

# items.jsonから商品一覧を取得
def get_items_from_json():
    with open(items_file) as f:
        items = json.load(f)
    return items

# mercari.sqlite3から商品一覧を取得
def get_items_from_sqlite():
    con = sqlite3.connect(sqlite3_file)
    cur = con.cursor()
    cur.execute('SELECT * FROM items')
    items = cur.fetchall()
    cur.close()
    con.close()
    return items



# items.jsonファイルに登録された商品のn番目の詳細を取得
@app.get("/items/{item_number}")
def get_items_item(item_number: int):
    with open(items_file) as f:
        items = json.load(f)
    if item_number > len(items["items"]):
        return {"message": "item not found"}
    return items["items"][item_number-1]
    


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


# 商品をmercari.sqlite3から検索する
@app.get("/search")
def search_items(keyword: str):
    logger.debug(f"Search items name : {keyword}")
    con = sqlite3.connect(sqlite3_file)
    cur = con.cursor()
    items = cur.execute("SELECT * FROM items WHERE name LIKE ?", (f"%{keyword}%",))
    if not items:
        logger.debug(f"Items not found: {keyword}")
    items = cur.fetchall()

    cur.close()
    con.close()
    return items

