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
    # db_path = "/Users/tomoka/Build/mercari-build-training/db/mercari2.sqlite3"
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn

# 商品情報のPydanticモデル
class Item(BaseModel):
    name: str
    category: str
    image_name: str

# 商品情報を保存するエンドポイント
# @app.post("/items/")
# async def create_item(item: Item, file: UploadFile = File(...)):
#     conn = get_db_connection()
#     cursor = conn.cursor()
#     file_location = f"images/{file.filename}"
#     with open(file_location, "wb+") as file_object:
#         file_object.write(file.file.read())
#     cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
#                    (item.name, item.category, file_location))
#     conn.commit()
#     conn.close()
#     return {"name": item.name, "category": item.category, "image_name": file.filename}

#test
@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = None):
    conn = get_db_connection()
    cursor = conn.cursor()

    if image:
        file_location = f"images/{image.filename}"
        with open(file_location, "wb+") as file_object:
            file_object.write(image.file.read())
        image_name = image.filename
        cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
                   (name, category, image_name))
    else:
        # ファイルが提供されなかった場合のデフォルトの画像名や処理をここに記述
        # file_location = "aaaaaaaaa"  # デフォルトの画像パスは存在しないため、適切な値に修正する必要がある
        image_name = "No image"  # デフォルトの画像名
        cursor.execute("INSERT INTO items (name, category) VALUES (?, ?)",
                       (name, category))

    # cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
    #                (name, category, image_name))
    conn.commit()
    conn.close()
    return {"name": name, "category": category, "image_name": image_name}
# async def create_item(item: Item, file: UploadFile = None):
#     conn = get_db_connection()
#     cursor = conn.cursor()

#     if file:
#         file_location = f"images/{file.filename}"
#         with open(file_location, "wb+") as file_object:
#             file_object.write(file.file.read())
#         image_name = file.filename
#     else:
#         # ファイルが提供されなかった場合のデフォルトの画像名や処理をここに記述
#         file_location = "aaaaaaaaa"  # デフォルトの画像パス
#         image_name = "default.jpg"  # デフォルトの画像名

#     cursor.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
#                    (item.name, item.category, file_location))
#     conn.commit()
#     conn.close()
#     return {"name": item.name, "category": item.category, "image_name": image_name}


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
    items = conn.execute("SELECT name, category, image_name FROM items WHERE name LIKE ?", ('%' + keyword + '%',)).fetchall()
    conn.close()
    
    # データベースから取得したRowオブジェクトを辞書リストに変換
    items_list = [dict(item) for item in items]
    
    return {"items": items_list}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=9000)




# import os
# import logging
# import pathlib
# import uvicorn
# from fastapi import FastAPI, HTTPException, Form, File, UploadFile
# from fastapi.responses import FileResponse
# from fastapi.middleware.cors import CORSMiddleware
# from typing import Optional, List, Dict
# from databases import Database
# from pydantic import BaseModel
# import hashlib
# import sqlite3


# app = FastAPI()
# logger = logging.getLogger("uvicorn")

# # DATABASE URL
# # DATABASE_URL = "sqlite:////Users/tomoka/Build/mercari-build-training/db/mercari.sqlite3"
# DATABASE_URL = "sqlite:////Users/tomoka/Build/mercari-build-training/db/items.db"
# # データベース接続の初期化
# database = Database(DATABASE_URL)
# con = sqlite3.connect(database)






# # 画像を保存するディレクトリを確認（存在しなければ作成）
# images_dir = "images"
# if not os.path.exists(images_dir):
#     os.makedirs(images_dir)

# # CORSの設定
# origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
# app.add_middleware(
#     CORSMiddleware,
#     allow_origins=origins,
#     allow_credentials=False,
#     allow_methods=["GET", "POST", "PUT", "DELETE"],
#     allow_headers=["*"],
# )






#####

# app = FastAPI()
# logger = logging.getLogger("uvicorn")

# # 画像を保存するディレクトリを確認（存在しなければ作成）
# images_dir = "images"
# if not os.path.exists(images_dir):
#     os.makedirs(images_dir)

# # CORSの設定
# origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
# app.add_middleware(
#     CORSMiddleware,
#     allow_origins=origins,
#     allow_credentials=False,
#     allow_methods=["GET", "POST", "PUT", "DELETE"],
#     allow_headers=["*"],
# )

# # ルートエンドポイント
# @app.get("/")
# def root():
#     return {"message": "Hello, world!"}

# # アイテム追加エンドポイント
# @app.post("/items")
# async def add_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = None):
#     # アイテム情報のログ出力
#     logger.info(f"Received item: {name}, Category: {category}")

#     # 画像ファイルがある場合は処理
#     image_name = ""
#     if image:
#         # 画像の内容を読み取り
#         contents = await image.read()
#         # 画像のハッシュ値を計算してファイル名を生成
#         hash_name = hashlib.sha256(contents).hexdigest()
#         image_name = f"{hash_name}.jpg"
#         image_path = os.path.join(images_dir, image_name)
#         # 画像をファイルに保存
#         with open(image_path, "wb") as file:
#             file.write(contents)
#         logger.info(f"Image saved: {image_name}")

#     # 新しいアイテムIDの決定
#     new_item_id = 1
#     try:
#         with open("items.json", "r") as file:
#             data = json.load(file)
#             if data["items"]:
#                 new_item_id = max(item["item_id"] for item in data["items"]) + 1
#     except FileNotFoundError:
#         raise HTTPException(
#             status_code=404, detail="'items.json' not found"
#         )


#     # アイテムデータの作成
#     item_data = {"item_id": new_item_id, "name": name, "category": category, "image_name": image_name}
#     data["items"].append(item_data)

#     with open("items.json", "w") as file:
#         json.dump(data, file, indent=4)

#     logger.info(f"Item added: {name}, Category: {category}, Image Name: {image_name}, Item ID: {new_item_id}")
#     return {"message": "Item added successfully", "item_id": new_item_id}

# # アイテムゲット
# @app.get("/items")
# async def get_items():
#     try:
#         with open("items.json", "r") as file:
#             data = json.load(file)
#             return data
#     except FileNotFoundError:
#         # return {"detail": "Items not found."}
#         raise HTTPException(
#             status_code=404, detail="'items.json' not found"
#         )

# # 画像取得エンドポイント
# @app.get("/image/{image_name}")
# async def get_image(image_name: str):
#     image_path = pathlib.Path(images_dir) / image_name
#     if not image_name.endswith(".jpg"):
#         raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
#     if not image_path.exists():
#         logger.debug(f"Image not found: {image_path}")
#         image_path = pathlib.Path(images_dir) / "default.jpg"
#     return FileResponse(image_path)

# # 特定のアイテムを取得するエンドポイント
# @app.get("/items/{item_id}")
# async def get_item(item_id: int):
#     try:
#         with open("items.json", "r") as file:
#             data = json.load(file)
#             # item_idに一致する商品を検索
#             item = next((item for item in data["items"] if item["item_id"] == item_id), None)
#             if item:
#                 return item
#             else:
#                 raise HTTPException(status_code=404, detail="Item not found")
#     except FileNotFoundError:
#         raise HTTPException(status_code=404, detail="Items file not found")
