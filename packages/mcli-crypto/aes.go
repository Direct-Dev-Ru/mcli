package mclicrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

type AesCypherType func(opts any) AesCypherType

func AesCypherDefault(opts any) AesCypherType {
	return AesCypherDefault
}

var AesCypher AesCypherType = AesCypherDefault

// Get sec key from file. if error fom file anf random set to true - returns random key
func (aesct AesCypherType) GetKey(path string, random bool) ([]byte, error) {

	result, err := GetKeyFromFile(path)
	if err != nil {
		if random {
			return GenerateKey()
		}
		return nil, fmt.Errorf("get key: %w", err)
	}
	return result, nil
}

func (aesct AesCypherType) DeriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func (aesct AesCypherType) Encrypt(key, data []byte, isSalted bool) ([]byte, error) {
	var err error
	var salt []byte

	if isSalted {
		key, salt, err = aesct.DeriveKey(key, nil)
		if err != nil {
			return nil, err
		}
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	if isSalted {
		ciphertext = append(ciphertext, salt...)
	}
	return ciphertext, nil
}

func (aesct AesCypherType) Decrypt(key, data []byte, isSalted bool) ([]byte, error) {
	var err error
	var salt []byte
	if isSalted {
		salt, data = data[len(data)-32:], data[:len(data)-32]

		key, _, err = aesct.DeriveKey(key, salt)
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
