import os
import logging
import pathlib
import sqlite3
import hashlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware

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

conn = sqlite3.connect('mercari.sqlite3')
c = conn.cursor()

c.execute("""CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    name TEXT,
    category_id INTEGER,
    image TEXT
)""")

c.execute("""CREATE TABLE IF NOT EXISTS category (
    id INTEGER PRIMARY KEY,
    name TEXT
)""")

conn.commit()
conn.close()

def hash_img(image):
    image = hashlib.sha256(image.strip('jpg').encode('utf-8')).hexdigest() + '.jpg'
    return image

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.get("/items")
def get_all_items():
    conn = sqlite3.connect('mercari.sqlite3')
    c = conn.cursor()
    c.execute("""SELECT * from items""")
    items = c.fetchall()
    r =[]
    for item in items:
        c.execute("""SELECT * from category WHERE id == (?)""",[item[2]])
        r.append({f"name:{item[1]}, category:{(c.fetchone())[1]}, image:{item[3]}"}) 
    
    conn.close()
    return {f"items:{r}"}

@app.get("/search")
async def search_items(keyword: str):
    
    conn = sqlite3.connect('mercari.sqlite3')
    c = conn.cursor()
    c.execute("""SELECT * from items WHERE name == (?)""",[keyword])
    searchByName = c.fetchall()
    if (len(searchByName)!=0):
        r =[]
        for item in searchByName:
            c.execute("""SELECT * from category WHERE id == (?)""",[item[2]])
            r.append({f"name:{item[1]}, category:{(c.fetchone())[1]}, image:{item[3]}"}) 
        return {f"{r}"}

    c.execute("""SELECT * from category WHERE name == (?)""",[keyword])
    searchByCategory = c.fetchone()
    if (searchByCategory):
        c.execute("""SELECT * from items WHERE category_id == (?)""",[searchByCategory[0]])
        r =[]
        for item in (c.fetchall()):
            r.append({f"name:{item[1]}, category:{searchByCategory[1]}, image:{item[3]}"}) 
        return {f"{r}"}
    return("Sorry items are not found, please search for keyword name or category ")

@app.get("/items/{item_id}")
def get_by_id(item_id: int):
    conn = sqlite3.connect('mercari.sqlite3')
    c = conn.cursor()
    c.execute("""SELECT * from items WHERE id == (?)""",[item_id])
    r = c.fetchone()
    c.execute("""SELECT * from items WHERE id == (?)""",[r[2]])
    id = c.fetchone()
    r = {f"name:{r[1]} category:{id[1]} image:{r[3]}"}
    conn.close()
    return(f"{r}")


@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: str = Form(...)):
    conn = sqlite3.connect('mercari.sqlite3')
    c = conn.cursor()

    c.execute("""SELECT * from category WHERE name == (?)""",[category])
    categoryData = c.fetchone()
    if(categoryData == None):
        c.execute("""INSERT INTO category VALUES (?,?)""",(None,category))
        c.execute("""SELECT * from category WHERE name == (?)""",[category])
        categoryData = c.fetchone()
    
    c.execute("""INSERT INTO items VALUES (?,?,?,?)""",(None,name,categoryData[0],hash_img(image)))
    conn.commit()
    conn.close()

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
