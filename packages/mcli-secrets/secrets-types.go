package mclisecrets

import (
	"encoding/hex"
	json "encoding/json"
	"fmt"
	"sync"
	"time"
)

type SecretsWriter interface {
	SetContent(path string, content []byte) (int, error)
}
type SecretsReader interface {
	GetContent(path string) ([]byte, error)
}
type SecretsCypher interface {
	Encrypt(key, data []byte, isSalted bool) ([]byte, error)
	Decrypt(key, data []byte, isSalted bool) ([]byte, error)
	GetKey(path string, random bool) ([]byte, error)
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
	sync.Mutex
	Secrets   []SecretEntry
	Wrt       SecretsWriter
	Rdr       SecretsReader
	Srl       SecretsSerializer
	Cypher    SecretsCypher
	vaultPath string
	keyPath   string
}

var DefaultSer DefaultSerializer = DefaultSerializer{}

func NewSecretsEntries(rd SecretsReader, wr SecretsWriter, cyp SecretsCypher,
	ser SecretsSerializer) SecretsEntries {

	if ser == nil {
		ser = DefaultSer
	}
	return SecretsEntries{Secrets: make([]SecretEntry, 0, 10), Wrt: wr, Rdr: rd, Srl: ser, Cypher: cyp}
}

func (ses *SecretsEntries) AddEntry(se SecretEntry) error {
	ses.Lock()
	defer ses.Unlock()
	ses.Secrets = append(ses.Secrets, se)
	return nil
}

func (ses *SecretsEntries) FillStore(vaultPath, keyPath string) error {
	storeContent, err := ses.Rdr.GetContent(vaultPath)
	if err != nil {
		return fmt.Errorf("read store fault: %w", err)
	}

	if len(storeContent) > 0 {
		key, err := ses.Cypher.GetKey(keyPath, false)
		if err != nil {
			// if _, err = os.Stat(keyPath); !errors.Is(err, os.ErrNotExist) {
			// 	key, err = ses.Cypher.GetKey(keyPath, true)
			// 	fmt.Println(key)
			// } else {
			return fmt.Errorf("%w", err)
			// }
		}
		cypherData, err := hex.DecodeString(string(storeContent))
		if err != nil {
			return fmt.Errorf("hex decrypting fault: %w", err)
		}

		storeContent, err = ses.Cypher.Decrypt(key, cypherData, true)
		if err != nil {
			return fmt.Errorf("decrypting fault: %v", err)
		}
		ses.Srl.Unmarshal(storeContent, &ses.Secrets)
	}
	ses.vaultPath = vaultPath
	ses.keyPath = keyPath
	return nil
}

func (ses *SecretsEntries) Save(vaultPath, keyPath string) error {
	ses.Lock()
	defer ses.Unlock()
	if len(vaultPath) == 0 {
		vaultPath = ses.vaultPath
	}
	if len(keyPath) == 0 {
		keyPath = ses.keyPath
	}

	raw, err := ses.Srl.Marshal(ses.Secrets)
	if err != nil {
		return err
	}

	key, err := ses.Cypher.GetKey(keyPath, false)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	cypherData, err := ses.Cypher.Encrypt(key, raw, true)
	if err != nil {
		return fmt.Errorf("encrypt fault: %w", err)
	}
	hexCypherData := hex.EncodeToString(cypherData)

	_, err = ses.Wrt.SetContent(vaultPath, []byte(hexCypherData))
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// GetEncContent - gets encrypted content and return byte slice
// key is input parameter - byte slice
func (ses *SecretsEntries) GetEncContent(key []byte) ([]byte, error) {
	ses.Lock()
	defer ses.Unlock()
	if len(key) == 0 {
		return nil, fmt.Errorf("zero key length")
	}

	raw, err := ses.Srl.Marshal(ses.Secrets)
	if err != nil {
		return nil, err
	}

	cypherData, err := ses.Cypher.Encrypt(key, raw, true)
	if err != nil {
		return nil, fmt.Errorf("encrypt fault: %w", err)
	}
	hexCypherData := make([]byte, hex.EncodedLen(len(cypherData)))

	hex.Encode(hexCypherData, cypherData)

	return hexCypherData, nil
}

// SetFromEncContent - sets secrets from given encoded content and key and returns error if occured
func (ses *SecretsEntries) GetFromEncContent(content []byte, key []byte) error {
	ses.Lock()
	defer ses.Unlock()
	if len(content) == 0 {
		return fmt.Errorf("from GetFromEncContent: zero content length")
	}
	if len(key) == 0 {
		return fmt.Errorf("from GetFromEncContent: zero key length")
	}

	cypherData := make([]byte, hex.DecodedLen(len(content)))
	_, err := hex.Decode(cypherData, content)
	if err != nil {
		return fmt.Errorf("from GetFromEncContent: %w", err)
	}

	storeContent, err := ses.Cypher.Decrypt(key, cypherData, true)
	if err != nil {
		return fmt.Errorf("decrypting fault: %v", err)
	}
	ses.Srl.Unmarshal(storeContent, &ses.Secrets)
}
