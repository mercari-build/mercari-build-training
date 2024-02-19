import uvicorn
from fastapi import FastAPI, HTTPException, File, UploadFile, Form
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import json
import hashlib
import os

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
async def add_item(name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)):
    # Hash the image content
    content = await image.read()
    sha256_hash = hashlib.sha256(content).hexdigest()
    image_filename = f"{sha256_hash}.jpg"
    image_path = os.path.join('images', image_filename)

    # Save the image file
    with open(image_path, 'wb') as f:
        f.write(content)

    # Add the item with the image reference
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


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=9000)
