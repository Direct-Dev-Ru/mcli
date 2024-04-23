package mcliredis

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_type "mcli/packages/mcli-type"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var cypher mcli_type.SecretsCypher = mcli_crypto.AesCypher

func GenKey(length int) []byte {
	if length == 0 {
		length = 32
	}
	k := make([]byte, length)
	if _, err := rand.Read(k); err != nil {
		return nil
	}
	return k
}

func startRedisContainer(redisHostPort string) (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, err
	}
	// resource, err := pool.Run("redis", "7", nil)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "7",
		Env: []string{
			"SOMEVAR1=test1",
			"SOME_PASSWORD=test2",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {

		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		// Set up port mappings
		config.PortBindings = map[docker.Port][]docker.PortBinding{
			"6379/tcp": {{HostIP: "", HostPort: redisHostPort}},
		}
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		return nil, nil, err
	}
	return pool, resource, nil
}

func TestRedisStore(t *testing.T) {
	// Start Redis container on port 6380, publish on port 6380
	pool, resource, err := startRedisContainer("6380")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		// Clean up Redis container
		pool.Purge(resource)
	}()

	time.Sleep(1 * time.Second)

	// Initialize RedisStore
	rs := &RedisStore{
		RedisPool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", "localhost:6380", redis.DialPassword(""), redis.DialDatabase(10))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		},
		KeyPrefix: "prefix",
	}
	rs.SetMarshalling(json.Marshal, json.Unmarshal)
	rs.SetEncrypt(true, GenKey(32), cypher)

	// Test SetRecord
	valueToTest := "testValue"
	err = rs.SetRecord("testKey", valueToTest, "testPrefix")
	if err != nil {
		t.Errorf("error setting record: %v", err)
	}

	// Test GetRecord
	result, err, ok := rs.GetRecord("testKey", "testPrefix")
	fmt.Println(string(result), err, ok)

	if !ok {
		t.Error("expected record to exist, but it doesn't")
	}
	if string(result) != valueToTest {
		t.Errorf("retrieved value don't equal to expected %s", valueToTest)
	}

}
