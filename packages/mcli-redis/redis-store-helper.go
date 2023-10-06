package mcliredis

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func (rs *RedisStore) KeyExists(conn redis.Conn, key string) error {

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return err
	}
	if !exists {
		return NewKeyNotFoundError(key)
	}
	return nil
}

func (rs *RedisStore) GetResultKey(key string, keyPrefixes ...string) (resultKey string, err error) {
	prefix := ""
	err = nil
	if len(keyPrefixes) > 0 {
		prefix = keyPrefixes[0]
		resultKey = fmt.Sprintf("%s:%s", prefix, key)
	} else if len(rs.KeyPrefix) > 0 {
		resultKey = fmt.Sprintf("%s:%s", rs.KeyPrefix, key)
	}
	return
}

func (rs *RedisStore) ExecuteCommand(conn redis.Conn, command string, args ...interface{}) (interface{}, error) {
	r, err := conn.Do(command, args...)
	if err != nil {
		return nil, err
	}
	return r, nil
}
