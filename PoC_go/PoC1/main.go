package main

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type Item struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// store in-memory
var (
	store        = map[int64]Item{}
	lock         = sync.RWMutex{}
	nextID int64 = 1
	reqCnt int64
)

func main() {
	r := gin.Default()

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// List all items
	r.GET("/items", func(c *gin.Context) {
		lock.RLock()                         // Read lock for concurrent access
		items := make([]Item, 0, len(store)) // Preallocate slice (precise type)
		for _, v := range store {
			items = append(items, v)
		} //iterate over map (store) to collect items by (V)
		lock.RUnlock()               // Unlock after reading
		c.JSON(http.StatusOK, items) // Return items as JSON
	})

	// Get single item
	r.GET("/items/:id", func(c *gin.Context) {
		idStr := c.Param("id") // store id in url to variable
		lock.RLock()           // Read lock for concurrent access
		for _, it := range store {
			if idStr == fmt.Sprintf("%d", it.ID) {
				lock.RUnlock() // Unlock before returning
				c.JSON(http.StatusOK, it)
				return
			} // if idStr matches it.ID, return item
		} // iterate over store to find item (it)
		lock.RUnlock()                                           // Unlock after reading
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"}) // item not found
	})

	// Create item
	r.POST("/items", func(c *gin.Context) {
		var in Item                             // input item
		if err := c.BindJSON(&in); err != nil { // bind JSON to input item (in)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id := atomic.AddInt64(&nextID, 1) - 1 // generate unique ID
		in.ID = id                            // assign ID to input item
		lock.Lock()                           // Write lock for concurrent access
		store[id] = in                        // store item in map
		lock.Unlock()                         // Unlock after writing
		c.JSON(http.StatusCreated, in)        // return created item
	})

	// Count requests
	r.Use(func(c *gin.Context) {
		atomic.AddInt64(&reqCnt, 1) // increment request count
		c.Next()                    // proceed to next handler
	})

	// Metrics
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{ // return metrics as JSON
			"total_requests": atomic.LoadInt64(&reqCnt), // load request count
			"items_count":    len(store),                // count of items in store
		})
	})

	r.Run(":8080") // start server on port 8080
}
