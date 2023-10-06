package mcliinterface

type KeyAndVaultProvider interface {
	GetKey() ([]byte, error)
	GetKeyPath() (string, error)
	GetVault() (interface{}, error)
	SetVault() (interface{}, error)
	GetVaultPath() (string, error)
}

type SecretsCypher interface {
	Encrypt(key, data []byte, isSalted bool) ([]byte, error)
	Decrypt(key, data []byte, isSalted bool) ([]byte, error)
	GetKey(path string, random bool) ([]byte, error)
}
type SecretsWriter interface {
	SetContent(path string, content []byte) (int, error)
}
type SecretsReader interface {
	GetContent(path string) ([]byte, error)
}

type SecretsSerializer interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}
