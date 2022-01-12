package cache

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/go-redis/redis"
)

type Cache struct {
	key      string                      // redis key
	redis    *redis.Client               // redis cli
	model    interface{}                 // the data model instance, require reflect.Type from it
	FetchFun func() (interface{}, error) // execute in case of cache does not exists
	ttl      time.Duration               // cache ttl
}

func (c *Cache) cache() error {

	// has cache already ?
	s, err := c.redis.Get(c.key).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	// from cache
	if err == nil {
		result := reflect.New(reflect.TypeOf(c.model)).Interface()
		if err = json.Unmarshal([]byte(s), result); err != nil {
			return err
		}

		reflect.ValueOf(c.model).Elem().Set(reflect.ValueOf(result).Elem())
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
