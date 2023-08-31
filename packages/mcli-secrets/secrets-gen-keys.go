package mclisecrets

import (
	"crypto/rand"
	"fmt"
	"os"
)

func GenKey(length int) []byte {
	if length == 0 {
		length = 32
	}
	k := make([]byte, length)
	if _, err := rand.Read(k); err != nil {
		return nil
	}
	return k
}

func SaveKeyToFilePlain(filePath string, key []byte) error {
	if len(key) == 0 {
		key = GenKey(32)
	}
	if key == nil {
		return fmt.Errorf("failed to generate key")
	}

	// Convert byte slice to hexadecimal string representation
	hexKey := fmt.Sprintf("%x", key)

	// Write hexadecimal string to file
	err := os.WriteFile(filePath, []byte(hexKey), 0600)
	if err != nil {
		return fmt.Errorf("error writing key to file: %v", err)
	}
	return nil
}

func LoadKeyFromFilePlain(filePath string) ([]byte, error) {
	// Read hexadecimal string from file
	readHexKey, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading key from file: %v", err)
	}

	// Convert hexadecimal string back to byte slice
	var readKey []byte
	_, err = fmt.Sscanf(string(readHexKey), "%x", &readKey)
	if err != nil {
		return nil, fmt.Errorf("error converting hexadecimal string: %v", err)
	}
	return readKey, nil
}

func LoadKeyFromHexString(hexKeyString string) (string, error) {

	// Convert hexadecimal string back to byte slice
	var readKey []byte
	_, err := fmt.Sscanf(string(hexKeyString), "%x", &readKey)
	if err != nil {
		return "", fmt.Errorf("error converting hexadecimal string: %v", err)
	}
	return string(readKey), nil
}
