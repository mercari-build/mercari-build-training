import os
import logging
import uvicorn
from fastapi import FastAPI, HTTPException, File, UploadFile, Form
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import List, Optional
import sqlite3
from pydantic import BaseModel

app = FastAPI()
logger = logging.getLogger("uvicorn")

# ルートエンドポイント
@app.get("/")
def root():
    return {"message": "Hello, world!"}

# データベース接続関数
def get_db_connection():
    db_path = "/Users/tomoka/Build/mercari-build-training/db/mercari.sqlite3"
    #まず、ここでは.sqlite3に接続する。
    #先に、ターミナルで、.sqlite3に紐づけられたitems.sqlとcategories.sqlがあることを前提にする。
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn

# 商品情報のPydanticモデル
class Item(BaseModel):
    name: str
    category: str
    image_name: str

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = None):
    conn = get_db_connection()
    cursor = conn.cursor()

    # category idの選択
    cursor.execute("SELECT id FROM categories WHERE name = ?", (name,))
    category = cursor.fetchone()
    # return {"category": category}

    if not category:
        cursor.execute("INSERT INTO categories (name) VALUES (?)", (name,))
        conn.commit()
        category_id = cursor.lastrowid
    else:
        category_id = category[0]

    if image:
        file_location = f"images/{image.filename}"
        with open(file_location, "wb+") as file_object:
            file_object.write(image.file.read())
        image_name = image.filename
        # cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
        #            (name, category, image_name))
        # category id の変更点
        cursor.execute("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
                   (name, category_id, image_name))
    else:
        image_name = "No image"  # デフォルトの画像名
        cursor.execute("INSERT INTO items (name, category_id) VALUES (?, ?)",
                       (name, category_id))

    conn.commit()
    conn.close()
    # return {"name": name, "category": category, "image_name": image_name}
    # category id の変更点
    return {"name": name, "category_id": category_id, "image_name": image_name}

# 保存された商品情報を取得するエンドポイント
@app.get("/items/", response_model=List[Item])
def read_items():
    conn = get_db_connection()
    items = conn.execute("SELECT * FROM items").fetchall()
    conn.close()
    return [dict(item) for item in items]

# 画像を提供するエンドポイント
@app.get("/images/{filename}", response_class=FileResponse)
def read_image(filename: str):
    return FileResponse(path=f"images/{filename}")

# CORSの設定
origins = ["*"]  # ここで適切なオリジンに設定

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/search")
def search_item(keyword: str):
    conn = get_db_connection()
    # itemsから引っ張ってくる場合
    # items = conn.execute("SELECT name, category_id, image_name FROM items WHERE name LIKE ?", ('%' + keyword + '%',)).fetchall()
    # conn.close()
    # # データベースから取得したRowオブジェクトを辞書リストに変換
    # items_list = [dict(item) for item in items]
    # return {"items": items_list}

    # categoriesから引っ張ってくる場合
    # LIKEを使うと似ている単語が収集される
    # categories = conn.execute("SELECT id FROM categories WHERE name LIKE ?", ('%' + keyword + '%',)).fetchall()
    categories = conn.execute("SELECT id, name FROM categories WHERE name = ?", (keyword,)).fetchall()

    conn.close()
    categories_list = [dict(category) for category in categories]
    
    return {"categories": categories_list}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=9000)
