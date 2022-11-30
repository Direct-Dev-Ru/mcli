package mclisecrets

import (
	json "encoding/json"
	"time"
)

type SecretsWriter interface {
	SetContent(path string, content string) (int, error)
}
type SecretsReader interface {
	GetContent(path string) ([]byte, error)
}
type SecretsCypher interface {
	Encrypt(key, data []byte) ([]byte, error)
	Decrypt(key, data []byte) ([]byte, error)
}
type SecretsSerializer interface {
	Marshal(data any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type DefaultSerializer struct {
}

func (ds DefaultSerializer) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (ds DefaultSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
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

var DefaultSer DefaultSerializer = DefaultSerializer{}

func NewSecretsEntries(rd SecretsReader, wr SecretsWriter, cyp SecretsCypher,
	ser SecretsSerializer) SecretsEntries {

	if ser == nil {
		ser = DefaultSer
	}
	return SecretsEntries{Secrets: make([]SecretEntry, 0, 10), Wrt: wr, Rdr: rd, Srl: ser, Cypher: cyp}
}
