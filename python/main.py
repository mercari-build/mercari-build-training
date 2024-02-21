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


# 新しい商品をitem.jsonに保存
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


# 新しい商品をmercari_sqlite3に保存
# itemsテーブルに既に存在する場合は保存しない
def save_items_to_sqlite3(name, category, image_name):
    con = sqlite3.connect(sqlite3_file)
    cur = con.cursor()
    
    cur.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_check = cur.fetchone()
    if not category_check:
        logger.debug(f"The item's category doesn't exist in the categories yet. ")
        try:
            cur.execute("INSERT INTO categories(name) VALUES (?)", (category,) )
            con.commit()
            category_check = cur.lastrowid

        except sqlite3.Error as e:
            con.rollback()
            logger.error(f"エラーが発生したためロールバック: {e}")
    else:
        category_id = category_check[0]

    cur.execute("SELECT id FROM items WHERE name = ?", (name,))
    exist_check = cur.fetchone()
    if not exist_check:
        logger.debug(f"The item doesn't exist yet. ")
        try:
            cur.execute("INSERT INTO items(name, category_id, image_name) VALUES (?, ?, ?)", (name, category_id, image_name) )
            con.commit()
            
        except sqlite3.Error as e:
            con.rollback()
            logger.error(f"エラーが発生したためロールバック: {e}")
    con.close()


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

# category_id :int から対応する category :string を返す
def get_category_name(category_id):
    con = sqlite3.connect(sqlite3_file)
    cur = con.cursor()
    cur.execute("SELECT name FROM categories WHERE id = ?", (category_id,))
    category = cur.fetchone()
    cur.close()
    con.close()
    return category[0]


# 商品一覧を取得
@app.get("/items")
def get_items():
    #items = get_items_from_json()    # items.jsonから
    items = get_items_from_sqlite()    # mercari.sqlite3から
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
    items_list = {"items":[]}
    for i in range(len(items)):
        category_id = items[i][2]
        category = get_category_name(category_id)
        item = {"name":items[i][1], "category":category, "image_name":items[i][3]}
        items_list["items"].append(item)
    cur.close()
    con.close()
    return items_list



# items.jsonに登録された商品のn番目の詳細を取得
@app.get("/items/{item_number}")
def get_nth_item(item_number: int):
    #item = get_Nth_item_from_json(item_number)     # items.jsonから
    item = get_Nth_item_from_sqlite(item_number)     # mercari.sqlite3から
    return item


# items.jsonに登録された商品のn番目の詳細を取得
def get_Nth_item_from_json(item_number: int):
    with open(items_file) as f:
        items = json.load(f)
    if item_number > len(items["items"]):
        return {"message": "item not found"}
    return items["items"][item_number-1]
    
# mercari.sqlite3に登録された商品のn番目の詳細を取得
def get_Nth_item_from_sqlite(item_number: int):
    con = sqlite3.connect(sqlite3_file)
    cur = con.cursor()
    cur.execute('SELECT * FROM items')
    items = cur.fetchall()

    if item_number > len(items):
        return {"message": "item not found"}
    
    category_id = items[item_number-1][2]
    cur.execute("SELECT name FROM categories WHERE id = ?", (category_id,))
    category = cur.fetchone()
    item = {"name":items[item_number-1][1], "category":category[0], "image_name":items[item_number-1][3]}

    return item
        

# imageを返す
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
    the_name_item_list={"items":[]}
    if not items:
        return logger.debug(f"The name's items not found: {keyword}")
    items = cur.fetchall()
    for i in range(len(items)):
        category_id = items[i][2]
        category = get_category_name(category_id)
        the_name_item = {"name":items[i][1], "category":category, "image_name":items[i][3]}
        the_name_item_list["items"].append(the_name_item)

    cur.close()
    con.close()
    return the_name_item_list

