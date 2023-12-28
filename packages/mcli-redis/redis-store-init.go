package mcliredis

import (
	"encoding/json"
	"fmt"
	"time"

	mcli_type "mcli/packages/mcli-type"

	"github.com/gomodule/redigo/redis"
)

// vars
// var RedisPool *redis.Pool
var MapRedisPool map[string]*redis.Pool = make(map[string]*redis.Pool)

// types
type KeyNotFoundError struct {
	Key string
}

func NewKeyNotFoundError(key string) *KeyNotFoundError {
	e := &KeyNotFoundError{Key: key}
	return e
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key '%s' not found", e.Key)
}

type RedisStore struct {
	RedisPool       *redis.Pool
	RedisDatabaseNo int
	KeyPrefix       string
	Encrypt         bool
	Cypher          mcli_type.SecretsCypher
	Marshal         func(any) ([]byte, error)
	Unmarshal       func([]byte, any) error
	encryptKey      []byte
}

func NewRedisStore(poolname, host, password, keyPrefix string, databaseNo int) (*RedisStore, error) {
	if poolname == "" {
		poolname = "default"
	}
	if host == "" || host == ":" {
		host = "localhost:6379"
	}
	redisPool := MapRedisPool[poolname]

	var err error
	if redisPool == nil && len(host) > 0 && len(password) > 0 {
		redisPool, err = InitCache(host, password, databaseNo)
		if err != nil {
			return nil, err
		}
		MapRedisPool[poolname] = redisPool
	}
	return &RedisStore{RedisPool: redisPool, KeyPrefix: keyPrefix, RedisDatabaseNo: databaseNo,
		Encrypt: false, Marshal: json.Marshal, Unmarshal: json.Unmarshal}, nil
}

func InitCache(host, password string, databaseNo int) (*redis.Pool, error) {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.Dial("tcp", host, redis.DialPassword(password), redis.DialDatabase(databaseNo))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Ping the Redis server
	// Select the desired database (e.g., database 1)
	_, err = conn.Do("SELECT", databaseNo)
	if err != nil {
		return nil, fmt.Errorf("error selecting database: %w", err)
	}
	_, err = redis.String(conn.Do("PING"))
	if err != nil {
		return nil, fmt.Errorf("error ping redis: %w", err)
	}
	redisPool := NewRedisPool(host, password, databaseNo)
	return redisPool, nil
}

func NewRedisPool(host, password string, databaseNo int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host, redis.DialPassword(password), redis.DialDatabase(databaseNo))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func (r *RedisStore) SetMarshalling(fMarshal func(any) ([]byte, error), fUnMarshal func([]byte, any) error) {
	if fMarshal != nil {
		r.Marshal = fMarshal
	}
	if fUnMarshal != nil {
		r.Unmarshal = fUnMarshal
	}
}

func (r *RedisStore) GetMarshal() func(any) ([]byte, error) {
	return r.Marshal
}

func (r *RedisStore) GetUnMarshal() func([]byte, any) error {
	return r.Unmarshal
}

func (r *RedisStore) SetEcrypt(encrypt bool, encryptKey []byte, cypher mcli_type.SecretsCypher) {
	r.Encrypt = encrypt
	r.Cypher = cypher
	r.encryptKey = encryptKey
	if len(r.encryptKey) == 0 {
		r.Encrypt = false
	}
	if !encrypt {
		r.Cypher = nil
	}
}

func (r *RedisStore) Close() {
	if r.RedisPool != nil {
		r.RedisPool.Close()
	}
}
