# app.py
from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
from typing import Optional, List
from threading import Lock, RLock
import itertools

app = FastAPI()

# ----- Model (equivalent to Go struct) -----
class Item(BaseModel):
    id: Optional[int] = None
    title: str
    done: bool = False

# ----- In-memory store (thread safe) -----
store = {}
lock = RLock()

# atomic-like counter using itertools
id_counter = itertools.count(1)
req_count = 0
req_lock = Lock()

# ----- Middleware to count requests -----
@app.middleware("http")
async def count_requests(request: Request, call_next):
    global req_count
    with req_lock:
        req_count += 1
    response = await call_next(request)
    return response

# ----- Health -----
@app.get("/health")
def health():
    return {"status": "ok"}

# ----- List items -----
@app.get("/items", response_model=List[Item])
def list_items():
    with lock:
        return list(store.values())

# ----- Get item by ID (same loop logic as Go) -----
@app.get("/items/{id}", response_model=Item)
def get_item(id: int):
    with lock:
        for item in store.values():
            if item.id == id:
                return item
    raise HTTPException(status_code=404, detail="not found")

# ----- Create item -----
@app.post("/items", response_model=Item, status_code=201)
def create_item(in_item: Item):
    new_id = next(id_counter)
    in_item.id = new_id
    with lock:
        store[new_id] = in_item
    return in_item

# ----- Metrics -----
@app.get("/metrics")
def metrics():
    with lock:
        items_count = len(store)

    with req_lock:
        total_requests = req_count

    return {
        "total_requests": total_requests,
        "items_count": items_count
    }
