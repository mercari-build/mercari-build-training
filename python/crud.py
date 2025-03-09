from sqlalchemy.orm import Session
from python import schemas
from python import models
import hashlib
import pathlib

# Define the path to the images
images = pathlib.Path(__file__).parent.resolve() / "images"

def read_items(db: Session):
    items = db.query(models.Items.name, models.Categories.name, models.Items.image_name).\
               join(models.Categories, models.Items.category_id == models.Categories.id).all()
    return [schemas.Item(name=item[0], category=item[1], image_name=item[2]) for item in items]

def read_item(db: Session, item_id: int):
    item = db.query(models.Items.name, models.Categories.name, models.Items.image_name).\
              join(models.Categories, models.Items.category_id == models.Categories.id).\
              filter(models.Items.id == item_id).first()
    return schemas.Item(name=item[0], category=item[1], image_name=item[2])

def search_item(db: Session, keyword: str):
    items = db.query(models.Items.name, models.Categories.name).\
               join(models.Categories, models.Items.category_id == models.Categories.id).\
               filter(models.Items.name.ilike(f"%{keyword}%")).all()
    return [schemas.SearchItem(name=item[0], category=item[1]) for item in items]

def create_item(db: Session, item: schemas.Item):
    category = db.query(models.Categories).\
                  filter(models.Categories.name == item.category).first()

    if category is None: 
        category = models.Categories(name=item.category) 
        db.add(category) 
        db.commit()  
        db.refresh(category) 

    category_id = category.id

    db_item = models.Items(
        name=item.name,
        category_id=category_id, 
        image_name=item.image_name
    )
    db.add(db_item)
    db.commit()
    db.refresh(db_item)
    return db_item



async def upload_image(image):
    #load image
    image_contents = await image.read() 

    # Hashing images with SHA-256
    sha256 = hashlib.sha256(image_contents).hexdigest()
    image_name = f"{sha256}.jpg"  

    # Save images to images directory
    image_path = images / image_name
    with open(image_path, "wb") as f:
        f.write(image_contents)
    return image_name
