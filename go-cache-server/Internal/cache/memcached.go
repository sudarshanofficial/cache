package cache

import "github.com/bradfitz/gomemcache/memcache"

type MemcachedCache struct {
	client *memcache.Client
	ttl    int32
}

func NewMemcachedCache(server string, ttl int32) *MemcachedCache {
	client := memcache.New(server)
	return &MemcachedCache{client: client, ttl: ttl}
}

func (m *MemcachedCache) Get(key string) (string, error) {
	item, err := m.client.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (m *MemcachedCache) Set(key string, value string) error {
	return m.client.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: m.ttl})
}
func (m *MemcachedCache) SetWithTTL(key string, value string, expiration int32) error {
	return m.client.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: m.ttl})
}

func (m *MemcachedCache) Delete(key string) error {
	return m.client.Delete(key)
}

func (m *MemcachedCache) Clear() error {
	return m.client.FlushAll()
}
