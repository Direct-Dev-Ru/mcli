package mclicrypto

import (
	"time"
)

type SecretsWriter interface {
	SetContent(path string, content string) (int, error)
}
type SecretsReader interface {
	GetContent(path string) ([]byte, error)
}
type SecretsCypher interface {
	Encript(key, data []byte) ([]byte, error)
	Decrypt(key, data []byte) ([]byte, error)
}
type SecretsSerializer interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type SecretEntry struct {
	Name        string
	Login       string
	Secret      string
	Description string
	CreatedAt   time.Time
}

type SecretsEntries struct {
	Secrets []SecretEntry
	Wrt     SecretsWriter
	Rdr     SecretsReader
	Srl     SecretsSerializer
	Cypher  SecretsCypher
}

func NewSecretsEntries(rd SecretsReader, wr SecretsWriter, cyp SecretsCypher, ser SecretsSerializer) SecretsEntries {
	return SecretsEntries{Secrets: make([]SecretEntry, 0, 10), Wrt: wr, Rdr: rd, Srl: ser, Cypher: cyp}
}
