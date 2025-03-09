from pydantic import BaseModel, Field
from typing import List

#  Pydantic model
class Item(BaseModel):
    name: str
    category: str
    image_name: str

    class Config:
        orm_mode = True


class HelloResponse(BaseModel):
    message: str

class AddItemResponse(BaseModel):
    message: str

class GetItemsResponse(BaseModel):
    items: List[Item]

class GetItemResponse(BaseModel):
    item: Item
    

# Search result model
class SearchItem(BaseModel):
    name: str
    category: str

class SearchItemsResponse(BaseModel):
    items: List[SearchItem]
