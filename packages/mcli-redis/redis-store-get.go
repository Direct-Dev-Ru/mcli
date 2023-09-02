package mcliredis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

func (rs *RedisStore) GetRecord(key string, keyPrefixes ...string) (result string, err error, ok bool) {
	if len(key) == 0 {
		return "", nil, false
	}
	prefix := ""
	resultKey := key
	if len(keyPrefixes) > 0 {
		prefix = keyPrefixes[0]
		resultKey = fmt.Sprintf("%s:%s", prefix, key)
	} else if len(rs.KeyPrefix) > 0 {
		resultKey = fmt.Sprintf("%s:%s", rs.KeyPrefix, key)
	}

	conn := rs.RedisPool.Get()
	defer conn.Close()

	var createIfKeyDontExists bool = false

	// Check if the key exists
	// fmt.Println(resultKey)
	if err = rs.KeyExists(conn, resultKey); err != nil {
		// fmt.Println(err)
		if _, ok = err.(*KeyNotFoundError); ok {
			if createIfKeyDontExists {
				createIfKeyDontExists = false
				//  TODO create empty resultKey
				return "", nil, false
			}
			return "", nil, false
		}
		return "", err, false
	} else {
		// Key exists, retrieve its value
		value, err := redis.String(conn.Do("GET", resultKey))
		// fmt.Println(value, err)
		if err != nil {
			return "", err, true
		}
		return value, nil, true
	}
}

func (rs *RedisStore) GetRecordEx(key string, keyPrefixes ...string) (string, int, error) {
	result, err, ok := rs.GetRecord(key, keyPrefixes...)
	if err != nil || !ok {
		return "", -1, err
	}
	if ok {
		conn := rs.RedisPool.Get()
		defer conn.Close()
		resultKey, _ := rs.GetResultKey(key, keyPrefixes...)
		remainingTime, err := redis.Int(conn.Do("TTL", resultKey))
		if err != nil {
			return "", -1, err
		}
		return result, remainingTime, nil
	}
	return "", -1, fmt.Errorf("key do not exists")
}

func (rs *RedisStore) GetRecords(pattern string, keyPrefixes ...string) (map[string]string, error) {
	resultMap := make(map[string]string, 0)
	prefix := ""
	resultPattern := pattern
	if len(keyPrefixes) > 0 {
		prefix = keyPrefixes[0]
		resultPattern = fmt.Sprintf("%s:%s", prefix, pattern)
	} else if len(rs.KeyPrefix) > 0 {
		resultPattern = fmt.Sprintf("%s:%s", rs.KeyPrefix, pattern)
	}

	conn := rs.RedisPool.Get()
	defer conn.Close()

	// Use the SCAN command to iterate through keys matching the pattern
	cursor := 0
	for {
		values, nextCursor, err := scanKeys(conn, cursor, resultPattern)
		if err != nil {
			return nil, err
		}

		for _, key := range values {
			// fmt.Println("Key:", value)

			value, err := redis.String(conn.Do("GET", key))
			if err != nil {
				return nil, err
			}
			resultMap[key] = value
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return resultMap, nil
}

// timeout in millisecond
func (rs *RedisStore) GetRecordWithTimeOut(key string, timeout int, keyPrefixes ...string) (string, error, bool) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond) // Set your desired timeout here
	defer cancel()

	resultChan := make(chan string)
	go func() {
		value, err, _ := rs.GetRecord(key, keyPrefixes...)
		if err != nil {
			resultChan <- fmt.Sprintf("error: %v", err)
		} else {
			resultChan <- value
		}
	}()

	select {
	case result := <-resultChan:
		if strings.HasPrefix(result, "error:") {
			return "", fmt.Errorf("%s", result), false
		}
		return result, nil, true
	case <-ctx.Done():
		return "", fmt.Errorf("command timed out for %v millisecond", timeout), false
	}
}

func scanKeys(conn redis.Conn, cursor int, pattern string) ([]string, int, error) {
	reply, err := conn.Do("SCAN", cursor, "MATCH", pattern)
	if err != nil {
		return nil, 0, err
	}
	// Parse SCAN reply
	values, err := redis.Values(reply, nil)
	if err != nil {
		return nil, 0, err
	}

	nextCursor, err := redis.Int(values[0], nil)
	if err != nil {
		return nil, 0, err
	}

	keys, err := redis.Strings(values[1], nil)

	if err != nil {
		return nil, 0, err
	}
	return keys, nextCursor, err
}
