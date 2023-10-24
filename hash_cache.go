package cache

import (
	"runtime"
	"sync"
	"time"
)

type hashCache struct {
	items               map[string]*item
	defaultExpiration   time.Duration
	putListeners        []PutListener
	removeListeners     []RemoveListener
	expirationListeners []ExpirationListener
	mutex               sync.RWMutex
}

type key_value struct {
	key   string
	value interface{}
}

func (c *hashCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, ok := c.items[key]
	if !ok || item.expired() {
		return nil, false
	}
	return item.value, true
}

func (c *hashCache) GetExpiration(key string) (time.Duration, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, ok := c.items[key]
	if !ok {
		// NotFound
		return 0, false
	}
	return item.expiration, true
}

func (c *hashCache) GetExpectedExpiration(key string) (time.Duration, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, ok := c.items[key]
	if !ok {
		// NotFound
		return 0, false
	}
	if item.expiration == NoExpiration {
		return NoExpiration, true
	}
	if item.expired() {
		return Expired, true
	}
	return item.expectedExpiration(), true
}

func (c *hashCache) put(key string, value interface{}, expiration time.Duration) {
	c.mutex.Lock()
	c.items[key] = newItem(value, expiration, c.defaultExpiration)
	c.mutex.Unlock()
	if len(c.putListeners) > 0 {
		c.mutex.RLock()
		listeners := c.putListeners
		c.mutex.RUnlock()
		for _, listener := range listeners {
			listener(key, value)
		}
	}
}

func (c *hashCache) Put(key string, value interface{}, expiration time.Duration) {
	c.put(key, value, expiration)
}

func (c *hashCache) PutIfAbsent(key string, value interface{}, expiration time.Duration) bool {
	if c.Exists(key) {
		return false
	}
	c.put(key, value, expiration)
	return true
}

func (c *hashCache) PutIfExists(key string, value interface{}, expiration time.Duration) bool {
	if !c.Exists(key) {
		return false
	}
	c.put(key, value, expiration)
	return true
}

func (c *hashCache) Exists(key string) bool {
	c.mutex.RLock()
	item, exists := c.items[key]
	defer c.mutex.RUnlock()
	return exists && !item.expired()
}

func (c *hashCache) Remove(key string) bool {
	var value interface{}
	c.mutex.Lock()
	if item, exists := c.items[key]; !exists || item.expired() {
		c.mutex.Unlock()
		return false
	} else {
		value = item.value
	}
	delete(c.items, key)
	c.mutex.Unlock()
	if len(c.removeListeners) > 0 {
		c.mutex.RLock()
		listeners := c.removeListeners
		c.mutex.RUnlock()
		for _, listener := range listeners {
			listener(key, value)
		}
	}
	return true
}

func (c *hashCache) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

func (c *hashCache) Keys() []string {
	keys := make([]string, 0)
	c.ForEach(func(key string, value interface{}) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

func (c *hashCache) Values() []interface{} {
	values := make([]interface{}, 0)
	c.ForEach(func(key string, value interface{}) bool {
		values = append(values, value)
		return true
	})
	return values
}

func (c *hashCache) ForEach(consumer Consumer) {
	kvs := make([]*key_value, 0)
	c.mutex.RLock()
	for key, item := range c.items {
		if item.expired() {
			continue
		}
		kvs = append(kvs, &key_value{
			key:   key,
			value: item.value,
		})
	}
	c.mutex.RUnlock()
	for _, kv := range kvs {
		if !consumer(kv.key, kv.value) {
			break
		}
	}
}

func (c *hashCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]*item)
}

func (c *hashCache) DeleteExpired() {
	expireds := make([]*key_value, 0)
	c.mutex.Lock()
	for key, item := range c.items {
		if item.expired() {
			expireds = append(expireds, &key_value{key: key, value: item.value})
		}
	}
	for _, expired := range expireds {
		delete(c.items, expired.key)
	}
	c.mutex.Unlock()
	if len(c.expirationListeners) > 0 {
		c.mutex.RLock()
		listeners := c.expirationListeners
		c.mutex.RUnlock()
		for _, expired := range expireds {
			for _, listener := range listeners {
				listener(expired.key, expired.value)
			}
		}
	}
}

func (c *hashCache) AddPutListener(listener PutListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.putListeners = append(c.putListeners, listener)
}

func (c *hashCache) AddRemoveListener(listener RemoveListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.removeListeners = append(c.removeListeners, listener)
}

func (c *hashCache) AddExpirationListener(listener ExpirationListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.expirationListeners = append(c.expirationListeners, listener)
}

func NewHashCache(defaultExpiration, cleanupInterval time.Duration) Cache {
	cache := &hashCache{
		items:               make(map[string]*item),
		defaultExpiration:   defaultExpiration,
		putListeners:        make([]PutListener, 0),
		removeListeners:     make([]RemoveListener, 0),
		expirationListeners: make([]ExpirationListener, 0),
	}
	if cleanupInterval > 0 {
		j := &janitor{
			interval: cleanupInterval,
			stop:     make(chan struct{}),
		}
		go j.run(cache)
		runtime.SetFinalizer(cache, j.stopJanitor)
	}
	return cache
}
