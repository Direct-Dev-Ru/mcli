package mcliredis

import (
	"encoding/json"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_type "mcli/packages/mcli-type"
	mcli_utils "mcli/packages/mcli-utils"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var cypherV2 mcli_type.SecretsCypher = mcli_crypto.AesCypher

func StartRedisContainerV2(redisHostPort string, fake bool) (*dockertest.Pool, *dockertest.Resource, error) {
	if fake {
		return nil, nil, nil
	}
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

type TestUserStruct struct {
	Id          int
	Name        string
	Email       string
	Age         int
	Description string
	Token       string
}

type TestUserStructScheme struct {
	scheme *mcli_type.Scheme
}

func (us TestUserStructScheme) GetScheme() *mcli_type.Scheme {

	us.scheme = mcli_type.NewScheme(mcli_type.StoreTypeRedis, "1")
	us.scheme.PKType = mcli_type.PKTypeSequence

	us.scheme.SetEncryptedFields([]string{"Token"})

	indexField1 := mcli_type.SchemeIndex{IndexName: "Email", Fields: []string{"Email"}, NotUnique: false}
	us.scheme.Indexes = []mcli_type.SchemeIndex{indexField1}

	us.scheme.Prefix = "users"

	// us.scheme.RecordType = mcli_type.RecordTypePlain
	us.scheme.RecordType = mcli_type.RecordTypeHashTable

	return us.scheme
}

func TestRedisStoreV2(t *testing.T) {
	// Start Redis container on port 6380, publish on port 6380
	var waitAfterStartSeconds int64 = 2
	startNewContainer := true
	killContainerAfter := false
	pool, resource, err := StartRedisContainerV2("6380", !startNewContainer)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		// Clean up Redis container
		if killContainerAfter {
			pool.Purge(resource)
		}
	}()

	time.Sleep(time.Duration(waitAfterStartSeconds) * time.Second)

	// Initialize RedisStore
	rs := &RedisStore{
		RedisPool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", "localhost:6380", redis.DialPassword(""), redis.DialDatabase(5))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		},
		KeyPrefix: "testns",
	}
	rs.SetMarshalling(json.Marshal, json.Unmarshal)
	rs.SetEcrypt(false, GenKey(32), cypherV2)

	// Test SetRecord
	valueToTest1 := TestUserStruct{Id: 1, Name: "testuser1", Token: "testpassword1",
		Email: "test1@tesdomain.com", Age: 45, Description: "this is a test user 1"}

	valueToTest2 := TestUserStruct{Id: 2, Name: "testuser2", Token: "testpassword2",
		Email: "test2@tesdomain.com", Age: 45, Description: "this is a test user 2"}
	// setup options
	opt := &mcli_utils.CommonOption{}
	opt.SetOptionMap("scheme", TestUserStructScheme{})

	err = rs.SetRecordV2("", valueToTest1, opt)
	if err != nil {
		t.Errorf("error setting record 1: %v", err)
	}

	err = rs.SetRecordV2("", valueToTest2, opt)
	if err != nil {
		t.Errorf("error setting record 2: %v", err)
	}

	// Test GetRecord

	// result, err, ok := rs.GetRecord("testKey", "testPrefix")
	// fmt.Println(string(result), err, ok)

	// if !ok {
	// 	t.Error("expected record to exist, but it doesn't")
	// }
	// if string(result) != valueToTest {
	// 	t.Errorf("retrieved value don't equal to expected %s", valueToTest)
	// }

}
