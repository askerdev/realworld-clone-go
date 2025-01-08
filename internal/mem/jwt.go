package mem

import "time"

type JWTCache struct {
	storage map[uint64]time.Time
}

func NewJWTCache() *JWTCache {
	return &JWTCache{}
}

func (c *JWTCache) Get(key uint64) (time.Time, bool) {
	time, ok := c.storage[key]
	return time, ok
}

func (c *JWTCache) Set(key uint64, value time.Time) error {
	if c.storage == nil {
		c.storage = map[uint64]time.Time{}
	}

	c.storage[key] = value

	return nil
}
