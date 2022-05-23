import os
import logging
import pathlib
import json
import sqlite3
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

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    logger.info(f"Receive item: {name}, category: {category}")
    
    # connect to database
    conn = sqlite3.connect("../db/mercari.sqlite3")
    c = conn.cursor()

    c.execute("INSERT INTO items(name,category) values (?, ?)", (name, category))
    conn.commit()
    conn.close()

    # items_list = {"items": []}
    # check if json file exists
    # if os.path.isfile("items.json"):
    #     with open("items.json", "r") as items_json_f:
    #         # load existing data
    #         items_list = json.load(items_json_f)
    
    # add new item
    # new_item = {"name" : name, "category": category}
    # items_list["items"].append(new_item)
    
    # with open("items.json", "w") as items_json_f:
    #     # write new data to json file
    #     json.dump(items_list, items_json_f)

    return {"message": f"item received: {name}, category: {category}"}

@app.get("/items")
def get_item():
    # connect to database
    conn = sqlite3.connect("../db/mercari.sqlite3")
    c = conn.cursor()

    result = c.execute("SELECT * FROM items").fetchall()
    items_list = {
        "items": [{"id": id, "name": name, "category": category} for (id, name, category) in result]
    }
    
    conn.commit()
    conn.close()

    # items_list = {"items" : []}
    # if os.path.isfile("items.json"):
    #     with open("items.json", "r") as items_json_f:
    #         items_list = json.load(items_json_f)
    return items_list

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
