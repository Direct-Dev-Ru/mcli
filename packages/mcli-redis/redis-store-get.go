package mcliredis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

func (rs *RedisStore) GetRecord(key string, keyPrefixes ...string) (result []byte, err error, ok bool) {
	if len(key) == 0 {
		return nil, nil, false
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
				return nil, nil, false
			}
			return nil, nil, false
		}
		return nil, err, false
	} else {
		// Key exists, retrieve its value
		rawValue, err := redis.Bytes(conn.Do("GET", resultKey))
		if err != nil {
			return nil, err, true
		}
		if rs.Encrypt {
			if rs.Cypher == nil {
				return nil, fmt.Errorf("decryption error: %s", "cypher is nil"), false
			}
			rawValue, err = rs.Cypher.Decrypt(rs.encryptKey, rawValue, true)
			if err != nil {
				return nil, fmt.Errorf("decryption error: %w", err), false
			}
		}
		storedData := StoreFormat{}
		err = rs.Unmarshal(rawValue, &storedData)
		if err != nil {
			return nil, err, true
		}
		returnValue := storedData.Value

		switch storedData.ValueType {
		case "string":
			return returnValue, nil, true
		default:
			return returnValue, nil, true
		}

		// return string(rawValue), nil, true
	}
}

func (rs *RedisStore) GetRecordEx(key string, keyPrefixes ...string) ([]byte, int, error) {
	result, err, ok := rs.GetRecord(key, keyPrefixes...)
	if err != nil || !ok {
		return nil, -1, err
	}
	if ok {
		conn := rs.RedisPool.Get()
		defer conn.Close()
		resultKey, _ := rs.GetResultKey(key, keyPrefixes...)
		remainingTime, err := redis.Int(conn.Do("TTL", resultKey))
		if err != nil {
			return nil, -1, err
		}
		return result, remainingTime, nil
	}
	return nil, -1, fmt.Errorf("key do not exists")
}

func (rs *RedisStore) GetRecords(pattern string, keyPrefixes ...string) (map[string][]byte, error) {
	resultMap := make(map[string][]byte, 0)
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
			rawValue, err := redis.Bytes(conn.Do("GET", key))
			if err != nil {
				return nil, err
			}
			if rs.Encrypt {
				if rs.Cypher == nil {
					return nil, fmt.Errorf("decryption error: %s", "cypher is nil")
				}
				rawValue, err = rs.Cypher.Decrypt(rs.encryptKey, rawValue, true)
				if err != nil {
					return nil, fmt.Errorf("decryption error: %w", err)
				}
			}
			storedData := StoreFormat{}
			err = rs.Unmarshal(rawValue, &storedData)
			if err != nil {
				return nil, err
			}
			returnValue := storedData.Value

			switch storedData.ValueType {
			case "string":
				resultMap[key] = returnValue
			default:
				resultMap[key] = returnValue
			}

		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return resultMap, nil
}

// timeout in millisecond
func (rs *RedisStore) GetRecordWithTimeOut(key string, timeout int, keyPrefixes ...string) ([]byte, error, bool) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond) // Set your desired timeout here
	defer cancel()

	resultChan := make(chan []byte)
	go func() {
		value, err, _ := rs.GetRecord(key, keyPrefixes...)
		if err != nil {
			resultChan <- []byte(fmt.Sprintf("error: %v", err))
		} else {
			resultChan <- value
		}
	}()

	select {
	case result := <-resultChan:
		if strings.HasPrefix(string(result), "error:") {
			return nil, fmt.Errorf("%s", result), false
		}
		return result, nil, true
	case <-ctx.Done():
		return nil, fmt.Errorf("command timed out for %v millisecond", timeout), false
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
