import os
import logging
import pathlib
import json
import hashlib
import sqlite3
from fastapi import FastAPI, Form, HTTPException, File, UploadFile, Query
from fastapi.responses import FileResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
# step3-6 Loggerについて調べる
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.parent.resolve() / "db"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

file_path = "items.json"

@app.get("/")
def root():
    return {"message": "Hello, world!"}

# step3-3 商品一覧を取得する
@app.get("/items")
def get_items():
    # ファイル読み込み 
    with open(file_path, "r") as file:
        items_data = json.load(file)

    logger.info(f"Receive items: {items_data}")
    # return JSONResponse(json.dumps(items_data))
        
    # データベースの商品一覧を取得
    # return select_items()

    # tableを削除してしまうので、1回しか実行しない！！！
    # split_tables()

    # 分割されたデータベースから商品の一覧を取得
    return select_join_items()

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    # 画像のファイル名の取得
    image_filename = await store_image(image)

    # 新しい商品をJSONに追加
    new_item = {"name": name, "category": category, "image_name": image_filename}
    add_item_to_json(new_item)

    # 新しい商品をデータベースに追加
    insert_items(new_item)
    
    logger.info(f"Receive item: {name}, {category}, {image_filename}")
    return {"message": f"item received: {name}, {category}, {image_filename}"}

# step3-5 商品の詳細を返す
@app.get("/items/{item_id}")
def get_item_id(item_id: int):
    with open(file_path, "r") as file:
        items_data = json.load(file)

    if item_id < 1 or item_id > len(items_data['items'])+1:
        raise HTTPException(status_code=400, detail="item id is not a valid number")
    
    logger.info(f"Receive item: {items_data['items'][item_id-1]}")
    return items_data["items"][item_id-1]

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

@app.get("/search")
def get_search_items(keyword: str = Query(...)):
    return search_items(keyword)

# step3-2 新しい商品を登録する
def add_item_to_json(new_item):
    # ファイルの読み込み 
    with open(file_path, "r") as file:
        items_data = json.load(file)

    # itemの追加 itemsキーが存在しなければ作成
    items_list = items_data.get("items", [])
    items_list.append(new_item)

    # ファイルの書き込み
    with open(file_path, "w") as file:
        json.dump({"items": items_list}, file)

# step3-4 画像を登録する
async def store_image(image):
    image_bytes = await image.read()
    image_hash = hashlib.sha256(image_bytes).hexdigest()
    image_filename = f"{image_hash}.jpg"

    # バイナリファイルへ書き込み
    with open(images / image_filename, "wb") as image_file:
        image_file.write(image_bytes)

    logger.info(f"Receive name: {image_filename}")
    return image_filename

# step4-1 SQLiteに情報を移項する
# https://qiita.com/saira/items/e08c8849cea6c3b5eb0c
def insert_items(new_item):
    conn = sqlite3.connect(db/"items.db")
    cur = conn.cursor()

    # 存在しない場合は、新規作成
    cur.execute('''CREATE TABLE IF NOT EXISTS items
                (id INTEGER PRIMARY KEY,
                name TEXT,
                category TEXT,
                image_name TEXT)''')
    
    # データの挿入
    data = [new_item["name"], new_item["category"], new_item["image_name"]]
    sql = 'INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)'
    cur.execute(sql,data)

    conn.commit()
    conn.close()

# step4-1 SQLiteに情報を移項する
def select_items():
    conn = sqlite3.connect(db/"items.db")
    cur = conn.cursor()

    cur.execute('SELECT * FROM items')
    item_list = cur.fetchall()

    conn.close()

    return item_list

# step4-2 商品を検索する
# 参考：https://www.sejuku.net/blog/73619
def search_items(keyword):
    conn = sqlite3.connect(db/"items.db")
    cur = conn.cursor()

    cur.execute("SELECT * FROM items WHERE name LIKE ?", ('%' + keyword + '%',))
    item_list = cur.fetchall()

    conn.close()

    return item_list

# step4-3 カテゴリの情報を別のテーブルに移す
# items table -> items table + categories table
# 元のitems tableを削除してしまうので、1回しか実行しない！！！
def split_tables():
    conn = sqlite3.connect(db/"items.db")
    cur = conn.cursor()

    # categories table の作成
    cur.execute('''CREATE TABLE IF NOT EXISTS categories (
                id INTEGER PRIMARY KEY,
                name TEXT)''')
    
    # items table からユニークな category を抽出し categories table に挿入
    cur.execute('''INSERT OR IGNORE INTO categories (name) 
                SELECT DISTINCT category FROM items''')
    
    # new_items table の作成
    cur.execute('''CREATE TABLE IF NOT EXISTS new_items (
                id INTEGER PRIMARY KEY,
                name TEXT,
                category_id INTEGER,
                image_name TEXT,
                FOREIGN KEY (category_id) REFERENCES categories (id))''')
    
    # items table からデータを取得し、new_items table に挿入
    cur.execute('''INSERT INTO new_items (id, name, category_id, image_name) 
                SELECT id, name, (SELECT id FROM categories WHERE categories.name = items.category), image_name FROM items''')

    # items table を削除
    cur.execute('''DROP TABLE items''')

    # new_items table の名前を items に変更
    cur.execute('''ALTER TABLE new_items RENAME TO items''')

    # debug用
    cur.execute('SELECT * from categories')
    categories_list = cur.fetchall()
    logger.info(categories_list)

    # debug用
    cur.execute('SELECT * from items')
    items_list = cur.fetchall()
    logger.info(items_list)
    
    conn.commit()
    conn.close()

# step4-3 カテゴリの情報を別のテーブルに移す
# 参考：https://www.javadrive.jp/sqlite/join/index1.html
def select_join_items():
    conn = sqlite3.connect(db/"items.db")
    cur = conn.cursor()

    # SELECT (取得するカラム) FROM テーブル名1 INNER JOIN テーブル名2 ON (結合条件);
    cur.execute("SELECT items.id, items.name, categories.name AS category, items.image_name FROM items INNER JOIN categories ON items.category_id = categories.id")
    items_list = cur.fetchall()

    conn.close()

    return items_list
