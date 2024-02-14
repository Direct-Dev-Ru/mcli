package mclicrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

type AesCypherType func(opts any) AesCypherType

func AesCypherDefault(opts any) AesCypherType {
	return AesCypherDefault
}

var AesCypher AesCypherType = AesCypherDefault

// GetKey generates or retrieves an encryption key based on the provided path and random flag.
//
// If random is true, it generates a new random encryption key.
// Otherwise, it retrieves the encryption key based on the provided path.
// The path can be one of the following formats:
//   - If path starts with "env:", it retrieves the key from the environment variable specified in the path.
//     Length of env value should be more than 8 symbols.
//   - Otherwise, it retrieves the key from the file specified in the path.
//
// If an error occurs during key generation or retrieval, it returns an error.
// If an error is nil in returns []byte containing 32 byte key
func (aesct AesCypherType) GetKey(path string, random bool) ([]byte, error) {
	if random {
		return GenerateKey()
	}

	if strings.HasPrefix(path, "env:") {
		envVarName := strings.Replace(path, "env:", "", 1)
		strKey := strings.TrimSpace(os.Getenv(envVarName))
		if len(strKey) < 8 {
			return nil, fmt.Errorf("value of environment variable %v is empty or less then 8 symbols", envVarName)
		}
		byteKey := SHA_256(strKey)
		return byteKey, nil
	}

	result, err := GetKeyFromFile(path)
	if err != nil {
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

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
