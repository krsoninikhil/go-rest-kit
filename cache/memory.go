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
	if time.Now().After(c.validity[key]) || !ok {
		return nil, errors.WithStack(ErrKeyNotFound)
	}

	return value, nil
}
