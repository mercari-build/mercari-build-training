import os
import logging
import pathlib
from fastapi import FastAPI, Form, HTTPException, Depends, File, UploadFile
from fastapi.responses import FileResponse
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
from sqlalchemy.orm import Session

from python import crud, models, schemas
from python.database import SessionLocal, engine


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


# STEP 5-1: set up the database connection
@asynccontextmanager
async def lifespan(app: FastAPI):
    yield

app = FastAPI(lifespan=lifespan)

models.Base.metadata.create_all(bind=engine)

logger = logging.getLogger("uvicorn")
logger.level = logging.DEBUG

origins = [os.environ.get("FRONT_URL", "http://localhost:3000")]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=False,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)



#Read
@app.get("/", response_model=schemas.HelloResponse)
async def hello():
    return schemas.HelloResponse(**{"message": "Hello, world!"})

@app.get("/items", response_model=schemas.GetItemsResponse)
async def get_items(db: Session = Depends(get_db)):
    items = crud.read_items(db)
    return schemas.GetItemsResponse(items=items)

# Return product details
@app.get("/items/{item_id}", response_model=schemas.GetItemResponse)
def get_item(item_id: int, db: Session = Depends(get_db)):
    item = crud.read_item(db, item_id)
    if not item:
        raise HTTPException(status_code=404, detail="Item not found")
    return schemas.GetItemResponse(item=item)

# Search product
@app.get("/search", response_model=schemas.SearchItemsResponse)
def search_item(keyword: str, db: Session = Depends(get_db)):
    items = crud.search_item(db, keyword)
    if not items:
        raise HTTPException(status_code=404, detail="Item not found")
    return schemas.SearchItemsResponse(items=items)

# return an image (GET /images/{filename})
@app.get("/image/{image_name}")
def get_image(image_name: str):
    # Create image path
    images = pathlib.Path(__file__).parent.resolve() / "images"
    image_path = images / image_name

    if not image_name.endswith(".jpg"):
        raise HTTPException(status_code=400, detail="Image path does not end with .jpg")

    if not image_path.exists():
        logger.debug(f"Image not found: {image_path}")
        image_path = images / "default.jpg"

    return FileResponse(image_path)



#Create
# add new item to database(POST /items)
@app.post("/items", response_model=schemas.AddItemResponse)
async def add_item(
    name: str = Form(...),
    category: str = Form(...),
    image: UploadFile = File(...),
    db: Session = Depends(get_db),
):
    if not name:
        raise HTTPException(status_code=400, detail="name is required")
    if not category:
        raise HTTPException(status_code=400, detail="category is required")
    
    try:
        image_name = await crud.upload_image(image)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Image upload failed: {str(e)}")
    
    item = schemas.Item(name=name, category=category, image_name=image_name)
    created_item = crud.create_item(db=db, item=item)

    return schemas.AddItemResponse(**{"message": f"item received: {created_item.name}"})
