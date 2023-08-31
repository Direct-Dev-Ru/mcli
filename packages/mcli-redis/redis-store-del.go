package mcliredis

import (
	"fmt"
)

func (rs *RedisStore) RemoveRecord(key string, keyPrefixes ...string) (err error) {
	resultKey := key
	if len(keyPrefixes) > 0 {
		prefix := keyPrefixes[0]
		resultKey = fmt.Sprintf("%s:%s", prefix, key)
	} else if len(rs.KeyPrefix) > 0 {
		resultKey = fmt.Sprintf("%s:%s", rs.KeyPrefix, key)
	}

	conn := rs.RedisPool.Get()
	defer conn.Close()

	// Key exists, retrieve its value
	_, err = conn.Do("DEL", resultKey)
	return
}

func (rs *RedisStore) RemoveRecords(keys []string, keyPrefixes ...string) error {

	conn := rs.RedisPool.Get()
	defer conn.Close()
	for _, key := range keys {

		resultKeys := []string{key}
		if len(keyPrefixes) > 0 {
			resultKeys = make([]string, 0)
			for _, prefix := range keyPrefixes {
				resultKey := fmt.Sprintf("%s:%s", prefix, key)
				resultKeys = append(resultKeys, resultKey)
			}
		} else if len(rs.KeyPrefix) > 0 {
			resultKeys[0] = fmt.Sprintf("%s:%s", rs.KeyPrefix, key)
		}

		for _, k := range resultKeys {
			_, err := conn.Do("DEL", k)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
