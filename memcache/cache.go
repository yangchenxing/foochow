package memcache

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var (
	noExpiration  time.Time
	ErrNoCapacity = errors.New("容量为0")
	ErrNotFound   = errors.New("未找到")
	ErrExpired    = errors.New("数据过期")
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

func (cache *Cache) Size() int {
	return len(cache.data)
}

func (cache *Cache) Get(key string) (interface{}, bool, error) {
	cache.Lock()
	defer cache.Unlock()
	if cache.Capacity == 0 {
		return nil, false, ErrNoCapacity
	}
	if cache.data == nil {
		cache.data = make(map[string]*item)
		cache.queue = list.New()
	}
	value, found := cache.data[key]
	if !found {
		return nil, false, ErrNotFound
	}
	if value.expiration != noExpiration && value.expiration.Before(time.Now()) {
		cache.queue.Remove(value.queueLink)
		delete(cache.data, key)
		return nil, false, ErrExpired
	}
	return value.value, found, nil
}

func (cache *Cache) Set(key string, value interface{}) error {
	cache.Lock()
	defer cache.Unlock()
	if cache.Capacity == 0 {
		return ErrNoCapacity
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
		cache.data[key] = cacheItem
		cache.queue.PushBack(key)
		cacheItem.queueLink = cache.queue.Back()
	}
	return nil
}
