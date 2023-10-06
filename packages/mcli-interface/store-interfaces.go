package mcliinterface

type KVStorer interface {
	SetEcrypt(encrypt bool, key []byte, cypher SecretsCypher)
	SetMarshalling(fMarshal func(any) ([]byte, error), fUnMarshal func([]byte, any) error)
	GetMarshal() func(any) ([]byte, error)
	GetUnMarshal() func([]byte, any) error

	GetRecord(key string, keyPrefixes ...string) ([]byte, error, bool)
	GetRecordEx(key string, keyPrefixes ...string) ([]byte, int, error)
	GetRecords(pattern string, keyPrefixes ...string) (map[string][]byte, error)

	SetRecord(key string, value interface{}, keyPrefixes ...string) error
	SetRecords(records map[string]interface{}, keyPrefixes ...string) error
	SetRecordEx(key string, value interface{}, expiration int, keyPrefixes ...string) error
	SetRecordsEx(records map[string]interface{}, expired int, keyPrefixes ...string) error

	RemoveRecord(key string, keyPrefixes ...string) error
	RemoveRecords(keys []string, keyPrefixes ...string) error

	Close()
}
