package mclitype

type KeyAndVaultProvider interface {
	GetKey() ([]byte, error)
	SetKey([]byte) error
	GetKeyPath() (string, error)
	SetKeyPath(string) error
	GetVault() ([]byte, error)
	SetVault([]byte) error
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
