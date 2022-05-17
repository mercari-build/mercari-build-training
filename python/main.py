# uvicorn main:app --reload --port 9000 で起動
import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

import sqlite3
import hashlib

app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "image"
origins = [ os.environ.get('FRONT_URL', 'http://localhost:3000') ]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET","POST","PUT","DELETE"],
    allow_headers=["*"],
)

db_file = pathlib.Path(__file__).parent.resolve() / "db" / "items.db"
sqlite3_file = pathlib.Path(__file__).parent.resolve() / "db" / "mercari.sqlite3"

@app.on_event("startup")
def start_app() -> None:
    logger.info("Stating the app...")
    if not os.path.exists(db_file):
        raise FileNotFoundError
    
    if not os.path.exists(sqlite3_file):
        sqlite3_file.touch()

    conn = sqlite3.connect(sqlite3_file)
    cur = conn.cursor()

    with open(db_file, 'r') as file:
        schema = file.read()

    print(schema)

    cur.executescript(f"""
            {schema}
        """)
    conn.commit()
    conn.close()

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):
    logger.info(f"Receive item: {name}, {category}")

    image = image.split('/')[-1]

    if not image.endswith('.jpg'):
        raise HTTPException(
            status_code=400, detail="Image is not in \".jpg\" format"
        )

    hash_filename = hashlib.sha256(image.encode('utf-8')).hexdigest() + '.jpg'
    print(hash_filename)

    conn = sqlite3.connect(sqlite3_file)
    cur = conn.cursor()

    cur.execute(f"INSERT OR IGNORE INTO category (name) VALUES ('{category}')")
    cur.execute(f"SELECT id FROM category WHERE name='{category}'")

    category_id = cur.fetchone()[0]
    print(f'category_id: {category_id}')

    cur.execute(f"INSERT INTO items (name, category_id, image) VALUES ('{name}', '{category_id}', '{hash_filename}')")

    conn.commit()
    conn.close()

    return {"name": name, "category": category}

@app.get("/items")
def get_items():
    conn = sqlite3.connect(sqlite3_file)
    cur = conn.cursor()

    cur.execute("SELECT * FROM items")

    data = cur.fetchall()
    conn.close()

    return data

@app.get("/items/{item_id}")
def get_by_item_id(item_id: int):
    conn = sqlite3.connect(sqlite3_file)
    cur = conn.cursor()

    print(f"SELECT * FROM items WHERE id={item_id}")

    cur.execute(f"SELECT name, category, image FROM items WHERE id={item_id}")

    data = cur.fetchall()
    conn.close()

    return data

@app.get("/search")
def search_items(keyword: str):
    conn = sqlite3.connect(sqlite3_file)
    print(keyword)
    print('here')
    cur = conn.cursor()

    cur.execute(f"SELECT * FROM items WHERE name='{keyword}' OR category='{keyword}'")

    data = cur.fetchall()
    conn.close()

    return data

@app.get("/image/{items_image}")
async def get_image(items_image):
    # Create image path
    image = images / items_image

    if not items_image.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)

@app.on_event("shutdown")
def close_app():
    logger.info("Closing the app...")
