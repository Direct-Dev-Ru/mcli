package mcliredis

import (
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	"reflect"
	"strconv"
	"time"
)

type StoreFormat struct {
	ValueType string
	Value     []byte
	TimeStamp time.Time
}
type StoreFormatString struct {
	ValueType string
	Value     string
	TimeStamp time.Time
}

func (rs *RedisStore) getValueToStore(value interface{}) (string, error) {
	var (
		rawValue  []byte
		valueType string
		err       error
	)
	switch v := value.(type) {
	case int:
		// Handle integer value
		rawValue = []byte(strconv.Itoa(v))
		valueType = "int"
	case string:
		// Handle string value
		rawValue = []byte(v)
		valueType = "string"
	case []byte:
		rawValue = v
		valueType = "[]byte"
	default:
		// Handle other types
		rawValue, err = rs.Marshal(v)
		valueType = fmt.Sprintf("%v", reflect.ValueOf(v).Kind())
		if err != nil {
			return "", err
		}
		// fmt.Println(reflect.Struct)
		// if reflect.ValueOf(v).Kind() == reflect.Struct {
		// 	valueType = "struct"
		// }
	}
	// toStore := StoreFormatString{ValueType: valueType, Value: string(rawValue), TimeStamp: time.Now().UTC()}
	toStore := StoreFormat{ValueType: valueType, Value: rawValue, TimeStamp: time.Now().UTC()}
	// fmt.Println(toStore)
	// strBytes := []byte(valueType)

	// fullValueWithType := append(strBytes, rawValue...)

	valueToStore, err := rs.Marshal(toStore)
	if err != nil {
		return "", fmt.Errorf("preparing value to store error: %w", err)
	}

	if rs.Encrypt && len(rs.encryptKey) > 0 && rs.Cypher != nil {
		valueToStore, err = rs.Cypher.Encrypt(rs.encryptKey, valueToStore, true)
		if err != nil {
			return "", fmt.Errorf("encryption error: %w", err)
		}
		valueToStore = []byte(mcli_crypto.Base64ByteSliceEncode(valueToStore))

	}

	// strValueToStore := mcli_crypto.Base64ByteSliceEncode(valueToStore)

	return string(valueToStore), nil
}

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

	valueToStore, err := rs.getValueToStore(value)
	if err != nil {
		return err
	}

	for _, k := range resultKeys {

		_, err := rs.ExecuteCommand(conn, "SET", k, valueToStore)
		// _, err := conn.Do("SET", k, valueToStore)
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
		expiration = 999999999
	}

	conn := rs.RedisPool.Get()
	defer conn.Close()

	valueToStore, err := rs.getValueToStore(value)
	if err != nil {
		return err
	}

	for _, k := range resultKeys {
		_, err := rs.ExecuteCommand(conn, "SET", k, valueToStore, "EX", int(expiration))
		// _, err := conn.Do("SET", k, valueToStore, "EX", int(expiration))
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

		// rawValue, err := json.Marshal(val)
		// if err != nil {
		// 	return err
		// }
		// if rs.Encrypt && rs.Cypher != nil {
		// 	rawValue, err = rs.Cypher.Encrypt(rs.encryptKey, rawValue, true)
		// 	if err != nil {
		// 		return fmt.Errorf("encryption error: %w", err)
		// 	}
		// }
		// valueToStore := string(rawValue)

		valueToStore, err := rs.getValueToStore(val)
		if err != nil {
			return err
		}
		for _, k := range resultKeys {
			// _, err := conn.Do("SET", k, valueToStore)
			_, err := rs.ExecuteCommand(conn, "SET", k, valueToStore)
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

		valueToStore, err := rs.getValueToStore(val)
		if err != nil {
			return err
		}
		for _, k := range resultKeys {
			// _, err := conn.Do("SET", k, valueToStore, "EX", int(expiration))
			_, err := rs.ExecuteCommand(conn, "SET", k, valueToStore, "EX", int(expiration))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
