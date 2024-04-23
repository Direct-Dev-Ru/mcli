package mclitype

type KVStorer interface {
	SetEncrypt(encrypt bool, key []byte, cypher SecretsCypher)
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

type KVStorerV2 interface {
	SetEncryptV2(encrypt bool, key []byte, cypher SecretsCypher)
	SetMarshallingV2(fMarshal func(any) ([]byte, error), fUnMarshal func([]byte, any) error)
	GetMarshalV2() func(any) ([]byte, error)
	GetUnMarshalV2() func([]byte, any) error

	GetRecordV2(key string, options KVOptioner) ([]byte, error, bool)
	// GetRecordExV2(key string, keyPrefixes ...string) ([]byte, int, error)
	GetRecordsV2(pattern string, options KVOptioner) (map[string][]byte, error)

	SetRecordV2(key string, value interface{}, options KVOptioner) error
	SetRecordsV2(records map[string]interface{}, options KVOptioner) error

	// SetRecordExV2(key string, value interface{}, expiration int, keyPrefixes ...string) error
	// SetRecordsExV2(records map[string]interface{}, expired int, keyPrefixes ...string) error

	RemoveRecordV2(key string, keyPrefixes ...string) error
	RemoveRecordsV2(keys []string, keyPrefixes ...string) error

	CloseV2()
}

type KVSchemer interface {
	// [seq] or [field:{fieldname}] or [guid]
	GetSchemePrimaryKeyType() string
	// get fields to make lookup index
	GetSchemeIndexFields() []string
	// plain or hash
	GetSchemeRecordType() string
}

type KVSchemerV2 interface {
	GetScheme() *Scheme
}

type KVOptioner interface {

	// get map[string] of options
	GetOptionMap(string) (interface{}, bool)

	// set to map[string] of options
	SetOptionMap(string, interface{}) interface{}

	GetStringOption(optionName string) (string, bool)
	GetIntOption(optionName string) (int, bool)
	GetBoolOption(optionName string) (bool, bool)
}
