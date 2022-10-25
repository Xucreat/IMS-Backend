from typing import Union

from fastapi import FastAPI
from pydantic import BaseModel

# 运行方法 
# 1．进入 examples 目录：cd examples
# 2．运行：uvicorn "fastapi_demo:app" --reload --port 8080

app = FastAPI()

class Item(BaseModel):
    name: str
    price: float
    is_offer: Union[bool, None] = None


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.get("/items/{item_id}")
def read_item(item_id: int, q: Union[str, None] = None):
    return {"item_id": item_id, "q": q}


@app.put("/items/{item_id}")
def update_item(item_id: int, item: Item):
    return {"item_name": item.name, "item_id": item_id}
