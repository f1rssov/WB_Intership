package cache

import (
	"sync"
	"wbtask/internal/model"
	lru "github.com/hashicorp/golang-lru"
)

type OrderCache struct {
	cache *lru.Cache
	mu    sync.RWMutex
}

func NewOrderCache(size int) (*OrderCache, error) {
	c, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &OrderCache{cache: c}, nil
}

func (c *OrderCache) Get(key string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val, ok := c.cache.Get(key); ok {
		return val.(*model.Order), true
	}
	return nil, false
}

func (c *OrderCache) Set(key string, value *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Add(key, value)
}
