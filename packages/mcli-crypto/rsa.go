package mclicrypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func GenerateRsaCrtRequest(name, user, path string) error {

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	keyDer := x509.MarshalPKCS1PrivateKey(key)
	keyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyDer,
	}
	keyFile, err := os.Create(filepath.Clean(path + "/" + name + "-key.pem"))
	if err != nil {
		return err
	}

	pem.Encode(keyFile, &keyBlock)
	keyFile.Close()

	commonName := user

	emailAddress := "info@direct-dev.ru"
	org := "Direct-Dev.ru"
	orgUnit := "IT Dept."
	city := "Omsk"
	state := "Omsk"
	country := "RU"
	subject := pkix.Name{
		CommonName:         commonName,
		Country:            []string{country},
		Locality:           []string{city},
		Organization:       []string{org},
		OrganizationalUnit: []string{orgUnit},
		Province:           []string{state},
	}
	asn1, err := asn1.Marshal(subject.ToRDNSequence())
	if err != nil {
		return err
	}
	csr := x509.CertificateRequest{
		RawSubject:         asn1,
		EmailAddresses:     []string{emailAddress},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	bytes, err := x509.CreateCertificateRequest(rand.Reader, &csr, key)
	if err != nil {
		return err
	}
	csrFile, err := os.Create(filepath.Clean(path + "/" + name + ".csr"))
	if err != nil {
		return err
	}
	pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: bytes})
	csrFile.Close()
	return nil
}

func GenerateRsaCert(crtName, path string, crtDomains []string) error {

	if len(crtDomains) == 0 {
		crtDomains = []string{"localhost"}
	}
	// http://golang.org/pkg/crypto/x509/#Certificate
	template := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"RU"},
			Organization: []string{"Direct-Dev-Ru"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 5, 5),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames:    crtDomains,
	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return err
	}

	publickey := &privatekey.PublicKey

	// create a self-signed certificate. template => parent
	var parent = template
	cert, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, privatekey)

	if err != nil {
		return err
	}

	// save private key
	pkey := x509.MarshalPKCS1PrivateKey(privatekey)
	os.WriteFile(filepath.Clean(path+"/"+crtName+"-private"), pkey, 0777)

	// save public key
	pubkey, _ := x509.MarshalPKIXPublicKey(publickey)
	os.WriteFile(filepath.Clean(path+"/"+crtName+"-public"), pubkey, 0777)

	// save cert
	os.WriteFile(filepath.Clean(path+"/"+crtName+"-cert"), cert, 0777)

	var certOut, keyOut bytes.Buffer
	pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	pem.Encode(&keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privatekey)})
	os.WriteFile(filepath.Clean(path+"/"+crtName+".key"), keyOut.Bytes(), 0644)
	os.WriteFile(filepath.Clean(path+"/"+crtName+".cert"), certOut.Bytes(), 0644)

	// these are the files save with encoding/gob style
	// privkeyfile, _ := os.Create(filepath.Clean(path + "/" + "privategob.key"))
	// privkeyencoder := gob.NewEncoder(privkeyfile)
	// privkeyencoder.Encode(privatekey)
	// privkeyfile.Close()

	// pubkeyfile, _ := os.Create(filepath.Clean(path + "/" + "publickgob.key"))
	// pubkeyencoder := gob.NewEncoder(pubkeyfile)
	// pubkeyencoder.Encode(publickey)
	// pubkeyfile.Close()

	// this will create plain text PEM file.
	// pemfile, _ := os.Create(filepath.Clean(path + "/" + "certpem.pem"))
	// var pemkey = &pem.Block{
	// 	Type:  "RSA PRIVATE KEY",
	// 	Bytes: x509.MarshalPKCS1PrivateKey(privatekey)}
	// pem.Encode(pemfile, pemkey)
	// pemfile.Close()
	return nil
}

func GenRsa() ([]*rsa.PrivateKey, error) {
	firstPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	// first = maria
	// firstPublicKey := &firstPrivateKey.PublicKey
	secondPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	// secondPublicKey := &secondPrivateKey.PublicKey

	return []*rsa.PrivateKey{firstPrivateKey, secondPrivateKey}, nil
}
