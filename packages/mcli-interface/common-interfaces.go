package mcliinterface

type Cacher interface {
	GetAndSetIfNotExists(key string, value ...interface{}) (interface{}, error)
	Set(key string, value ...interface{}) (interface{}, error)
	Remove(key string) error
	Optimize() bool
}
