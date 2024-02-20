import os
import logging
import pathlib
import json
import hashlib
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import FileResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()
logger = logging.getLogger("uvicorn")
# step3-6 Loggerについて調べる
logger.level = logging.DEBUG
images = pathlib.Path(__file__).parent.resolve() / "images"
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
    return JSONResponse(json.dumps(items_data))

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    # 画像のファイル名の取得
    image_filename = await store_image(image)

    new_item = {"name": name, "category": category, "image_name": image_filename}
    
    # 新しい商品をJSONに追加
    add_item_to_json(new_item)

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