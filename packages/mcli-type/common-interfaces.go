package mclitype

import "time"

type Cacher interface {
	GetAndSetIfNotExists(key string, value ...interface{}) (interface{}, error)
	Get(key string) (interface{}, error)
	Set(key string, updateFunc func(params ...interface{}) (interface{}, error), ttl time.Duration, value ...interface{}) (interface{}, error)
	Remove(key string) error
	Optimize(bool) bool
}
