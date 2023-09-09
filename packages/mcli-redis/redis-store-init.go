package mcliredis

import (
	"fmt"
	"time"

	mcli_secrets "mcli/packages/mcli-secrets"

	"github.com/gomodule/redigo/redis"
)

// vars
var RedisPool *redis.Pool

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
	RedisPool  *redis.Pool
	KeyPrefix  string
	Encrypt    bool
	Cypher     mcli_secrets.SecretsCypher
	encryptKey []byte
}

func NewRedisStore(host, password, keyPrefix string) (*RedisStore, error) {
	if RedisPool == nil || len(host) > 0 || len(password) > 0 {
		_, err := InitCache(host, password)
		if err != nil {
			return nil, err
		}
	}
	return &RedisStore{RedisPool: RedisPool, KeyPrefix: keyPrefix, Encrypt: false}, nil
}

func InitCache(host, password string) (*redis.Pool, error) {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.Dial("tcp", host, redis.DialPassword(password))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Ping the Redis server
	_, err = redis.String(conn.Do("PING"))
	if err != nil {
		return nil, err
	}

	RedisPool = NewRedisPool(host, password)
	return RedisPool, nil
}

func NewRedisPool(host, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host, redis.DialPassword(password))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func (r *RedisStore) SetEcrypt(encrypt bool, encryptKey []byte, cypher mcli_secrets.SecretsCypher) {
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
