package cache

import (
	"time"

	"github.com/pkg/errors"
)

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
	value, ok := c.data[key]
	if !ok || time.Now().After(c.validity[key]) {
		return nil, errors.WithStack(ErrKeyNotFound)
	}

	return value, nil
}

func (c *InMemory) Delete(key string) error {
	delete(c.data, key)
	delete(c.validity, key)
	return nil
}

func (c *InMemory) Clear() error {
	c.data = make(map[string]interface{})
	c.validity = make(map[string]time.Time)
	return nil
}

func (c *InMemory) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		if time.Now().Before(c.validity[key]) {
			keys = append(keys, key)
		}
	}
	return keys
}
