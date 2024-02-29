import uvicorn
from fastapi import FastAPI, HTTPException, File, UploadFile, Form
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import json
import hashlib
import os
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse

app = FastAPI()

os.makedirs("images", exist_ok=True)

app.mount("/static", StaticFiles(directory="images"), name="static")

class Item(BaseModel):
    name: str
    category: str

items = []

try:
    with open("items.json", "r") as f:
        items = json.load(f)["items"]
except FileNotFoundError:
    items = []

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
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    content = await image.read()
    sha256_hash = hashlib.sha256(content).hexdigest()
    image_filename = f"{sha256_hash}.jpg"
    image_path = os.path.join('images', image_filename)

    with open(image_path, 'wb') as f:
        f.write(content)

    item = {"name": name, "category": category, "image_name": image_filename}
    if any(i["name"] == item["name"] for i in items):
        raise HTTPException(status_code=400, detail="Item already exists")
    items.append(item)
    save_items()
    return {"message": f"Item received: {name}", "item": item}

@app.get("/items/{item_id}")
def get_item(item_id: int):
    if item_id < 0 or item_id >= len(items):
        raise HTTPException(status_code=404, detail="Item not found")
    
    return items[item_id]

@app.get("/image/{item_id}.jpg")
def serve_image(item_id: int):
    if item_id < 0 or item_id >= len(items):
        raise HTTPException(status_code=404, detail="Item not found")
    
    item = items[item_id]
    image_filename = item.get("image_name")
    if not image_filename:
        raise HTTPException(status_code=404, detail="Image not found")

    image_path = os.path.join('images', image_filename)
    if not os.path.exists(image_path):
        raise HTTPException(status_code=404, detail="Image file not found")
    
    return FileResponse(image_path, media_type='image/jpeg')

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=9000)
