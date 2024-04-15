import uvicorn
from fastapi import FastAPI, HTTPException, File, UploadFile, Form, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import json
import hashlib
import os
import sqlite3

class Item(BaseModel):
    name: str
    category: str

app = FastAPI()

# Load items on server start
with open("items.json", "r") as f:
    items = json.load(f)["items"]

# Ensure the images directory exists
os.makedirs("images", exist_ok=True)

origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Function to save items to file
def save_items():
    with open("items.json", "w") as f:
        json.dump({"items": items}, f)

@app.get("/")
def root():
    return {"message": "Hello, world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    conn = sqlite3.connect('mercari.sqlite3')  
    cur = conn.cursor() 
    # Execute SQL query to insert the new item
    cur.execute("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)", (name, category, image_name))
    conn.commit() 
    conn.close()  
    return {"message": "Item added successfully"}  

@app.get("/items")
def get_items():
    conn = sqlite3.connect('mercari.sqlite3')
    cur = conn.cursor()  
    cur.execute("""
        SELECT items.id, items.name, categories.name AS category_name, items.image_name
        FROM items
        JOIN categories ON items.category_id = categories.id
    """)
    items = cur.fetchall()
    conn.close()
    return {"items": [{"id": item[0], "name": item[1], "category": item[2], "image_name": item[3]} for item in items]}

@app.get("/search")
def search_items(keyword: str = Query(None, title="Search keyword")):
    conn = sqlite3.connect('mercari.sqlite3')
    cur = conn.cursor()
    # Use the LIKE operator in SQL for pattern matching
    cur.execute("SELECT * FROM items WHERE name LIKE ?", ('%' + keyword + '%',))
    items = cur.fetchall()
    conn.close()
    items_list = [{'id': item[0], 'name': item[1], 'category': item[2], 'image_name': item[3]} for item in items]
    return {"items": items_list}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=9000)
