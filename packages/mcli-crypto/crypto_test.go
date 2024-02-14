package mclicrypto

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
	pathToSave := "/home/kay/project/golang/golang-cli/.test-data"
	// Test case for generating a certificate with default values
	err := GenerateRSACert("test1", pathToSave, false, nil)
	if err != nil {
		t.Errorf("Failed to generate certificate: %v", err)
	}

	// Check if certificate files exist
	files := []string{"test1-private", "test1-public", "test1-cert", "test1.key", "test1.crt"}
	for _, file := range files {
		if _, err := os.Stat(pathToSave + "/" + file); os.IsNotExist(err) {
			t.Errorf("Certificate file %s does not exist", file)
		}
	}

	// Clean up generated files
	os.Remove(pathToSave + "/" + "test1-private")
	os.Remove(pathToSave + "/" + "test1-public")
	os.Remove(pathToSave + "/" + "test1-cert")

	// Test case for generating a certificate with custom domains
	err = GenerateRSACert("test2", pathToSave, false, []string{"site.direct-dev.ru", "www.site.direct-dev.ru"})
	if err != nil {
		t.Errorf("Failed to generate certificate: %v", err)
	}

	// Check if certificate files exist
	files = []string{"test2-private", "test2-public", "test2-cert", "test2.key", "test2.crt"}
	for _, file := range files {
		if _, err := os.Stat(pathToSave + "/" + file); os.IsNotExist(err) {
			t.Errorf("Certificate file %s does not exist", file)
		}
	}

	// Clean up generated files
	os.Remove(pathToSave + "/" + "test2-private")
	os.Remove(pathToSave + "/" + "test2-public")
	os.Remove(pathToSave + "/" + "test2-cert")
}

func TestGenerateCACert(t *testing.T) {
	pathToSave := "/home/kay/project/golang/golang-cli/.test-data"
	caFileName := "ca-test"
	// Test case for generating a certificate with default values
	err := GenerateCACertificate(pathToSave, caFileName, "Direct-Dev.Ru", "Direct-Dev.Ru", "RU", "OMSK")
	if err != nil {
		t.Errorf("Failed to generate certificate: %v", err)
	}

	// Check if certificate files exist
	files := []string{"ca-test.crt", "ca-test.key"}
	for _, file := range files {
		if _, err := os.Stat(pathToSave + "/" + file); os.IsNotExist(err) {
			t.Errorf("Certificate file %s does not exist", file)
		}
	}
}

func TestGenerateCertWithCASign(t *testing.T) {
	pathToSave := "/home/kay/project/golang/golang-cli/.test-data"
	caPath := "/home/kay/project/golang/golang-cli/.test-data" + "/ca-test.crt"

	_, _, err := GenerateCertificateWithCASign(pathToSave, caPath, "Direct-Dev.ru", "RU", "OMSK", []string{"example.direct-dev.ru"}, []net.IP{})
	if err != nil {
		t.Errorf("Failed to generate certificate: %v", err)
	}

	serverTLSConf, clientTLSConf, err := CertSetup("/home/kay/project/golang/golang-cli/.test-data/example.direct-dev.ru.crt",
		"/home/kay/project/golang/golang-cli/.test-data/ca/ca.crt")
	if err != nil {
		t.Errorf("Failed to get tls configs: %v", err)
	}

	// set up the httptest.Server using our certificate signed by our CA
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "success!")
		t.Logf("success!")
	}))
	server.TLS = serverTLSConf
	server.StartTLS()
	defer server.Close()

	// communicate with the server using an http.Client configured to trust our CA
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}
	http := http.Client{
		Transport: transport,
	}
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("Failed to GET server: %v", err)
	}

	// verify the response
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
	}
	body := strings.TrimSpace(string(respBodyBytes[:]))
	if body == "success!" {
		t.Log(body)
	} else {
		t.Errorf("not success !!!")
	}

}
