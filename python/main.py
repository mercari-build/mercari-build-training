import os
import logging
import pathlib
from fastapi import FastAPI, Form, File, UploadFile, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
# additionally imported
import sqlite3
import hashlib


# ----config----------------------------
app = FastAPI()
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get('FRONT_URL', 'http://localhost:3000')]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

json_file = str(pathlib.Path(__file__).parent.resolve() / "items.json")
db_file = str(pathlib.Path(os.path.dirname(__file__)
                           ).parent.resolve() / ".." / "db" / "items.db")
sqlite_file = str(pathlib.Path(os.path.dirname(__file__)
                               ).parent.resolve() / ".." / "db" / "mercari.sqlite3")


# ----endpoints--------------------------
@app.on_event("startup")
def set_up_files():
    # create sqlite3 file if not exist
    if not os.path.exists(sqlite_file):
        open(sqlite_file, 'w').close()

    logger.info("Launching the app...")

    con = sqlite3.connect(sqlite_file)
    cur = con.cursor()

    # update schema
    with open(db_file, encoding='utf-8') as file:
        schema = file.read()
    cur.executescript(f"""{schema}""")
    con.commit()  

    # set up the category
    cur.execute("""SELECT category_id FROM category""")
    result_category = cur.fetchone()

    if result_category is None:
        logger.info("Setting up the category")
        category_list = [("Fashion",), ("Beauty",), ("Home",), ("Furniture",),
                         ("Jewelry",), ("Kids",), ("Toys",), ("Books",)]
        cur.executemany(
            """INSERT INTO category(category_name) VALUES (?)""", category_list)
        con.commit()

    # set up the database
    cur.execute("""SELECT id FROM items""")
    result_items = cur.fetchone()

    if result_items is None:
        logger.info("Setting up the database")
        cur.execute("""INSERT INTO items(name, category_id, image_filename) VALUES (?, ?, ?)""",
                    ("Sample", 0, "sample.jpg"))
        cur.execute("""DELETE FROM items WHERE id=(?)""", (0,))
        con.commit()

    # create size table
    cur.execute("""CREATE TABLE size (
        size_id integer PRIMARY KEY,
        gender text,
        category_id integer NOTNULL)""")
    con.commit()

    # create size_chart table
    cur.execute("""CREATE TABLE size_chart (
        size_id integer PRIMARY KEY,
        country_code text NOT NULL,
        jp_size integer NOT NULL)""")
    con.commit()

    # Add columns to items table
    cur.execute("ALTER TABLE items ADD size_id integer NULL")
    con.commit()

    cur.execute("ALTER TABLE items ADD brand text NULL")
    con.commit()

    con.close()
    logger.info("Finishing set-up")
    return None


@app.get("/")
def root():
    return {"message": "Welcome to Mercari Items Database"}


@app.post("/items")
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    logger.info(f"Receive item - name:{name}, category:{category}")

    if not (image.filename.endswith(".jpg") or image.filename.endswith(".jpeg")):
        raise HTTPException(
            status_code=400, detail="Image is not in .jpg format")

    split_lst = image.filename.split(".")
    hashed_name = f"{hashlib.sha256(split_lst[0].encode('utf-8')).hexdigest()}.{split_lst[1]}"

    image_contents = await image.read()

    image_path = images / hashed_name
    with open(image_path, 'wb') as image_file:
        image_file.write(image_contents)

    con = sqlite3.connect(sqlite_file)
    cur = con.cursor()

    cur.execute(
        "INSERT OR IGNORE INTO category(category_name) VALUES (?)", (category, ))

    # retrieve category id
    cur.execute(
        "SELECT category_id FROM category WHERE category_name = (?)", (category, ))
    category_id = cur.fetchone()[0]  # fetchone --> return (id,)
    # insert item
    cur.execute("""INSERT INTO items(name, category_id, image_filename) VALUES(?,?,?)""",
                (name, category_id, hashed_name))
    con.commit()

     # insert values to size table
    cur.execute("INSERT INTO size(size_id, gender, category_id) VALUES(?,?,?)")
    con.commit()

    # insert values to size_chart table
    # jp_size for women: 3,5,7,9,11,13,15,17,19,21
    # jp_size for men: 36,37,38,39,40,41,42,,43
    # cur.execute("INSERT INTO size_chart(size_id, country_code, jp_size) VALUES(?,?,)")
    # con.commit()

    con.close()
    return {f"message: item received: {name} in {category}"}


@app.get("/items")
def get_item():
    logger.info("Get all items")

    con = sqlite3.connect(sqlite_file)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # select all items
    cur.execute(
        """SELECT items.name, category.category_name as category, 
        items.image_filename FROM items INNER JOIN category 
        ON category.category_id = items.category_id""")

    lst = cur.fetchall()
    con.close()

    if lst == []:
        raise HTTPException(
            status_code=404, detail="No item to list")

    items_json = {"items": lst}
    return items_json


@app.get("/search")
def search_item(keyword: str):  # query parameter
    logger.info(f"Search item with {keyword}")

    con = sqlite3.connect(sqlite_file)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # select item matching keyword
    # cur.execute("SELECT * from items WHERE name LIKE (?)", (f"%{keyword}%", ))
    cur.execute(
        """SELECT items.name, category.category_name as category, 
        items.image_filename FROM items INNER JOIN category ON 
        category.category_id = items.category_id WHERE items.name LIKE (?)""", (f"%{keyword}%", ))

    lst = cur.fetchall()
    con.close()
    if lst == []:
        raise HTTPException(
            status_code=404, detail="No matching item")

    message = {"items": lst}
    return message


@app.get("/items/{items_id}")
def get_item_by_id(items_id: int):
    logger.info(f"Search item with ID: {items_id}")

    con = sqlite3.connect(sqlite_file)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # select item matching keyword
    cur.execute(
        """SELECT items.name, category.category_name as category, 
        items.image_filename FROM items INNER JOIN category 
        ON category.category_id = items.category_id WHERE id=(?)""", (items_id,))
    item = cur.fetchone()
    con.close()
    if item is None:
        raise HTTPException(
            status_code=404, detail="No matching item")
    return item


@app.delete("/items/{items_id}")
def delete_item_by_id(items_id: int):
    logger.info(f"Delete item with ID: {items_id}")

    con = sqlite3.connect(sqlite_file)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # select item matching keyword
    cur.execute(
        """SELECT items.name, category.category_name as category,
        items.image_filename FROM items INNER JOIN category
        ON category.category_id = items.category_id WHERE id=(?)""", (items_id,))
    item = cur.fetchone()

    if item is None:
        raise HTTPException(
            status_code=404, detail="No matching item to delete")

    # no need to inner join with category table
    cur.execute("""DELETE FROM items WHERE id=(?)""", (items_id,))
    con.commit()

    con.close()
    return {f"message: item (ID: {items_id}) deleted"}


@app.get("/image/{image_filename}")
async def get_image(image_filename):
    # Create image path
    image = images / image_filename

    if not image_filename.endswith(".jpg"):
        raise HTTPException(
            status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.info(f"Image not found: {image}")
        image = images / "default.jpg"
    logger.info(f"hello {image}")
    return FileResponse(image)


@app.on_event("shutdown")
def close():
    logger.info("Closing the app...")
