import os
import json
import logging
import pathlib
import hashlib
from fastapi import FastAPI, Form, UploadFile, File, HTTPException, Path
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

# Define file paths
BASE_DIR = pathlib.Path(__file__).parent.resolve()
ITEMS_FILE = BASE_DIR / "items.json"
IMAGES_DIR = BASE_DIR / "images"  
DB_PATH = BASE_DIR / "db" / "mercari.sqlite3"

# Ensure necessary directories exist
IMAGES_DIR.mkdir(parents=True, exist_ok=True)

# Configure logging
logging.basicConfig(level=logging.DEBUG, format="%(asctime)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

# Function to load items from JSON
def load_items():
    if ITEMS_FILE.exists():
        try:
            with open(ITEMS_FILE, "r", encoding="utf-8") as file:
                return json.load(file).get("items", [])
        except json.JSONDecodeError:
            return []
    return []

# Function to save items to JSON
def save_items(items):
    with open(ITEMS_FILE, "w", encoding="utf-8") as file:
        json.dump({"items": items}, file, indent=4)

# Function to hash images using SHA-256
def hash_image(file_data):
    sha256 = hashlib.sha256()
    sha256.update(file_data)
    return sha256.hexdigest()

# Response models
class Item(BaseModel):
    name: str
    category: str
    image_path: str  

class GetItemsResponse(BaseModel):
    items: list[Item]

# Initialize FastAPI
app = FastAPI()

# Enable CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["GET", "POST", "DELETE"],
    allow_headers=["*"],
)

# Get all items (Replaces "Hello World" response)
@app.get("/items", response_model=GetItemsResponse)
def get_all_items():
    items = load_items()
    return GetItemsResponse(items=items)

# Get a single item by ID
@app.get("/items/{item_id}")
def get_single_item(item_id: int = Path(..., ge=0)):
    items = load_items()
    if item_id >= len(items):
        raise HTTPException(status_code=404, detail="Item not found")
    return items[item_id]

# Add new item with image upload
@app.post("/items")
async def add_item(
    name: str = Form(...), category: str = Form(...), image: UploadFile = File(...)
):
    if not name or not category or not image:
        raise HTTPException(status_code=400, detail="Name, category, and image are required")

    # Read and hash image
    image_data = await image.read()
    hashed_image_name = hash_image(image_data) + ".jpg"
    image_path = IMAGES_DIR / hashed_image_name

    # Save the image
    with open(image_path, "wb") as img_file:
        img_file.write(image_data)

    # Load and update items
    items = load_items()
    new_item = {"name": name, "category": category, "image_path": str(image_path)}
    items.append(new_item)
    save_items(items)

    logger.debug(f"New item added: {new_item}")
    return {"message": f"Item '{name}' added successfully."}

# Get image by filename
@app.get("/images/{image_name}")
async def get_image(image_name: str):
    image_path = IMAGES_DIR / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Invalid image format")

    if not image_path.exists():
        logger.debug(f"Image not found: {image_name}")
        raise HTTPException(status_code=404, detail="Image not found")

    return FileResponse(image_path)
