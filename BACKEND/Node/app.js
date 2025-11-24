const express = require('express');
const app = express();

app.use(express.json());

// store in-memory
const store = new Map();
let nextID = 1;
let reqCnt = 0;

// Middleware: count requests 
app.use((req, res, next) => {
  reqCnt += 1;
  next();
});

// Health endpoint
app.get('/health', (req, res) => { 
  res.status(200).json({ status: 'ok' });   // return health status as JSON
});

// List all items
app.get('/items', (req, res) => {
  const items = Array.from(store.values());
  res.status(200).json(items);
});

// Get single item by ID
app.get('/items/:id', (req, res) => {
  const idStr = req.params.id;  // get id from URL parameter
  for (const item of store.values()) {  // iterate over items
    if (idStr === String(item.id)) {    // compare string IDs
      return res.status(200).json(item);    // return found item as JSON
    }
  }
  return res.status(404).json({ error: 'not found' });  // item not found
});

// Create item
app.post('/items', (req, res) => {
  const inItem = req.body;  // get item from request body
  if (!inItem || typeof inItem.title !== 'string') {  // validate input
    return res.status(400).json({ error: 'invalid payload' });  // bad request
  }
  const id = nextID++;  // assign new ID
  const newItem = {
    id,
    title: inItem.title,
    done: Boolean(inItem.done)
  };    // create new item
  store.set(id, newItem);   // store item in memory
  return res.status(201).json(newItem); // return created item as JSON
});

// Metrics
app.get('/metrics', (req, res) => {
  res.status(200).json({
    total_requests: reqCnt,
    items_count: store.size
  });   // return metrics as JSON
});

const PORT = process.env.PORT || 8080;
app.listen(PORT, () => {
  console.log(`PoC Node running on :${PORT}`);  // log server start
});
