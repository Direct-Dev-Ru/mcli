package mclicrypto

import (
	"bytes"
	"os"
	"testing"
)

func TestGetKey(t *testing.T) {
	aesct := AesCypher // Create an instance of AesCypherType

	// Test case for random key generation
	randomKey, err := aesct.GetKey("", true)
	if err != nil {
		t.Errorf("Random key generation failed: %v", err)
	}
	if len(randomKey) != 32 { // Assuming GenerateKey() returns a key of length 32
		t.Errorf("Random key length is not 32 bytes: %d", len(randomKey))
	}

	// Test case for retrieving key from environment variable
	os.Setenv("TEST_KEY", "test1234567890") // Set environment variable
	envKey, err := aesct.GetKey("env:TEST_KEY", false)
	if err != nil {
		t.Errorf("Failed to retrieve key from environment variable: %v", err)
	}
	// t.Log(envKey)
	expectedKey := []byte{122, 254, 203, 8, 251, 107, 123, 171, 75, 180, 92, 117, 93, 133, 122, 231, 29, 144, 237, 43, 55, 137, 159, 41, 184, 187, 198, 174, 117, 138, 81, 117}
	if !bytes.Equal(envKey, expectedKey) {
		t.Errorf("Retrieved key from environment variable does not match expected key")
	}

	// You can add more test cases for other scenarios (e.g., retrieving key from file)

	// Write key to temporary file
	tmpfile, err := os.CreateTemp("/tmp/", "mcli-key-*.txt")
	if err != nil {
		t.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up temporary file

	_, err = tmpfile.Write([]byte("test1234567890"))
	if err != nil {
		t.Errorf("failed to write key to temporary file: %v", err)
	}

	fileKey, err := aesct.GetKey(tmpfile.Name(), false)
	if err != nil {
		t.Errorf("Failed to retrieve key from environment variable: %v", err)
	}
	expectedKey = []byte{122, 254, 203, 8, 251, 107, 123, 171, 75, 180, 92, 117, 93, 133, 122, 231, 29, 144, 237, 43, 55, 137, 159, 41, 184, 187, 198, 174, 117, 138, 81, 117}
	if !bytes.Equal(fileKey, expectedKey) {
		t.Errorf("Retrieved key from file %s does not match expected key", tmpfile.Name())
	}

}

func TestGenerateRSACert(t *testing.T) {
	// Test case for generating a certificate with default values
	err := GenerateRSACert("test1", ".test-data", false, nil)
	if err != nil {
		t.Errorf("Failed to generate certificate: %v", err)
	}

	// Check if certificate files exist
	files := []string{"test1-private", "test1-public", "test1-cert", "test1.key", "test1.crt"}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Certificate file %s does not exist", file)
		}
	}

	// Clean up generated files
	os.Remove("test1-private")
	os.Remove("test1-public")
	os.Remove("test1-cert")
	os.Remove("test1.key")
	os.Remove("test1.crt")

	// Test case for generating a certificate with custom domains
	err = GenerateRSACert("test2", ".", false, []string{"example.com", "www.example.com"})
	if err != nil {
		t.Errorf("Failed to generate certificate: %v", err)
	}

	// Check if certificate files exist
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Certificate file %s does not exist", file)
		}
	}

	// Clean up generated files
	os.Remove("test2-private")
	os.Remove("test2-public")
	os.Remove("test2-cert")
	os.Remove("test2.key")
	os.Remove("test2.crt")
}
