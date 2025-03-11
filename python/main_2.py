# import os
# import logging
# import pathlib
# import hashlib
# from fastapi import FastAPI, Form, HTTPException, Depends, UploadFile, File
# from fastapi.responses import FileResponse
# from fastapi.middleware.cors import CORSMiddleware
# import sqlite3
# from pydantic import BaseModel
# from contextlib import asynccontextmanager
# import json
# from typing import Optional



# # Define the path to the images & sqlite3 database
# images = pathlib.Path(__file__).parent.resolve() / "images"
# db = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"
# # items.jsonに新しいアイテムを追加した時のデータを追加するためにjsonファイルのパスを指定
# items_file = pathlib.Path(__file__).parent.resolve() / "items.json"

# class Item(BaseModel):
#     name: str
#     category: str
#     image_name: str

# def get_db():
#     if not db.exists():
#         yield

#     conn = sqlite3.connect(db)
#     conn.row_factory = sqlite3.Row  # Return rows as dictionaries
#     try:
#         yield conn
#     finally:
#         conn.close()


# # STEP 5-1: set up the database connection
# def setup_database():
#     db_dir = db_path.parent
#     db_dir.mkdir(parents=True, exist_ok=True)
    
#     conn = sqlite3.connect(db_path)
#     try:
#         conn.execute("""
#             CREATE TABLE IF NOT EXISTS items (
#               id INTEGER PRIMARY KEY,
#               name TEXT,
#               category TEXT,
#               image_name TEXT
#             );         
#         """)
#         conn.commit()
#     finally:
#         conn.close()


# @asynccontextmanager
# async def lifespan(app: FastAPI):
#     setup_database()
#     yield


# app = FastAPI(lifespan=lifespan)

# logger = logging.getLogger("uvicorn")
# logger.level = logging.INFO
# images = pathlib.Path(__file__).parent.resolve() / "images"
# origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
# app.add_middleware(
#     CORSMiddleware,
#     allow_origins=origins,
#     allow_credentials=False,
#     allow_methods=["GET", "POST", "PUT", "DELETE"],
#     allow_headers=["*"],
# )


# class HelloResponse(BaseModel):
#     message: str


# # APIサーバが正しく動作しているかの簡単なテストとして利用
# @app.get("/", response_model=HelloResponse)
# def hello():
#     return HelloResponse(**{"message": "Hello, world!"})


# class AddItemResponse(BaseModel):
#     message: str


# # POST-/item リクエストで呼び出され、アイテム情報の追加を行う
# @app.post("/items", response_model=AddItemResponse)
# def add_item(
#     name: str = Form(...),
#     category: str = Form(...),
#     image: Optional[UploadFile] = File(None), 
#     db: sqlite3.Connection = Depends(get_db),
# ):
#     if not name:
#         raise HTTPException(status_code=400, detail="name is required")
    
#     if not category:
#         raise HTTPException(status_code=400, detail="category is required")

#     # 画像が送られて来なかった時も空文字を登録する
#     image_name = ""
#     if image is not None:
#         # アップロードされた画像ファイルの内容をバイト列として読み込む
#         image_bytes = image.file.read()
#         if image_bytes:
#             # 画像のバイト列データをハッシュ化
#             hash_value = hashlib.sha256(image_bytes).hexdigest()
#             # 画像の名前をハッシュ化した後の名前に変更
#             hashed_image_name = f"{hash_value}.jpg"
#             image_path = images / hashed_image_name
            
#             # 画像を保存する場所（パス）がバイナリ書き込みモードで開かれ書き込まれる
#             with open(image_path, "wb") as f:
#                 f.write(image_bytes)
                
#             image_name = hashed_image_name
    
#     # 新しいアイテムを作成
#     new_item = Item(name=name, category=category, image_name=image_name)
#     insert_item(new_item)
    
#     return AddItemResponse(**{"message": f"item received: {name} / category received: {category} / image received: {image_name}"})


# # GET-/image/{image_name} リクエストで呼び出され、指定された画像を返す
# @app.get("/image/{image_name}")
# async def get_image(image_name: str):
#     # 画像ファイルのパスを生成する
#     image_file_path = images / image_name

#     if not image_name.endswith(".jpg"):
#         raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

#     if not image_file_path.exists():
#         logger.debug(f"Image not found: {image_file_path}")
#         image_file_path = images / "default.jpg"

#     return FileResponse(image_file_path)


# # GET-/items リクエストで呼び出され、items.jsonファイルの内容(今まで保存された全てのitemの情報)を返す
# @app.get("/items")
# def get_items():
#     if not items_file.exists():
#         return {"items": []}
    
#     try:
#         with open(items_file, "r", encoding="utf-8") as f:
#             data = json.load(f)
#     except json.JSONDecodeError:
#         data = {"items": []}
        
#     return data

# @app.get("/items/{item_id}", response_model=Item)
# def get_item(item_id: int):
#     if not items_file.exists():
#         raise HTTPException(status_code=404, detail="Items file not found")
    
#     try:
#         with open(items_file, "r", encoding="utf-8") as f:
#             data = json.load(f)
#     except json.JSONDecodeError:
#         raise HTTPException(status_code=500, detail="Failed to open the file")
    
#     items_list = data.get("items", [])
    
#     if not items_list:
#         raise HTTPException(status_code=404, detail="no items found")
    
#     item_index = item_id - 1
#     if item_index < 0 or item_index >= len(items_list):
#         raise HTTPException(status_code=400, detail="index is invalid")
    
#     return items_list[item_index]
    

# # app.post("/items" ... のハンドラ内で用いられる。items.jsonファイルへ新しい要素の追加を行う。
# def insert_item(item: Item):
#     # STEP 4-1: add an implementation to store an item
#     global items_file
    
#     # ファイルが存在しない場合初期状態で作成する
#     if not items_file.exists():
#         with open(items_file, "w", encoding="utf-8") as f:
#             json.dump({"items": []}, f, ensure_ascii=False, indent=2)
            
#     # 既存のファイルがあった場合読み込み
#     try:
#         # すでにデータがある場合
#         with open(items_file, "r", encoding="utf-8") as f:
#             data = json.load(f)
#         # ファイルはあるがデータが空の場合
#     except json.JSONDecodeError:
#         data = {"items": []}
        
#     # 新しいアイテムを追加する
#     data["items"].append({"name": item.name, "category": item.category, "image_name": item.image_name})
    
#     #更新したデータを書き戻す
#     with open(items_file, "w", encoding="utf-8") as f:
#         json.dump(data, f, ensure_ascii="False", indent=2)
        