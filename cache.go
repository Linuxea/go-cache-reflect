package cache

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/go-redis/redis"
)

type Cache struct {
	key      string                      // redis key
	redis    *redis.Client               // redis cli
	model    interface{}                 // the dest model pointer
	FetchFun func() (interface{}, error) // execute in case of cache does not exists
	ttl      time.Duration               // cache ttl
}

func (c *Cache) cache() error {

	if reflect.ValueOf(c.model).Kind() != reflect.Ptr {
		return errors.New("model should be pointer")
	}

	// has cache already ?
	s, err := c.redis.Get(c.key).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	// from cache
	if err == nil {
		if err = json.Unmarshal([]byte(s), c.model); err != nil {
			return err
		}

		return nil
	}

	// from realtime execute

	fetchData, err := c.FetchFun()
	if err != nil {
		return err
	}
	// put into cache
	b, err := json.Marshal(fetchData)
	if err != nil {
		return err
	}
	_, err = c.redis.Set(c.key, b, c.ttl).Result()
	if err != nil {
		return err
	}

	reflect.ValueOf(c.model).Elem().Set(reflect.ValueOf(fetchData).Elem())
	return nil
}
