package mclistore

// import (
// 	mcli_secrets "mcli/packages/mcli-secrets"
// )

// type KVStorer interface {
// 	SetEncrypt(encrypt bool, key []byte, cypher mcli_secrets.SecretsCypher)

// 	GetRecord(key string, keyPrefixes ...string) (string, error, bool)
// 	GetRecordEx(key string, keyPrefixes ...string) (string, int, error)
// 	GetRecords(pattern string, keyPrefixes ...string) (map[string]string, error)

// 	SetRecord(key string, value interface{}, keyPrefixes ...string) error
// 	SetRecords(records map[string]interface{}, keyPrefixes ...string) error
// 	SetRecordEx(key string, value interface{}, expiration int, keyPrefixes ...string) error
// 	SetRecordsEx(records map[string]interface{}, expired int, keyPrefixes ...string) error

// 	RemoveRecord(key string, keyPrefixes ...string) error
// 	RemoveRecords(keys []string, keyPrefixes ...string) error

// 	Close()
// }
