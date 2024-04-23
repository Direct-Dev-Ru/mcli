package mclihttp

import (
	"crypto/rand"
	"encoding/json"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_redis "mcli/packages/mcli-redis"
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

func TestCredentialType(t *testing.T) {

	user := NewCredential("user1", "pwd1", true, nil)

	pwd, err := user.GetString("Password")

	user.SetCredential("user1", "pwd2")

	_, _ = pwd, err
}

func TestUserStore(t *testing.T) {
	// Start Redis container on port 6380, publish on port 6380
	redisHostPort := "6381"
	pool, resource, err := startRedisContainer(redisHostPort)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		// Clean up Redis container
		pool.Purge(resource)
	}()

	time.Sleep(500 * time.Millisecond)

	// Initialize your RedisStore
	rs := &mcli_redis.RedisStore{
		RedisPool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", "localhost:"+redisHostPort, redis.DialPassword(""))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		},
		KeyPrefix: "userlist",
	}
	rs.SetMarshalling(json.Marshal, json.Unmarshal)
	rs.SetEncrypt(true, GenKey(32), cypher)
	us := NewUserStore(rs, "userlist")
	user := NewCredential("user_test", "pwd1", true, nil)

	// test SetUser
	err = us.SetUser(user)
	if err != nil {
		t.Errorf("error setting credential instance: %v", err)
	}
	// test GetUser
	user.Email = "test@mail.ru"
	user.Password = "test"

	// test SetUser again
	err = us.SetUser(user)
	if err != nil {
		t.Errorf("error setting credential instance: %v", err)
	}
	// test GetUser
	passwordToTest := "pwd1"
	emailToTest := "test@mail.ru"
	userFromStore, err, ok := us.GetUser(user.Username)
	if err != nil {
		t.Errorf("error getting record: %v", err)
	}
	if !ok {
		t.Error("expected record to exist, but it doesn't")
	}
	t.Log(userFromStore)

	email, err := userFromStore.GetString("Email")
	if err != nil {
		t.Errorf("error getting Email field: %v", err)
		return
	}
	if email != emailToTest {
		t.Errorf("retrieved email value don't equal to expected %s != %s", email, emailToTest)
		return
	}
	check_ok, err := us.CheckPassword(user.Username, passwordToTest)
	if !check_ok {
		t.Errorf("password check test do not pass")
		return
	}
	if err != nil {
		t.Errorf("error checking password: %v", err)
		return
	}

	passwordToTest = "new_pwd1"
	err = us.SetPassword(user.Username, passwordToTest, false)
	if err != nil {
		t.Errorf("error setting new password: %v", err)
		return
	}
	check_ok, err = us.CheckPassword(user.Username, passwordToTest)
	if !check_ok {
		t.Errorf("new password check test do not pass")
		return
	}
	if err != nil {
		t.Errorf("error checking new password: %v", err)
		return
	}

	users, err := us.GetUsers("*test*")
	if err != nil {
		t.Errorf("error get pack of users: %v", err)
		return
	}
	if _, ok := users[user.Username]; !ok {
		t.Errorf("in received users bundle there are no user %v", user.Username)
		return
	}
}
