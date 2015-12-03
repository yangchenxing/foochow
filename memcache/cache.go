package memcache

import (
	"container/list"
	"sync"
	"time"
)

var (
	noExpiration time.Time
)

type item struct {
	value      interface{}
	expiration time.Time
	queueLink  *list.Element
}

type Cache struct {
	sync.Mutex
	Capacity   uint32
	Expiration time.Duration
	data       map[string]*item
	queue      *list.List
}

func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.Lock()
	defer cache.Unlock()
	if cache.Capacity == 0 {
		return nil, false
	}
	if cache.data == nil {
		cache.data = make(map[string]*item)
		cache.queue = list.New()
		return nil, false
	}
	value, found := cache.data[key]
	if !found {
		return nil, false
	}
	if value.expiration != noExpiration && value.expiration.Before(time.Now()) {
		cache.queue.Remove(value.queueLink)
		delete(cache.data, key)
		return nil, false
	}
	return value.value, found
}

func (cache *Cache) Set(key string, value interface{}) {
	cache.Lock()
	defer cache.Unlock()
	if cache.Capacity == 0 {
		return
	}
	if cache.data == nil {
		cache.data = make(map[string]*item)
		cache.queue = list.New()
	}
	if cacheItem, found := cache.data[key]; found {
		cacheItem.value = value
		if cache.Expiration > 0 {
			cacheItem.expiration = time.Now().Add(cache.Expiration)
		}
	} else {
		if cache.queue.Len() == int(cache.Capacity) {
			delete(cache.data, cache.queue.Front().Value.(string))
			cache.queue.Remove(cache.queue.Front())
		}
		cacheItem := &item{
			value: value,
		}
		if cache.Expiration > 0 {
			cacheItem.expiration = time.Now().Add(cache.Expiration)
		}
		cache.queue.PushBack(key)
		cacheItem.queueLink = cache.queue.Back()
	}
}
