package mclisecrets

import (
	"encoding/hex"
	json "encoding/json"
	"fmt"
	"sync"
	"time"

	mcli_type "mcli/packages/mcli-type"
)

type DefaultSerializer struct {
}

func (ds DefaultSerializer) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (ds DefaultSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type SecretEntry struct {
	SecretPlainEntry
	store *SecretsEntries
}
type SecretPlainEntry struct {
	Name        string
	Login       string
	Secret      string
	Description string
	ExtraData   map[string]string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

func (se *SecretEntry) GetPlainEntry() SecretPlainEntry {
	return SecretPlainEntry{Name: se.Name, Login: se.Login, Secret: se.Secret,
		Description: se.Description, CreatedAt: se.CreatedAt, UpdatedAt: se.UpdatedAt, ExtraData: se.ExtraData}
}

func (se *SecretEntry) encodeSecret(phrase string, keyPath string, isSalted bool) (string, error) {
	key, err := se.store.Cypher.GetKey(keyPath, false)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	encData, err := se.store.Cypher.Encrypt(key, []byte(phrase), isSalted)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	encodedString := hex.EncodeToString(encData)
	return encodedString, nil
}

func (se *SecretEntry) decodeSecret(encContent string, keyPath string, isSalted bool) (string, error) {
	key, err := se.store.Cypher.GetKey(keyPath, false)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	cypherData, err := hex.DecodeString(encContent)
	if err != nil {
		return "", fmt.Errorf("hex decrypting fault: %w", err)
	}

	decryptingContent, err := se.store.Cypher.Decrypt(key, cypherData, true)
	if err != nil {
		return "", fmt.Errorf("decrypting fault: %v", err)
	}
	return string(decryptingContent), err
}

func (se *SecretEntry) SetSecret(phrase string, isSalted, encode bool) (string, error) {
	if encode {
		encodedString, err := se.encodeSecret(phrase, se.store.keyPath, isSalted)
		if err == nil {
			se.Secret = "enc:" + encodedString
		}
		return se.Secret, err
	}
	se.Secret = phrase
	return phrase, nil
}

func (se *SecretEntry) GetSecret(keyPath string, isSalted bool) (string, error) {
	secretData := se.Secret
	if len(secretData) > 5 && secretData[:4] == "enc:" {
		if len(keyPath) == 0 && se.store != nil {
			keyPath = se.store.keyPath
		}
		decodedString, err := se.decodeSecret(secretData[4:], keyPath, isSalted)
		return decodedString, err
	}
	return secretData, nil
}

func (se *SecretEntry) Update(seSource SecretEntry) error {
	if !(len(seSource.Secret) > 0 && seSource.Name == se.Name) {
		return fmt.Errorf("secretEntry update: empty secret or different secret names")
	}
	se.Login = seSource.Login
	se.Secret = seSource.Secret
	se.Description = seSource.Description
	se.CreatedAt = seSource.CreatedAt
	return nil
}

// SecretEntries - struct for storing array of secrets and maintain some operations on them
type SecretsEntries struct {
	sync.Mutex
	Secrets   []SecretEntry
	SecretMap map[string]*SecretEntry
	Wrt       mcli_type.SecretsWriter
	Rdr       mcli_type.SecretsReader
	Srl       mcli_type.SecretsSerializer
	Cypher    mcli_type.SecretsCypher
	vaultPath string
	keyPath   string
	knvp      mcli_type.KeyAndVaultProvider
}

var DefaultSer DefaultSerializer = DefaultSerializer{}

func NewSecretsEntriesV2(rd mcli_type.SecretsReader, wr mcli_type.SecretsWriter, cyp mcli_type.SecretsCypher,
	ser mcli_type.SecretsSerializer, knvp mcli_type.KeyAndVaultProvider) *SecretsEntries {

	ses := NewSecretsEntries(rd, wr, cyp, ser)
	if knvp == nil {
		knvp, _ = NewDefaultKeyAndVaultProvider("secrets.vault", "default.key")
	}
	ses.knvp = knvp

	return &ses
}

func NewSecretsEntries(rd mcli_type.SecretsReader, wr mcli_type.SecretsWriter, cyp mcli_type.SecretsCypher,
	ser mcli_type.SecretsSerializer) SecretsEntries {

	if ser == nil {
		ser = DefaultSer
	}
	return SecretsEntries{Secrets: make([]SecretEntry, 0, 10),
		SecretMap: make(map[string]*SecretEntry, 0), Wrt: wr, Rdr: rd, Srl: ser, Cypher: cyp}
}

func (ses *SecretsEntries) NewEntry(name, login, descr string) (SecretEntry, error) {
	ses.Lock()
	defer ses.Unlock()
	if len(name) == 0 {
		return SecretEntry{}, fmt.Errorf("add secret entry: name is empty")
	}
	spe := SecretPlainEntry{Name: name, Description: descr, ExtraData: make(map[string]string, 0),
		Login: login, Secret: "", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	secretEntry := SecretEntry{SecretPlainEntry: spe, store: ses}
	return secretEntry, nil
}

func (ses *SecretsEntries) AddEntry(se SecretEntry) error {
	ses.Lock()
	defer ses.Unlock()
	var update bool = false
	for i := 0; i < len(ses.Secrets); i++ {
		if ses.Secrets[i].Name == se.Name {
			if err := ses.Secrets[i].Update(se); err != nil {
				return err
			}
			return nil
		}
	}
	if !update {
		se.store = ses
		ses.Secrets = append(ses.Secrets, se)
	}

	se.store = ses
	ses.SecretMap[se.Name] = &se

	return nil
}

func (ses *SecretsEntries) FillStoreV2() error {
	knvp := ses.knvp
	vaultContent, err := knvp.GetVault()
	if err != nil {
		return fmt.Errorf("reading vault store fault: %w", err)
	}
	if len(vaultContent) > 0 {
		key, err := knvp.GetKey()
		if err != nil {
			return fmt.Errorf("geting key error %w", err)
		}
		cypherData, err := hex.DecodeString(string(vaultContent))
		if err != nil {
			return fmt.Errorf("hex vault content decrypting fault: %w", err)
		}
		vaultContent, err = ses.Cypher.Decrypt(key, cypherData, true)
		if err != nil {
			return fmt.Errorf("decrypting vault content fault: %v", err)
		}
		ses.Srl.Unmarshal(vaultContent, &ses.Secrets)
	}
	for i := 0; i < len(ses.Secrets); i++ {
		ses.Secrets[i].store = ses
		ses.SecretMap[ses.Secrets[i].Name] = &ses.Secrets[i]
	}
	ses.vaultPath, _ = knvp.GetVaultPath()
	ses.keyPath, _ = knvp.GetKeyPath()
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
			return fmt.Errorf("%w", err)
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
	ses.SecretMap = make(map[string]*SecretEntry)
	for i := 0; i < len(ses.Secrets); i++ {
		ses.Secrets[i].store = ses
		ses.SecretMap[ses.Secrets[i].Name] = &ses.Secrets[i]
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

func (ses *SecretsEntries) SaveV2() error {
	knvp := ses.knvp
	ses.Lock()
	defer ses.Unlock()
	vaultPath, _ := knvp.GetVaultPath()
	if len(vaultPath) == 0 {
		vaultPath = ses.vaultPath
	}

	raw, err := ses.Srl.Marshal(ses.Secrets)
	if err != nil {
		return err
	}

	key, err := knvp.GetKey()
	if err != nil {
		return fmt.Errorf("get key error: %w", err)
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
	return ses.Srl.Unmarshal(storeContent, &ses.Secrets)

}

func (ses *SecretsEntries) GetSecretPlainMap() map[string]SecretPlainEntry {
	ses.Lock()
	defer ses.Unlock()
	smap := make(map[string]SecretPlainEntry, len(ses.Secrets))
	for _, se := range ses.Secrets {
		// fmt.Print(se.Name)
		smap[se.Name] = se.GetPlainEntry()
	}
	return smap
}
