package mclihttp

import (
	"net/http"
)

type Middleware interface {
	http.Handler
	SetInnerHandler(http.Handler)
}

type CredentialStorer interface {
	GetUser(username string) (*Credential, error, bool)
	GetAllUsers(pattern string) ([]*Credential, error)
	SetPassword(username, password string) error
	CheckPassword(username, password string) (bool, error)
}

// key-value storer interface
type KVStorer interface {
	GetRecord(key string, keyPrefixes ...string) (string, error, bool)
	GetRecordEx(key string, keyPrefixes ...string) (string, int, error)
	GetRecords(pattern string, keyPrefixes ...string) (map[string]string, error)

	SetRecord(key string, value interface{}, keyPrefixes ...string) error
	SetRecords(records map[string]interface{}, keyPrefixes ...string) error
	SetRecordEx(key string, value interface{}, expiration int, keyPrefixes ...string) error
	SetRecordsEx(records map[string]interface{}, expired int, keyPrefixes ...string) error

	RemoveRecord(key string, keyPrefixes ...string) error
	RemoveRecords(keys []string, keyPrefixes ...string) error
}
