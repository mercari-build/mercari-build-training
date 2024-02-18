import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException,File,UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
'''
環境変数
'''
json_path = "sample.json"
db_path = "/Users/horiguchitakahiro/Desktop/mercari2/mercari-build-training/db/mercari.sqlite3"

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
'''
curl -X GET \
    --url 'http://localhost:9000'
'''
@app.get("/")
def root():
    sql_connect = sqlite3.connect(db_path)
    table = "items" # テーブル名
    sql_command = f"""CREATE TABLE IF NOT EXISTS {table} (
    id INTEGER PRIMARY KEY,
    name TEXT,
    category TEXT,
    image BLOB
    );"""
    sql_connect.execute(sql_command) # SQL実行
    print("sql login success")
    sql_connect.close()
    return {"message": "Hello, world!"}
'''
curl -X GET \
    --url 'http://localhost:9000/items' \
'''
@app.get("/items")
def get_items():
    sql_connect = sqlite3.connect(db_path)
    sql_cur = sql_connect.cursor()
    sql_cur.execute('SELECT * FROM items')

    rev = sql_cur.fetchall()
    return rev


@app.get("/items/{item_id}")
def return_item_information(item_id):
    print("receive item = {item_id}")
    return item_id

'''
curl -X POST \
    --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
'''
@app.post("/items")
async def add_item(name: str = Form(),category:str = Form(),image:UploadFile = File()):
    data = await image.read()  # アップロードされた画像をbytesに変換する処理
    sha256 = hashlib.sha256(data).hexdigest()
    print('SHA256ハッシュ値：\n {0}'.format(sha256))
    logger.info(f"Receive item: {name}")
    logger.info(f"Recive category : {category}")
    sql_connect = sqlite3.connect(db_path)
    cur = sql_connect.cursor()
    sqlite_command = f"insert into items(name,category,image) values('{name}','{category}','{sha256}')"

    print(sqlite_command)
    cur.execute(sqlite_command)
    sql_connect.commit()
    cur.close()
    sql_connect.close()
    print("success sqlite insertion")
    #item_dict = {"name":name,"category":category}
    # with open("./edited.json", "w") as js:
    #     json.dump(item_dict, js, indent = 4)
    #     print("success creating json")
    return {"message": f"item received: name = {name} category ={category},image = {sha256}"}


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
