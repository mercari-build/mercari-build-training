import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
from pydantic import BaseModel
from contextlib import asynccontextmanager
import json
from typing import Optional

# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db_path = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
# items.jsonに新しいアイテムを追加した時のデータを追加するためにjsonファイルのパスを指定
items_file = pathlib.Path(__file__).parent.resolve() / "items.json"

class Item(BaseModel):
    name: str
    category: str
    image_name: str

def get_db():
    if not db_path.exists():
        yield

    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()


# STEP 5-1: set up the database connection
def setup_database():
    db_dir = db_path.parent
    db_dir.mkdir(parents=True, exist_ok=True)
    
    conn = sqlite3.connect(db_path)
    try:
        # UNIQUEによって同じカテゴリ名を重複して挿入できなくなる
        conn.execute("""
            CREATE TABLE IF NOT EXISTS categories (
                id INTEGER PRIMARY KEY,
                name TEXT UNIQUE  
            );    
        """)
        # category_idは、categoriesテーブルのidを参照する外部キー。
        conn.execute("""
            CREATE TABLE IF NOT EXISTS items (
              id INTEGER PRIMARY KEY,
              name TEXT,
              category_id INTEGER,
              image_name TEXT,
              FOREIGN KEY (category_id) REFERENCES categories(id)
            );         
        """)
        conn.commit()
    finally:
        conn.close()


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


# APIサーバが正しく動作しているかの簡単なテストとして利用
@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})


class AddItemResponse(BaseModel):
    message: str


# POST-/item リクエストで呼び出され、アイテム情報の追加を行う
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: Optional[UploadFile] = File(None), 
    db: sqlite3.Connection = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    
    if not category:
        raise HTTPException(status_code=400, detail="category is required")

    # 画像が送られて来なかった時も空文字を登録する
    image_name = ""
    if image is not None:
        # アップロードされた画像ファイルの内容をバイト列として読み込む
        image_bytes = image.file.read()
        if image_bytes:
            # 画像のバイト列データをハッシュ化
            hash_value = hashlib.sha256(image_bytes).hexdigest()
            # 画像の名前をハッシュ化した後の名前に変更
            hashed_image_name = f"{hash_value}.jpg"
            image_path = images / hashed_image_name
            
            # 画像を保存する場所（パス）がバイナリ書き込みモードで開かれ書き込まれる
            with open(image_path, "wb") as f:
                f.write(image_bytes)
                
            image_name = hashed_image_name
            
    # 5-3 カテゴリーidをゲット
    category_id = get_category(category, db)
    
    # 新しいアイテムを作成
    # ※PydanticモデルItemは入力検証用なので、ここでは直接DB挿入用の関数を利用する
    insert_item(name, category_id, image_name, db)
    
    return AddItemResponse(**{"message": f"item received: {name} / category received: {category} / category_id: {category_id} / image received: {image_name}"})


#  GET /search　商品検索エンドポイント
@app.get("/search")
def get_searched_item(keyword: str, db: sqlite3.Connection = Depends(get_db)):
    query = """
        SELECT 
            items.id AS id,
            items.name AS name,
            categories.name AS category,
            items.image_name AS image_name
        FROM items
        JOIN categories
        ON items.category_id = categories.id
        WHERE items.name LIKE ?
    """
    # "% %" で囲むことで、keywordという文字列を含むデータを検索する
    pattern = f"%{keyword}%"
    # sqlクエリを実行した結果返されるオブジェクトがcursorに入る。クエリ実行には,クエリとパラメータ(あれば)を渡す。
    cursor = db.execute(query, (pattern,))
    # 全ての実行結果を取得
    rows = cursor.fetchall()
    # rowsのオブジェクトを、要素が辞書型の配列に変換し、itemsに代入
    items = [dict(row) for row in rows]
    return {"items": items}
    

# GET-/items リクエストで呼び出され、items.jsonファイルの内容(今まで保存された全てのitemの情報)を返す
@app.get("/items")
def get_items(db: sqlite3.Connection = Depends(get_db)):
    cursor = db.execute("""
                        SELECT
                            items.id AS id, 
                            items.name AS name, 
                            categories.name AS category, 
                            items.image_name AS items_name 
                        FROM items
                        JOIN categories
                        ON items.category_id = categories.id
                    """)
    rows = cursor.fetchall()
    items = [dict(row) for row in rows]
    return {"items": items}


@app.get("/items/{item_id}", response_model=Item)
def get_item(item_id: int, db: sqlite3.Connection = Depends(get_db)):
    cursor = db.execute("""
                        SELECT 
                            items.id AS id, 
                            items.name AS name, 
                            categories.name AS category, 
                            items.image_name AS image_name 
                        FROM items 
                        JOIN categories
                        ON items.category_id = categories.id
                        WHERE items.id = ?
                    """, (item_id,))
    # 1行だけ取得するので fetch"one"になる
    row = cursor.fetchone()
    if row is None:
        raise HTTPException(status_code=404, detail="Item not found")
    return dict(row)


# GET-/image/{image_name} リクエストで呼び出され、指定された画像を返す
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # 画像ファイルのパスを生成する
    image_file_path = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_file_path.exists():
        logger.debug(f"Image not found: {image_file_path}")
        image_file_path = images / "default.jpg"

    return FileResponse(image_file_path)



# app.post("/items" ... のハンドラ内で用いられる。items.jsonファイルへ新しい要素の追加を行う。
def insert_item(name: str, category_id: int, image_name: str, db: sqlite3.Connection):
    db.execute(
        "INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
        (name, category_id, image_name)
    )
    db.commit()
    
    
# categories tableからカテゴリー名のidを返す関数 
def get_category(category_name: str, db: sqlite3.Connection) -> int:
    cursor = db.execute("SELECT id FROM categories WHERE name = ?", (category_name,))
    row = cursor.fetchone()
    # カテゴリー名が存在する場合はidを返す
    if row is not None:
        return row["id"]
    # 存在しない場合は新しくcategories tableに追加し、idを発行
    cursor = db.execute("INSERT INTO categories (name) VALUES (?)", (category_name,))
    db.commit()
    return cursor.lastrowid
