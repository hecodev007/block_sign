package cache

import (
	"reflect"
	"time"

	"github.com/robfig/go-cache"
)

var CacheStores *InMemoryStore

type InMemoryStore struct {
	cache.Cache
}

func NewInMemoryStore(defaultExpiration time.Duration) *InMemoryStore {
	if CacheStores == nil {
		CacheStores = &InMemoryStore{*cache.New(defaultExpiration, time.Minute)}
	}
	return CacheStores
}

func (c *InMemoryStore) Get(key string, value interface{}) error {
	val, found := c.Cache.Get(key)
	if !found {
		return ErrCacheMiss
	}
	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
		v.Elem().Set(reflect.ValueOf(val))
		return nil
	}
	return ErrNotStored
}

func GetCacheStore() *InMemoryStore {
	if CacheStores == nil {
		CacheStores = NewInMemoryStore(time.Duration(30))
	}
	return CacheStores
}

func (c *InMemoryStore) Set(key string, value interface{}, expires time.Duration) error {
	// NOTE: go-cache understands the values of DEFAULT and FOREVER
	c.Cache.Set(key, c.value(value), expires)
	return nil
}

func (c *InMemoryStore) Add(key string, value interface{}, expires time.Duration) error {
	err := c.Cache.Add(key, c.value(value), expires)
	if err == cache.ErrKeyExists {
		return ErrNotStored
	}
	return err
}

func (c *InMemoryStore) Replace(key string, value interface{}, expires time.Duration) error {
	if err := c.Cache.Replace(key, c.value(value), expires); err != nil {
		return ErrNotStored
	}
	return nil
}

func (c *InMemoryStore) Delete(key string) error {
	if found := c.Cache.Delete(key); !found {
		return ErrCacheMiss
	}
	return nil
}

func (c *InMemoryStore) Increment(key string, n uint64) (uint64, error) {
	newValue, err := c.Cache.Increment(key, n)
	if err == cache.ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return newValue, err
}

func (c *InMemoryStore) Decrement(key string, n uint64) (uint64, error) {
	newValue, err := c.Cache.Decrement(key, n)
	if err == cache.ErrCacheMiss {
		return 0, ErrCacheMiss
	}
	return newValue, err
}

func (c *InMemoryStore) Flush() error {
	c.Cache.Flush()
	return nil
}

func (c *InMemoryStore) value(value interface{}) interface{} {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		p := v.Elem()
		value = p.Interface()
	}
	return value
}
