import os
import logging
import pathlib
import hashlib
import json
import sqlite3
from fastapi import Path
from fastapi import FastAPI, Form, UploadFile, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG  # ログレベルをDEBUGに変更
# images ファイルへのパス
images = pathlib.Path(__file__).parent.resolve() / "images"
images.mkdir(parents=True, exist_ok=True)  # imagesディレクトリを作成する
# items.json ファイルのパス
items_file_path = pathlib.Path(__file__).parent.resolve() / "items.json"
# mercari.sqlite3 のパス
sqlite3_file = "mercari.sqlite3"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)


# items.json ファイルに保存
def save_items_to_json(items):
    with open(items_file_path, "w") as f:
        json.dump(items, f, indent=4)

def load_items_from_json():
    if items_file_path.exists():
        with open(items_file_path, "r") as f:
            return json.load(f)
    return {"items": []}

def save_image(file, filename):
    with open(images / filename, "wb") as image:
        image.write(file)

# dbに保存
def save_items_to_db(items):
    with open(items_file_path, "w") as f:
        json.dump(items, f, indent=4)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def get_items():
    # DBに接続する
    conn = sqlite3.connect(sqlite3_file)
    # SQLiteを操作するためのカーソルを作成
    cursor = conn.cursor()
    # DBのクエリを実行
    cursor.execute('SELECT * FROM items')   
    # 実行したクエリの中身を全て取得
    items = cursor.fetchall()
    # DBとの接続を閉じる
    cursor.close()
    conn.close()
    
    {"items": items}


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}, category: {category}, image: {image}")
    
    # 画像ファイルのハッシュを計算
    file_content = image.file.read()
    hash_value = hashlib.sha256(file_content).hexdigest()
    
    # 画像ファイルを保存
    image_filename = f"{hash_value}.jpg"
    save_image(file_content, image_filename)
    
    # 新しい商品情報を作成
    new_item = {"name": name, "category": category, "image": image_filename}
    
    # DBに接続する
    conn = sqlite3.connect(sqlite3_file)
    # SQLiteを操作するためのカーソルを作成
    cursor = conn.cursor()
    # データの挿入
    cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)", (name, category, image_filename))
    # 挿入した結果を保存
    conn.commit()
    # DBとの接続を閉じる
    cursor.close()
    conn.close()
    
    
    # 既存の商品リストを取得
    items_data = load_items_from_json()
    existing_items = items_data.get("items", [])
    
    # 新しい商品を追加
    existing_items.append(new_item)
    items_data["items"] = existing_items
    
    # 商品情報を JSON ファイルに保存
    save_items_to_json(items_data)
    
    return {"message": f"item received: {name}, category: {category}, image: {image_filename}"}

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

@app.get("/items/{item_id}")
def get_item(item_id: int = Path(..., title="The ID of the item to get")):
    conn = sqlite3.connect(sqlite3_file)
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM items WHERE id = ?", (item_id,))
    item = cursor.fetchone()
    cursor.close()
    conn.close()
    
    # 指定されたitem_idに対応する商品を取得
    if item:
        return item
    else:
        raise HTTPException(status_code=404, detail="Item not found")
    
    
@app.get("/search")
def search_items(keyword: str):
    conn = sqlite3.connect(sqlite3_file)
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM items WHERE name LIKE ?", ('%' + keyword + '%',))
    items = cursor.fetchall()
    cursor.close()
    conn.close()
    
    return {"items": items}