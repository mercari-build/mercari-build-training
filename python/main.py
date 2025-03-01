import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
import sqlite3

from contextlib import asynccontextmanager

DATABASE = "fastapi.db"

def get_db():
    conn = sqlite3.connect(DATABASE)
    conn.row_factory = sqlite3.Row # Return rows as dictionaries
    try:
        yield conn
    finally:
        conn.close()

@asynccontextmanager
async def lifespan(app: FastAPI):
    await create_tables_on_startup()
    yield

async def create_tables_on_startup():
    conn = sqlite3.connect("fastapi.db")
    cursor = conn.cursor()
    cursor.execute(
        """CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255),
		category VARCHAR(255)
	)"""
    )
    conn.commit()
    conn.close()

app = FastAPI(lifespan=lifespan)
logger = logging.getLogger("uvicorn")
logger.level = logging.INFO
images = pathlib.Path(__file__).parent.resolve() / "images"
origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {"message": "Hello, world!"}


@app.post("/items")
def add_item(name: str = Form(...), db: sqlite3.Connection = Depends(get_db)):
    logger.info(f"Receive item: {name}")
    cursor = db.cursor()
    cursor.execute("INSERT INTO items (name, category) VALUES (?, ?)", (name, "test"))
    db.commit()
    return {"message": f"item received: {name}"}


@app.get("/image/{image_name}")
async def get_image(image_name):
    # Create image path
    image = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image.exists():
        logger.debug(f"Image not found: {image}")
        image = images / "default.jpg"

    return FileResponse(image)


