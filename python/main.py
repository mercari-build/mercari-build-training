import os
import logging
import pathlib
import uvicorn
from fastapi import FastAPI, HTTPException, Form, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import Optional, List, Dict
from databases import Database
from pydantic import BaseModel
import hashlib

# DATABASE URL
# DATABASE_URL = "sqlite:////Users/tomoka/Build/mercari-build-training/db/mercari.sqlite3"
DATABASE_URL = "sqlite:////Users/tomoka/Build/mercari-build-training/db/items.db"
# データベース接続の初期化
database = Database(DATABASE_URL)


app = FastAPI()
logger = logging.getLogger("uvicorn")



# 画像を保存するディレクトリを確認（存在しなければ作成）
images_dir = "images"
if not os.path.exists(images_dir):
    os.makedirs(images_dir)

# CORSの設定
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# ルートエンドポイント
@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.on_event("startup")
async def startup_event():
    await database.connect()

@app.on_event("shutdown")
async def shutdown_event():
    await database.disconnect()

# # 以降のエンドポイントは、SQLiteデータベースとのインタラクションに更新する必要があります。
# # 以下は、jsonファイルを使用した以前の実装の例です。

# # アイテムゲット（この部分をSQLiteデータベースからデータを取得するように更新する必要があります）
# @app.get("/items")
# async def get_items():
#     query = "SELECT * FROM items"
#     items = await database.fetch_all(query=query)
#     return {"items": items}

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

# # 特定のアイテムを取得するエンドポイント（この部分もデータベースからデータを取得するように更新する必要があります）
# @app.get("/items/{item_id}")
# async def get_item(item_id: int):
#     query = "SELECT * FROM items WHERE id = :item_id"
#     item = await database.fetch_one(query=query, values={"item_id": item_id})
#     if item:
#         return item
#     else:
#         raise HTTPException(status_code=404, detail="Item not found")

# # キーワード検索エンドポイント（データベースを使用した実装例）
# @app.get("/search")
# async def search_items(keyword: Optional[str] = None) -> Dict[str, List[Dict]]:
#     if not keyword:
#         return {"items": []}
#     query = "SELECT * FROM items WHERE name LIKE :keyword"
#     items = await database.fetch_all(query=query, values={"keyword": f"%{keyword}%"})
#     return {"items": items}

# @app.post("/items")
# async def add_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = File(None)):
#     image_name = ""
#     if image:
#         contents = await image.read()
#         hash_name = hashlib.sha256(contents).hexdigest()  # hashlibが定義されているため、この行は正しく動作します。
#         image_name = f"{hash_name}.jpg"
#         image_path = os.path.join(images_dir, image_name)
#         with open(image_path, "wb") as file:
#             file.write(contents)

    
#     # データベースへのアイテム追加
#     query = "INSERT INTO items (name, category, image_name) VALUES (:name, :category, :image_name)"
#     values = {"name": name, "category": category, "image_name": image_name}
#     last_record_id = await database.execute(query, values=values)
    
#     return {"id": last_record_id, "name": name, "category": category, "image_name": image_name}

