from fastapi import FastAPI, File, UploadFile, Form
from typing import Optional
import json
import os
from hashlib import sha256

app = FastAPI()

@app.post("/items")
async def create_item(name: str = Form(...), category: str = Form(...), image: Optional[UploadFile] = File(None)):
    contents = await image.read() if image else None
    image_filename = None
    if contents:
        image_hash = sha256(contents).hexdigest()
        image_filename = f"{image_hash}.jpg"
        with open(f"images/{image_filename}", "wb") as f:
            f.write(contents)

    item = {"name": name, "category": category, "image_name": image_filename}
    if os.path.exists('items.json'):
        with open('items.json', 'r+', encoding='utf-8') as file:
            data = json.load(file)
            data['items'].append(item)
            file.seek(0)
            json.dump(data, file, ensure_ascii=False, indent=2)
            file.truncate()
    else:
        with open('items.json', 'w', encoding='utf-8') as file:
            json.dump({'items': [item]}, file, ensure_ascii=False, indent=2)
    
    return {"message": "Item received", "item": item}

@app.get("/items")
async def read_items():
    if os.path.exists('items.json'):
        with open('items.json', 'r', encoding='utf-8') as file:
            data = json.load(file)
            return data
    else:
        return {"items": []}
