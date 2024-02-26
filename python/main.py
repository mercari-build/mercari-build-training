import os
import logging
import pathlib
import json
from fastapi import FastAPI, Form, HTTPException, File, UploadFile
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import hashlib

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

@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    # 画像ファイルをSHA256でハッシュ化して保存
    content = await image.read()
    image_hash = hashlib.sha256(content).hexdigest()
    image_filename = f"{image_hash}.jpg"
    image_path = images_dir / image_filename
    with open(image_path, "wb") as f:
        f.write(content)

    new_item = {"name": name, "category": category, "image_name": image_filename}
    if items_file.exists():
        with open(items_file, "r+", encoding="utf-8") as file:
            data = json.load(file)
            new_item["id"] = len(data["items"]) + 1  
            data['items'].append(new_item)
            file.seek(0)
            file.truncate()
            json.dump(data, file, indent=4)
    else:
        with open(items_file, "w", encoding="utf-8") as file:
            json.dump({"items": [new_item]}, file, indent=4)
    logger.info(f"Item added: {name}")

    if items_file.exists():
        with open(items_file, "r", encoding="utf-8") as file:
            data = json.load(file)
            # ここで全アイテムの情報を含むレスポンスを返します
            return {"items": data["items"]}
    else:
        # アイテムファイルが存在しない場合（通常はあり得ないが、念のため）
        return {"items": [new_item]}

@app.get("/items")
def get_items():
    if items_file.exists():
        with open(items_file, "r", encoding="utf-8") as file:
            data = json.load(file)
            return JSONResponse(content=data)
    else:
        return JSONResponse(content={"items": []})

@app.get("/items/{item_id}")
def get_item(item_id: int):
    if items_file.exists():
        with open(items_file, "r", encoding="utf-8") as file:
            data = json.load(file)
            items = data["items"]
            item = next((item for item in items if item["id"] == item_id), None)
            if item:
                return item
            else:
                raise HTTPException(status_code=404, detail="Item not found")
    else:
        raise HTTPException(status_code=404, detail="Item not found")
