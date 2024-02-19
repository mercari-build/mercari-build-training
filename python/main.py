import os
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, HTTPException,File,UploadFile
from fastapi.responses import FileResponse
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3
import json
"""
source .venv/bin/activate
"""
"""
uvicorn main:app --reload --port 9000
"""
print("debug")

json_path = "sample.json"
# 環境変数の取得
db_path = os.environ.get('DB_PATH')
category_db_path = os.environ.get('CATEGORY_DB_PATH')

# 環境変数が設定されていない場合のデフォルト値の指定
db_path_local = "/Users/horiguchitakahiro/Desktop/mercari2/mercari-build-training/db/main.sqlite3"
category_db_path_local = "/Users/horiguchitakahiro/Desktop/mercari2/mercari-build-training/db/category.sqlite3"
db_path = os.environ.get('DB_PATH', db_path_local)
category_db_path = os.environ.get('CATEGORY_DB_PATH', category_db_path_local)

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
def get_category_id(category_name:str):
    '''
    Returns:
        (inserting_frag,index)
    '''
    conn = sqlite3.connect(category_db_path)
    c = conn.cursor()
    category_table = "category"
    sql_command = f"""SELECT id FROM {category_table} WHERE category = ?"""
    c.execute(sql_command, (category_name,))
    
    # 結果を取得
    data = c.fetchone()
    
    if data:
        conn.close()
        return False,data[0]
    else:
        sql_command = f"""SELECT MAX(id) FROM {category_table}"""
        c.execute(sql_command)
        data = c.fetchone()
        conn.close()
        logger.info(data)
        if data[0]:
            return True,data[0] + 1
        else:
            return True,1


'''
curl -X GET 'http://127.0.0.1:9000'
'''
@app.get("/")
def root():
    sql_connect = sqlite3.connect(db_path)
    table = "items" # テーブル名
    sql_command = f"""CREATE TABLE IF NOT EXISTS {table} (
    id INTEGER PRIMARY KEY,
    name TEXT,
    category INTEGER,
    image BLOB
    );"""
    sql_connect.execute(sql_command) # SQL実行

    category_sql_connect = sqlite3.connect(category_db_path)
    table = "category"
    sql_command = f"""CREATE TABLE IF NOT EXISTS {table} (
    id INTEGER PRIMARY KEY,
    category TEXT
    );"""
    category_sql_connect.execute(sql_command) # SQL実行
    category_sql_connect.close()
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
    json_data = {
    "items": [
        {
            "id": item[0],
            "name":item[1],
            "category": item[2],
            "image_name": item[3]
        } 
        for item in rev
    ]
    }

    return JSONResponse(content=json_data)
'''
curl -X GET \
    --url 'http://localhost:9000/category' \
'''
@app.get("/category")
def get_category():
    sql_connect = sqlite3.connect(category_db_path)
    sql_cur = sql_connect.cursor()
    sql_cur.execute('SELECT * FROM category')

    rev = sql_cur.fetchall()
    return rev

'''
curl -X GET 'http://127.0.0.1:9000/items/1'
'''

@app.get("/items/{item_id}")
def return_item_information(item_id):
    sql_connect = sqlite3.connect(db_path)
    sql_cur = sql_connect.cursor()
    table = "items"
    sql_command = f"""SELECT * FROM {table} WHERE id = {item_id}"""
    sql_cur.execute(sql_command)

    rev = sql_cur.fetchall()
    print(rev)
    if rev:
        item = rev[0]
        json_data = {
            "id": item[0],
            "category": item[1],
            "image_name": item[2]
        
        }

        rev_json = json.dumps(json_data)
    else:
        rev_json = {}
    return rev_json 

'''
curl -X POST \
    --url 'http://localhost:9000/items' \
  -F 'name=jacket' \
  -F 'category=fashion' \
  -F 'image=@images/local_image.jpg'
curl -X POST \
    --url 'http://localhost:9000/items' \
  -F 'name=ps4' \
  -F 'category=game' \
  -F 'image=@images/local_image.jpg'
curl -X POST \
     --url 'http://localhost:9000/items' \
  -F 'name=frieren' \
  -F 'category=book' \
  -F 'image=@images/local_image.jpg'
'''
'''
step4の要件定義として上のコマンドを順に1,2とすると1,2,1と実行してitemsとcategoryのAPI結果を参照すると動作確認ができます。
'''
@app.post("/items")
async def add_item(name: str = Form(),category:str = Form(),image:UploadFile = File()):
    data = await image.read()  # アップロードされた画像をbytesに変換する処理
    sha256 = hashlib.sha256(data).hexdigest()

    sql_connect = sqlite3.connect(db_path)
    category_connect = sqlite3.connect(category_db_path)
    cur = sql_connect.cursor()
    category_cur = category_connect.cursor()
    inserting_flag,category_id = get_category_id(category)

    item_inserting_command = f"insert into items(name,category,image) values('{name}','{category_id}','{sha256}')"

    cur.execute(item_inserting_command)

    sql_connect.commit()
    if inserting_flag:
        category_inserting_command = f"insert into category(category) values('{category}')"
        category_cur.execute(category_inserting_command)
        category_cur.close()
        category_connect.commit()

    cur.close()

    sql_connect.close()
    category_connect.close()
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
