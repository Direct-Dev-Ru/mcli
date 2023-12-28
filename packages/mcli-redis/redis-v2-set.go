package mcliredis

import (
	"fmt"
	mcli_type "mcli/packages/mcli-type"
	mcli_utils "mcli/packages/mcli-utils"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

func (rs *RedisStore) GetValueToStoreV2(value map[string]interface{}) (string, error) {
	var (
		rawValue  []byte
		valueType string
		err       error
	)

	rawValue, err = rs.Marshal(value)
	valueType = "map_of_interfaces"
	if err != nil {
		return "", err
	}

	toStore := StoreFormat{ValueType: valueType, Value: rawValue, TimeStamp: time.Now().UTC()}
	valueToStore, err := rs.Marshal(toStore)
	if err != nil {
		return "", fmt.Errorf("preparing value to store error: %w", err)
	}
	if rs.Encrypt && rs.Cypher != nil {
		valueToStore, err = rs.Cypher.Encrypt(rs.encryptKey, valueToStore, true)
		if err != nil {
			return "", fmt.Errorf("encryption error: %w", err)
		}
	}
	return string(valueToStore), nil
}

func (rs *RedisStore) SetRecordV2(key string, value interface{}, options mcli_type.KVOptioner) error {

	if options == nil {
		return rs.SetRecord(key, value)
	}

	prefix, ok := options.GetStringOption("prefix")
	if !ok {
		prefix = rs.KeyPrefix
	}

	scheme, ok := options.GetOptionMap("scheme")
	if !ok {
		// just simple store V1
		return rs.SetRecord(key, value, prefix)
	}

	// convert to KVSchemer
	kvScheme, ok := scheme.(mcli_type.KVSchemerV2)
	if !ok {
		// just simple store V1
		return rs.SetRecord(key, value, prefix)
	}

	storeScheme := kvScheme.GetScheme()

	if storeScheme.Prefix != "" {
		prefix = storeScheme.Prefix
	}

	conn := rs.RedisPool.Get()
	defer conn.Close()

	valueAsMap, err := mcli_utils.StructToMap(value)
	if err != nil {
		return fmt.Errorf("convert value as struct to map failed: %v", err)
	}

	// if key, passed to func is empty
	if len(key) == 0 {
		keyType := storeScheme.PKType
		switch {
		case keyType == mcli_type.PKTypeSequence:
			_, err := conn.Do("INCR", prefix+":RECORD_NUM")
			if err != nil {
				return err
			}
			incValueFromRedis, err := redis.Int(conn.Do("GET", prefix+":RECORD_NUM"))
			if err != nil {
				return err
			}
			key = strconv.Itoa(incValueFromRedis)

		case keyType == mcli_type.PKTypeFieldValue:
			keyFieldName := storeScheme.PKFieldName
			if keyFieldValue, ok := valueAsMap[keyFieldName]; ok {
				switch v := keyFieldValue.(type) {
				case string:
					key = v
				case int:
					key = strconv.Itoa(v)
				default:
					key = uuid.New().String()
				}
			}
		default:
			uuid := uuid.New()
			key = uuid.String()
		}
	}

	// Start a redis transaction
	conn.Send("MULTI")

	recType := storeScheme.RecordType

	overallKey := fmt.Sprintf("%s:%s", prefix, key)

	if recType == mcli_type.RecordTypePlain {
		// get value to store as plain raw data
		valueToStore, err := rs.getValueToStore(value)
		if err != nil {
			return err
		}
		//  save plain raw data
		err = conn.Send("SET", overallKey, valueToStore)
		if err != nil {
			conn.Do("DISCARD")
			return err
		}

		// store indexes
		for _, iField := range storeScheme.Indexes {
			indexName := iField.IndexName
			indexKey := ""
			indexNameFromFields := ""
			for i, indexField := range iField.Fields {
				indexNameFromFields += indexField
				if ivalue, ok := valueAsMap[indexField].(string); ok {
					indexKey += ivalue
				}
				if ivalue, ok := valueAsMap[indexField].(int); ok {
					indexKey += strconv.Itoa(ivalue)
				}
				if i < len(iField.Fields)-1 {
					indexNameFromFields += "^$:$^"
					indexKey += "^$:$^"
				}
			}
			if indexName == "" {
				indexName = indexNameFromFields
			}

			hSetOverallKey := fmt.Sprintf("%s:%s:%s", prefix, "lookup", indexName)
			err = conn.Send("HSET", hSetOverallKey, indexKey, overallKey)
			if err != nil {
				conn.Do("DISCARD")
				return err
			}
		}
	}
	// nested structs may be saved uncorrectly
	if recType == mcli_type.RecordTypeHashTable {
		for k, v := range valueAsMap {
			err = conn.Send("HSET", overallKey, k, v)
			if err != nil {
				conn.Do("DISCARD")
				return err
			}
		}
	}

	// Execute the transaction
	_, err = conn.Do("EXEC")
	if err != nil {
		conn.Do("DISCARD") // Discard the transaction if there's an error
		return err
	}
	return nil
}
