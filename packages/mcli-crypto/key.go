package mclicrypto

import (
	"crypto/rand"
	"fmt"
	"os"
)

func GenerateKey() (key []byte, err error) {
	key = make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		return nil, err
	}
	return
}

func GetKeyFromFile(path string) (theKey []byte, err error) {
	keyPath := path
	if path == "" {
		keyPath = os.Getenv("HOME") + "/.ssh/id_rsa"
	}
	var rawKey []byte
	_, err = os.Stat(keyPath)
	if err == nil {
		rawKey, err = os.ReadFile(keyPath)

		theKey = SHA_256(string(rawKey))
	} else if os.IsNotExist(err) {
		err = fmt.Errorf(fmt.Sprintf("file %s does not exists: %v", keyPath, err))
	} else {
		err = fmt.Errorf(fmt.Sprintf("file %s stat error: %v", keyPath, err))
	}

	return
}
