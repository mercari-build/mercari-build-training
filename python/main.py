import sqlite3
from sqlite3 import Connection
import os
import logging
import pathlib
import hashlib
import json
from fastapi import FastAPI, Form, HTTPException, UploadFile, Query
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles

# os.chdir('/Users/xiaotongye/Programs/mercari-build-training/python')
path = pathlib.Path(__file__).parent.parent.resolve()
# path = pathlib.Path(__file__).parent.resolve()
print(path)

app = FastAPI()
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
app.mount("/static", StaticFiles(directory="images", html=True), name="static")

# sql_file = path / "mercari.sqlite3"

sql_file = path / "db" / "mercari.sqlite3"
# sql_file = path / ".." / "db" / "mercari.sqlite3"
print('sql_file=====', sql_file)

images_dir = path / "images"
# images_dir = path / ".." / "images"
print('images===', images_dir)
logging.basicConfig(level=logging.DEBUG)


# 4-1 Create items table
def create_table():
    # create a database
    try:
        sql_connect = sqlite3.connect(sql_file, check_same_thread=False)
    except sqlite3.Error as e:
        logger.info("Failed to open the table.")       
        raise HTTPException(status_code=500, detail=str(e))  
    sql_cur = sql_connect.cursor()
    sql_cur.execute('''
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL
        )
    ''')    
    sql_cur.execute('''
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER,
            image_name TEXT,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        )
    ''')
    sql_connect.commit()

sql_connect: Connection

async def start_connection():
    
    try:
        global sql_connect
        sql_connect = sqlite3.connect(sql_file, check_same_thread=False)
    except sqlite3.Error as e:
        logger.info("Failed to open the table.")       
        raise HTTPException(status_code=500, detail=str(e))  
    sql_cur = sql_connect.cursor()
    sql_cur.execute('''
        CREATE TABLE IF NOT EXISTS categories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL
        )
    ''')    
    sql_cur.execute('''
        CREATE TABLE IF NOT EXISTS items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            category_id INTEGER,
            image_name TEXT,
            FOREIGN KEY (category_id) REFERENCES categories(id)
        )
    ''')
    sql_connect.commit()
    print('666')

async def close_connection():
    if sql_connect:
        sql_connect.close()

app.add_event_handler("startup", start_connection)
app.add_event_handler("shutdown", close_connection)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

# 4-1 Return item list from items table
@app.get("/items")
def get_items():
    sql_cur = sql_connect.cursor()

    try:
        sql_cur.execute("""
            SELECT items.id, items.name, categories.name AS category_name, image_name s
            FROM items
            INNER JOIN categories ON items.category_id = categories.id
        """)
        item_data = sql_cur.fetchall()
    except sqlite3.Error as e:
        logger.info("Failed to fetch data from the table.")  
        raise HTTPException(status_code=500, detail=str(e))
    # finally:
    #     sql_connect.close()

    items_list = [
        {"id": item[0], "name": item[1], "category": item[2], "image_name": item[3]}
        for item in item_data
    ]
    logger.info("Get all items successfully!") 
    return {"items": items_list}


# 4-1 Return an item by its item id
@app.get("/items/{item_id}")
def get_item(item_id: int):
    logger.info(f"Searching for the item with id: {item_id}")
    
    print(sql_connect)       
    sql_cur = sql_connect.cursor()

    try:
        sql_cur.execute("""
            SELECT items.id, items.name, categories.name AS category_name, image_name s
            FROM items
            INNER JOIN categories ON items.category_id = categories.id
            WHERE items.id = ?
        """, (item_id,))
        item_data = sql_cur.fetchone()
    except sqlite3.Error as e:
        logger.info("Failed to fetch data from the table.")  
        raise HTTPException(status_code=500, detail=str(e))
    # finally:
    #     sql_connect.close()

    if item_data:
        item_dict = {
            "id": item_data[0],
            "name": item_data[1],
            "category": item_data[2],
            "image_name": item_data[3]
        }
        logger.info(f"The search is completed successfully!")
        return item_dict
    else:
        logger.info("Failed to find the item.")  
        raise HTTPException(status_code=404, detail="Item not found.")


def get_category_id(category, sql_cur):
    sql_cur.execute("SELECT id FROM categories WHERE name = ?", (category,))
    category_id = sql_cur.fetchone()
    if category_id:
        return category_id[0]
    sql_cur.execute("INSERT INTO categories (name) VALUES (?)", (category,))
    return sql_cur.lastrowid

# 4-1 Add a new item to the items table
@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = Form(...)):
    logger.info(f"Receive item: {name}")

    # preprocess the image file
    try:
        img_bytes = image.file.read()
        print(img_bytes)
        img_name = hashlib.sha256(img_bytes).hexdigest() + os.path.splitext(image.filename)[1]
        img_path = images_dir / img_name
        # Here, the img path must be pathlib.Path instead of a str!!
        img_path.write_bytes(img_bytes)
    except:
        logger.info("Failed to load the image.") 
        raise HTTPException(status_code=500, detail='Failed to load the image.')
    
    sql_cur = sql_connect.cursor()

    try:
        category_id =  get_category_id(category, sql_cur)
        sql_cur.execute("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)", 
                       (name, category_id, img_name))
        sql_connect.commit()
    except sqlite3.Error as e:
        sql_connect.rollback()
        logger.info("Failed to add the item to the table.") 
        raise HTTPException(status_code=500, detail=str(e))  

    logger.info(f"The item {name} is received successfully!")
    return {"message": f"item received: {name}"}


@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images_dir / image_name

    if not image_name.endswith(".jpg"):
        logger.info("Failed to find the image.") 
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images_dir / "default.jpg"

    return FileResponse(image)

# 4-2 Search for an item
@app.get("/search")
def search_items(keyword: str = Query(None, min_length=1)):
    logger.info(f"Searching keyword: {keyword}")

    try:
        sql_connect = sqlite3.connect(sql_file)
    except sqlite3.Error as e:
        logger.info("Failed to open the table.") 
        raise HTTPException(status_code=500, detail=str(e))  
    sql_cur = sql_connect.cursor()

    try:
        keyword = f"%{keyword}%"
        sql_cur.execute("""
            SELECT items.id, items.name, items.image_name, categories.name AS category_name
            FROM items
            INNER JOIN categories ON items.category_id = categories.id
            WHERE items.name LIKE ?
        """, (keyword,))
        items_data = sql_cur.fetchall()
        sql_connect.commit()
    except sqlite3.Error as e:
        sql_connect.rollback()
        logger.info("Failed to fetch data from the table.") 
        raise HTTPException(status_code=500, detail=str(e))
    # finally:
    #     sql_connect.close()

    items_list = [
        {"id": item[0], "name": item[1], "image_name": item[2], "category": item[3]}
        for item in items_data
    ]

    logger.info("The search is completed successfully!")

    return items_list
