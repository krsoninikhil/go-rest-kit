package cache

import "time"

type InMemory struct {
	data     map[string]interface{}
	validity map[string]time.Time
}

func NewInMemory() *InMemory {
	return &InMemory{
		data:     make(map[string]interface{}),
		validity: make(map[string]time.Time),
	}
}

func (c *InMemory) Set(key string, value interface{}, ttl time.Duration) error {
	c.data[key] = value
	c.validity[key] = time.Now().Add(ttl)
	return nil
}

func (c *InMemory) Get(key string) (interface{}, error) {
	if time.Now().After(c.validity[key]) {
		return nil, ErrKeyNotFound
	}
	return c.data[key], nil
}
