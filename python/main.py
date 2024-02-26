import os
import logging
import pathlib
import json
import hashlib

from fastapi import FastAPI, Form, UploadFile, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware


app = FastAPI()

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

#items_json = pathlib.Path(__file__).parent.resolve() / "items.json"

#モード
#r:読み込み（デフォルト）
#w:書き込み
#a:追記
#r+:読み込みと書き込み

@app.get("/")
async def root():
    return {"message": "Hello, world!"}

@app.post("/items")
async def add_item(name: str = Form(...), category: str=Form(...), image: UploadFile= File(...)):
    logger.info(f"Receive item: {name},{category},image:{image.filename}")
    
    images=pathlib.Path(__file__).parent/"python"/"images"
    #アップロードされたファイルの内容を非同期で読み込む
    contents= await image.read()
    hash_sha256=hashlib.sha256(contents).hexdigest()
    image_filename=f"{hash_sha256}.jpg"
    image_path=images/image_filename
    with open(image_path,"wb") as f:
        f.write(contents)
        
    #loadでjsonファイルをPythonのデータ構造に変換する
    with open("items.json","r") as f:
        json_load=json.load(f)
    
    new_item={"items": [{"name": name, "category": category,"image_name": image_filename}]}
    json_load.append(new_item)
    #書き込みモードにしてnew_itemをitems.jsonに追加する
    with open("items.json","w") as f:
        json_dump=json.dump(json_load, f,indent=4)
    

    return {"message": f"item received: {name},{category},{image}"}

@app.get("/items")
def get_items():
    return json_load

@app.get("/items/{items_id}")
def get_items(items_id:int):

@app.get("/image/{image_name}")
async def get_image(image_name):
    #image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)