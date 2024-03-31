# Importing necessary libraries for the application
import os  # Provides functions for interacting with the operating system
import logging  # Provides logging capabilities
import pathlib  # Provides classes for working with filesystem paths
import hashlib  # Provides cryptographic hashing functions
import json  # Provides functions for working with JSON data

# Importing FastAPI and related modules
from fastapi import FastAPI, Form, Path, HTTPException  # Provides tools for building APIs
from fastapi.responses import FileResponse  # Provides tools for handling API responses
from fastapi.middleware.cors import CORSMiddleware  # Provides middleware for CORS handling

# Creating a FastAPI instance
app = FastAPI()

# Configuring logging
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO

# Defining paths
current_dir = pathlib.Path(__file__).resolve().parent  # Getting the current directory
images = current_dir / "images"  # Directory containing images
items_file = current_dir / "items.json"  # File for storing item data in JSON format

# CORS settings:
# - Allow Origins: List of origins (domains) that are allowed to access the resources on the server.
# - Allow Methods: List of HTTP methods allowed for cross-origin requests (e.g., GET, POST, PUT, DELETE).
# - Allow Headers: List of HTTP headers allowed to be included in cross-origin requests.
# - Allow Credentials: Boolean indicating whether cross-origin requests can include credentials (e.g., cookies, HTTP authentication).

# Configuring CORS middleware
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# Helper functions

def load_items_from_json():
    """Load items data from the JSON file."""
    try:
        with open(items_file, "r") as file:
            items = json.load(file)
    except FileNotFoundError:
        items = []
    return items

def save_items_to_json(items):
    """Save items data to the JSON file."""
    with open(items_file, "w") as file:
        json.dump(items, file, indent=2)

# Endpoints

@app.get("/")
def root():
    """Root endpoint to check if the API is running."""
    return {"message": "Hello world!"}

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...), image_path: str = Form(...)):
    """Endpoint to add a new item."""
    logger.info(f"Received item: {name}")
    
    items_data = load_items_from_json()
    
    # Converting the image filename to image_hash
    image_filename = os.path.basename(image_path)
    image_hash = hashlib.sha256(image_filename.encode()).hexdigest()
    
    # Appending the new item
    new_item = {"name": name, "category": category, "image": f"{image_hash}.jpg"}
    items_data.append(new_item)
    
    # Rewriting the file with updated data
    save_items_to_json(items_data)
    
    return {"message": f"Item received: {name}"}

@app.get("/items")
def get_items():
    """Endpoint to get all items."""
    items_data = load_items_from_json()
    return {"items": items_data}

@app.get("/items/{item_id}")
def get_one_item(item_id: int = Path(..., title="The ID of the item to retrieve")):
    """Endpoint to get a single item by ID."""
    items_data = load_items_from_json()
    if 0 <= item_id - 1 < len(items_data):
        return items_data[item_id - 1]
    else:
        raise HTTPException(status_code=404, detail="Item not found")

@app.get("/image/{image_name}")
async def get_image(image_name: str):
    """Endpoint to get an image by name."""
    # Create image path
    image_path = images / image_name
    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")
    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = images / "default.jpg"
    return FileResponse(image_path)
