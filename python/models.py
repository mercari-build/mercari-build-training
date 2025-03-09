from sqlalchemy import Column, ForeignKey, Integer, String
from python.database import Base

# SQLAlchemy model
class Items(Base):
    __tablename__ = 'items'
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String)
    category = Column(String)
    image_name = Column(String)