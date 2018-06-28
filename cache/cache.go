package cache

import (
	"container/list"
	"sync"
	"time"
)

// Cache is a thread-safe LRU cache in which items also have a time-to-live.
// Items are evicted if the cache capacity is at its max or the item has expired
// since its last `Get` call.
type Cache struct {
	capacity int
	ttl      int // seconds

	mux   sync.Mutex
	lru   *list.List
	cache map[string]*list.Element
}

type item struct {
	key   string
	value []byte
	exp   int64 // UTC timestamp
}

// New creates a new Cache instance.
// If cap is zero, the cache has no max capacity.
// If ttl is zero, the items have no expiration.
func New(cap int, ttl int) *Cache {
	return &Cache{
		capacity: cap,
		ttl:      ttl, // seconds
		lru:      list.New(),
		cache:    make(map[string]*list.Element),
	}
}

// Set adds or updates the item at a specific key, evicting
// other items if necessary.
func (c *Cache) Set(key string, value []byte) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if el, ok := c.cache[key]; ok {
		// Move element to front of lru list.
		c.lru.MoveToFront(el)

		// Update the value of the element.
		el.Value.(*item).value = value

		return
	}

	// Push new element to front of lru list.
	exp := time.Now().Add(time.Second * time.Duration(c.ttl)).Unix()
	el := c.lru.PushFront(&item{key, value, exp})

	// Cache the element.
	c.cache[key] = el

	// Evict if necessary.
	if c.capacity != 0 && c.lru.Len() > c.capacity {
		c.removeLRU()
	}
}

// Get gets an item at a specific key. If the key
// exists it moves the item to the front of the lru list -
// making it the most recently used. (MRU!)
func (c *Cache) Get(key string) (val []byte, ok bool) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Check if the key is in our cache.
	if el, ok := c.cache[key]; ok {
		if c.ttl == 0 || el.Value.(*item).exp > time.Now().Unix() {
			el := c.cache[key]
			c.lru.MoveToFront(el)

			return el.Value.(*item).value, true
		}

		// The value is expired. Evict.
		c.remove(key)
	}

	return
}

// Removes the least recently used item, the tail of the
// lru list.
func (c *Cache) removeLRU() {
	if el := c.lru.Back(); el != nil {
		c.remove(el.Value.(*item).key)
	}
}

// Removes the item from the key-element cache and lru list.
func (c *Cache) remove(key string) {
	el := c.cache[key]
	c.lru.Remove(el)
	delete(c.cache, key)
}
