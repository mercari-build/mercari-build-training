import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import hashlib
import sqlite3
from sqlite3 import Connection
from typing import List


app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO

# 画像を保存するフォルダのパスを設定
images_dir = pathlib.Path(__file__).parent.resolve() / "images"

# 商品情報を保存するJSONファイルのパスを設定
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"

# CORS設定
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

DATABASE_URL = "../db/mercari.sqlite3"

def get_db_connection() -> Connection:
    conn = sqlite3.connect(DATABASE_URL)
    conn.row_factory = sqlite3.Row  
    return conn


@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    # 画像ファイルをSHA256でハッシュ化して保存
    content = await image.read()
    image_hash = hashlib.sha256(content).hexdigest()
    image_filename = f"{image_hash}.jpg"
    image_path = images_dir / image_filename
    with open(image_path, "wb") as f:
        f.write(content)

    # 新しい商品情報をデータベースに挿入
    conn = get_db_connection()
    cur = conn.cursor()
    cur.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
                (name, category, image_filename))
    conn.commit()
    item_id = cur.lastrowid
    conn.close()

    logger.info(f"Item added: {name}")
    return {"id": item_id, "name": name, "category": category, "image_name": image_filename}


@app.get("/items")
def get_items():
    conn = get_db_connection()
    items = conn.execute("SELECT * FROM items").fetchall()
    conn.close()
    return {"items": [dict(item) for item in items]}


@app.get("/items/{item_id}")
def get_item(item_id: int):
    conn = get_db_connection()
    item = conn.execute("SELECT * FROM items WHERE id = ?", (item_id,)).fetchone()
    conn.close()
    if item:
        return dict(item)
    else:
        raise HTTPException(status_code=404, detail="Item not found")

@app.get("/search")
def search_items(keyword: str):
    conn = get_db_connection()
    # キーワードを含む商品を検索（大文字小文字を区別しないためにLOWER関数を使用）
    items = conn.execute("SELECT * FROM items WHERE LOWER(name) LIKE LOWER(?)", (f"%{keyword}%",)).fetchall()
    conn.close()
    return {"items": [dict(item) for item in items]}