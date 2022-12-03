package mclicrypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"
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
		rawString := string(rawKey)
		rawString = strings.ReplaceAll(rawString, "\r\n", "")
		rawString = strings.ReplaceAll(rawString, "\n", "")
		// fmt.Println("key:", []byte(rawString))
		theKey = SHA_256(rawString)
	} else if os.IsNotExist(err) {
		err = fmt.Errorf(fmt.Sprintf("file %s does not exists: %v", keyPath, err))
	} else {
		err = fmt.Errorf(fmt.Sprintf("file %s stat error: %v", keyPath, err))
	}

	return
}

func GetKeyFromString(s string) ([]byte, error) {
	var resultKey []byte

	if len(s) == 0 {
		return nil, fmt.Errorf("from getkeyfromstring %w", errors.New("input string is empty"))
	}
	// fmt.Println("key:", []byte(s))
	resultKey = SHA_256(s)

	if len(resultKey) == 0 {
		return nil, fmt.Errorf("from getkeyfromstring %w", errors.New("generated empty key"))
	}
	return resultKey, nil
}
