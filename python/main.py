import os
import logging
#　import json
import pathlib
import hashlib  # SHA-256　ハッシュ化用
import sqlite3
from fastapi import FastAPI, Query, Form, File, UploadFile, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from contextlib import asynccontextmanager


# Define the path to the images & sqlite3 database
images = pathlib.Path(__file__).parent.resolve() / "images"
db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
# items_json_path = pathlib.Path(__file__).parent.resolve() / "items.json"  # items.json のパス追記

def get_db():
    #　if not db.exists():
        #　yield
    conn = sqlite3.connect("db/mercari.sqlite3")
    conn.row_factory = sqlite3.Row  # Return rows as dictionaries　# 結果を辞書形式で扱えるようにする
    return conn


# STEP 5-1: set up the database connection
def setup_database():
    pass

@asynccontextmanager
async def lifespan(app: FastAPI):
    setup_database()
    yield


app = FastAPI(lifespan=lifespan)

logger = logging.getLogger("uvicorn")
logger.level = logging.INFO

origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# モデル定義

class HelloResponse(BaseModel):
    message: str

class AddItemResponse(BaseModel):
    message: str

class Item(BaseModel):
    name: str
    category: str  # カテゴリー追加
    image_name: str #　画像のファイル名追加

@app.get("/", response_model=HelloResponse)
def hello():
    return HelloResponse(**{"message": "Hello, world!"})

# 画像をSHA-256でハッシュ化し、保存する関数
def save_image(image: UploadFile) -> str:
    # 画像データ読み込み
    image_data = image.file.read()
    # ハッシュ値作成
    hash_value = hashlib.sha256(image_data).hexdigest()
    # ハッシュ値を使ってファイル名を作成(重複を防ぐため)
    file_name = f"{hash_value}.jpg"
    # 画像を保存
    image_path = images / file_name
    with open(image_path,"wb") as file:
        file.write(image_data)
    return file_name #画像のハッシュ名を返す


# items.jsonのデータ読み込み
#def load_items():
    #if items_json_path.exists():
        #with open(items_json_path, "r", encoding="utf-8") as file:
            #data = json.load(file)
            #return data.get("items", [])
    #return []

#item.jsonに商品を追加して保存する関数
# def save_item(item: Item):
    # items = load_items()
    # items.append({"name": item.name, "category": item.category, "image_name": item.image_name}) #商品を追加する

    #with open(items_json_path, "w", encoding="utf-8") as file:
        #json.dump({"items": items}, file, indent=2, ensure_ascii=False)  # JSONを保存


#登録された商品データを取得
@app.get("/items") 
def get_items(db: sqlite3.Connection = Depends(get_db)):
    cursor = db.cursor()
    cursor.execute("""
        SELECT items.id, items.name, categories.name as category, items.image_name
        FROM items
        JOIN categories ON items.category_id = categories.id
    """)
    items = cursor.fetchall()
    db.close()
    return{"items":[dict(item) for item in items]}
    #return {"items": load_items()}

#@app.get("/items/{item_id}")
#def get_item(item_id: int):
    #items = load_items()

    #if item_id < 0 or item_id >= len(items):
        #raise HTTPException(status_code=404, detail=f"Item ID {item_id} is not found")

    #return items[item_id]  # 指定されたIDの商品を返す

# add_item is a handler to add a new item for POST /items .
@app.post("/items", response_model=AddItemResponse)
def add_item(
    name: str = Form(...),
    category: str = Form(...), #カテゴリーを受け取る
    image: UploadFile = File(...), # 画像を受け取る
    db: sqlite3.Connection = Depends(get_db),
):
    if not name or not category:
        raise HTTPException(status_code=400, detail="name and category is required")
    
    image_name = save_image(image) #画像を保存、ハッシュ名を取得
    insert_item(db, Item(name=name, category=category, image_name=image_name))

    return AddItemResponse(**{"message": f"item received: {name}, {category}, {image_name}"})


# get_image is a handler to return an image for GET /images/{filename} .
@app.get("/image/{image_name}")
async def get_image(image_name: str):
    # Create image path
    image = images / image_name

    if not image.suffix.lower() in [".jpg", ".jpeg", ".png"]:  # jpg, jpeg, png のみ許可
        raise HTTPException(status_code=400, detail="Invalid image format")

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=404, detail=f"Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)



def insert_item(db: sqlite3.Connection, new_item: Item):
    # STEP 4-1: add an implementation to store an item　
    cursor = db.cursor()
    cursor.execute(
        "INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
        (new_item.name, new_item.category, new_item.image_name)
    )
    db.commit()

# キーワード検索をするエンドポイント
@app.get("/search")
def search_items(
    keyword: str= Query(..., description='検索キーワード'), 
    db: sqlite3.Connection = Depends(get_db)
):
    cursor = db.cursor()
    cursor.execute("SELECT * FROM items WHERE name LIKE ?", ('%' + keyword + '%',))
    items = cursor.fetchall()
    db.close()

    if not items:
        raise HTTPException(status_code=404,detail="No items found")
    
    return {"items": [dict(item) for item in items]}
