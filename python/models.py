from sqlalchemy import Column, ForeignKey, Integer, String
from python.database import Base

# SQLAlchemy model
class Items(Base):
    __tablename__ = 'items'
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String)
    category_id = Column(Integer, ForeignKey('categories.id', ondelete='SET NULL'))
    image_name = Column(String)

class Categories(Base):
    __tablename__ = 'categories'
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True)

