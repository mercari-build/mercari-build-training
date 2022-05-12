import os
import json
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

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

def write_json(new_data, filename="items.json"):
        with open(filename, 'r') as file:
            file_data = json.load(file)
        with open(filename, 'w') as file:
            file_data["items"].append(new_data)
            json.dump(file_data, file)

@app.post("/items")
def add_item(name: str = Form(...), category: str = Form(...)):
    write_json({"name": name, "category": category})
    logger.info(f"Receive item: {name} in {category}")
    return {"message": f"item received: {name} in {category}"}

#@app.get("/image/{items_image}")
#async def get_image(items_image):
    # Create image path