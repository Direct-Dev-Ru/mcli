package mcliredis

import (
	"encoding/json"
	"fmt"
)

// interfaces
// type RedisStorer interface {
// GetRecord(key string, keyPrefixes ...string) (string, error, bool)
// GetRecords(pattern string, keyPrefixes ...string) (map[string]string, error)

// SetRecord(key string, value interface{}, keyPrefixes ...string) error
// SetRecords(records map[string]interface{}, keyPrefixes ...string) error
// SetRecordEx(key, value interface{}, expired int, keyPrefixes ...string) error
// SetRecordsEx(records map[string]interface{}, expired int, keyPrefixes ...string) error

// RemoveRecord(key string, keyPrefixes ...string) error
// RemoveRecords(keys []string, keyPrefixes ...string) error
// }

func (rs *RedisStore) SetRecord(key string, value interface{}, keyPrefixes ...string) error {

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

	conn := rs.RedisPool.Get()
	defer conn.Close()

	rawValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if rs.Encrypt && rs.Cypher != nil {
		rawValue, err = rs.Cypher.Encrypt(rs.encryptKey, rawValue, true)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
	}

	valueToStore := string(rawValue)
	for _, k := range resultKeys {
		_, err := conn.Do("SET", k, valueToStore)

		if err != nil {
			return err
		}
		// fmt.Println("set data:", k, string(rawValue))
	}
	return nil
}

func (rs *RedisStore) SetRecordEx(key string, value interface{}, expiration int, keyPrefixes ...string) error {
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

	if expiration == -1 || expiration == 0 {
		expiration = 99999999
	}

	conn := rs.RedisPool.Get()
	defer conn.Close()

	rawValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if rs.Encrypt && rs.Cypher != nil {
		rawValue, err = rs.Cypher.Encrypt(rs.encryptKey, rawValue, true)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
	}

	valueToStore := string(rawValue)
	for _, k := range resultKeys {
		_, err := conn.Do("SET", k, valueToStore, "EX", int(expiration))
		if err != nil {
			return err
		}
		// fmt.Println("set data:", k, string(rawValue))
	}
	return nil
}

func (rs *RedisStore) SetRecords(values map[string]interface{}, keyPrefixes ...string) error {

	conn := rs.RedisPool.Get()
	defer conn.Close()
	for key, val := range values {

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

		rawValue, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if rs.Encrypt && rs.Cypher != nil {
			rawValue, err = rs.Cypher.Encrypt(rs.encryptKey, rawValue, true)
			if err != nil {
				return fmt.Errorf("encryption error: %w", err)
			}
		}
		valueToStore := string(rawValue)
		for _, k := range resultKeys {
			_, err := conn.Do("SET", k, valueToStore)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rs *RedisStore) SetRecordsEx(values map[string]interface{}, expiration int, keyPrefixes ...string) error {

	conn := rs.RedisPool.Get()
	defer conn.Close()
	for key, val := range values {

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

		rawValue, err := json.Marshal(val)
		if err != nil {
			return err
		}
		valueToStore := string(rawValue)
		for _, k := range resultKeys {
			_, err := conn.Do("SET", k, valueToStore, "EX", int(expiration))

			if err != nil {
				return err
			}
		}
	}
	return nil
}
