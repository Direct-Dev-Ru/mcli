package mclicrypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GenerateCertSerialNumber() (*big.Int, error) {
	// Generate a random number using a secure random number generator
	max := new(big.Int).Lsh(big.NewInt(1), 128) // 128-bit serial number
	serial, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, err
	}
	return serial, nil
}

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

// GenerateRSACert generates an RSA certificate and saves it to the specified path.
// crtName specifies the base name for the certificate files (without extension).
// path specifies the directory where the certificate files will be saved.
// isCA specifies whether the certificate is a certificate authority (CA) certificate.
// crtDomains specifies the list of domains for the certificate.
// If crtDomains is empty, the default domain "localhost" will be used
func GenerateRSACert(crtName, path string, isCA bool, crtDomains []string) error {

	if len(crtDomains) == 0 {
		crtDomains = []string{"localhost"}
	}
	// http://golang.org/pkg/crypto/x509/#Certificate
	template := &x509.Certificate{
		IsCA:                  isCA,
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
	pathToSave := filepath.Clean(path + "/" + crtName + "-private")
	err = os.WriteFile(pathToSave, pkey, 0777)

	if err != nil {
		return err
	}

	// save public key
	pubkey, _ := x509.MarshalPKIXPublicKey(publickey)
	os.WriteFile(filepath.Clean(path+"/"+crtName+"-public"), pubkey, 0777)

	// save cert
	os.WriteFile(filepath.Clean(path+"/"+crtName+"-cert"), cert, 0777)

	var certOut, keyOut bytes.Buffer
	pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	pem.Encode(&keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privatekey)})
	os.WriteFile(filepath.Clean(path+"/"+crtName+".key"), keyOut.Bytes(), 0644)
	os.WriteFile(filepath.Clean(path+"/"+crtName+".crt"), certOut.Bytes(), 0644)

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

func GetPublicKeyFromFile(path string) (*rsa.PublicKey, error) {

	certpem, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certpem)
	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(block.Bytes)
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	// fmt.Println(rsaPublicKey.N)
	// fmt.Println(rsaPublicKey.E)
	return rsaPublicKey, nil
}

func GetPrivateKeyFromByte(privpem []byte) (*rsa.PrivateKey, error) {

	block, _ := pem.Decode(privpem)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key = key8.(*rsa.PrivateKey)
	}
	return key, nil
}

func GetPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {

	privpem, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privpem)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key = key8.(*rsa.PrivateKey)
	}
	// fmt.Println(key.N)
	return key, nil
}

func RSAConfigSetup(rsaPrivateKeyLocation, privatePassphrase, rsaPublicKeyLocation string) (*rsa.PrivateKey, error) {
	if rsaPrivateKeyLocation == "" {
		// println("No RSA Key given, generating temp one")
		return GenRSA(4096)
	}

	priv, err := os.ReadFile(rsaPrivateKeyLocation)
	if err != nil {
		// println("No RSA private key found, generating temp one")
		return GenRSA(4096)
	}

	privPem, _ := pem.Decode(priv)

	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("RSA private key is of the wrong type. %v", privPem.Type)
	}

	if x509.IsEncryptedPEMBlock(privPem) && privatePassphrase == "" {
		return nil, fmt.Errorf("passphrase is required to open private pem file")
	}

	var privPemBytes []byte

	if privatePassphrase != "" {
		privPemBytes, err = x509.DecryptPEMBlock(privPem, []byte(privatePassphrase))
	} else {
		privPemBytes = privPem.Bytes
	}

	var parsedKey interface{}
	//PKCS1
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		//If what you are sitting on is a PKCS#8 encoded key
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
			return nil, fmt.Errorf("unable to parse RSA private key")
		}
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unable to parse RSA private key")
	}

	pub, err := os.ReadFile(rsaPublicKeyLocation)
	if err != nil {
		return nil, fmt.Errorf("no RSA public key found, generating temp one")
	}

	pubPem, _ := pem.Decode(pub)
	if pubPem == nil {
		return nil, fmt.Errorf("use `ssh-keygen -f id_rsa.pub -e -m pem > id_rsa.pem` to generate the pem encoding of your RSA public key - rsa public key not in pem format")
	}

	if pubPem.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("rsa public key is of the wrong type")
	}

	if parsedKey, err = x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		return nil, fmt.Errorf("unable to parse RSA public key, generating a temp one")
	}

	var pubKey *rsa.PublicKey
	if pubKey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("unable to parse RSA public key, generating a temp one")
	}
	privateKey.PublicKey = *pubKey

	return privateKey, nil
}

// GenRSA returns a new RSA key of bits length
func GenRSA(bits int) (*rsa.PrivateKey, error) {

	key, err := rsa.GenerateKey(rand.Reader, bits)
	return key, err
}

// GenerateCACertificate generates a self-signed Certificate Authority (CA) certificate and saves it to the specified path (specify only folder).
// It also saves private key in the same folder
// It takes the following parameters:
// - pathToSaveCA: The directory where the CA certificate and private key will be saved.
// - caFileName: The base filename for the CA certificate and private key files.
// - commonName: The common name (CN) for the CA certificate.
// - orgName: The organization name (O) for the CA certificate.
// - country: The country (C) for the CA certificate.
// - location: The location (L) for the CA certificate.
// It returns an error if any.
func GenerateCACertificate(pathToSaveCA, caFileName, commonName, orgName, country, location string) error {

	certSerial, err := GenerateCertSerialNumber()
	if err != nil {
		return err
	}

	var ca *x509.Certificate = &x509.Certificate{
		SerialNumber: certSerial,
		Subject: pkix.Name{
			Organization: []string{orgName},
			Country:      []string{country},
			Province:     []string{location},
			Locality:     []string{location},
			// StreetAddress: []string{""},
			// PostalCode:    []string{""},
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	// Now in caBytes we have our generated certificate, which we can PEM encode for later use:
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	err = os.WriteFile(filepath.Clean(pathToSaveCA+"/"+caFileName+".crt"), caPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	err = os.WriteFile(filepath.Clean(pathToSaveCA+"/"+caFileName+".key"), caPrivKeyPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ReadCertificateFromByte(certPEM []byte) (*x509.Certificate, error) {

	// Decode PEM-encoded certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
func ReadCertificateFromFile(certFilePath string) (*x509.Certificate, error) {
	// Read certificate file
	certPEM, err := os.ReadFile(certFilePath)
	if err != nil {
		return nil, err
	}

	// Decode PEM-encoded certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GenerateCertificateWithCASign generates a certificate signed by a Certificate Authority (CA) and saves it to the specified path.
// It takes the following parameters:
//   - pathToSaveCertificate: The directory where the certificate and private key will be saved ( specify with .crt filename -
//     .key file will be placed nearly in the same folder).
//   - caPath: The path to the CA certificate file (.crt - key file must be in the same folder).
//   - orgName: The organization name (O) for the certificate.
//   - country: The country (C) for the certificate.
//   - location: The location (L) for the certificate.
//   - DNSNames: A list of DNS names for the certificate.
//   - IPAddresses: A list of IP addresses for the certificate.
//
// It returns the certificate and private key bytes, along with any error encountered.
// if caPath dont contain vlid cert - ca files (.crt and .key) will be generated in the {{pathToSaveCertificate}}/ca folder
func GenerateCertificateWithCASign(pathToSaveCertificate, caPath, orgName, country, location string, DNSNames []string, IPAddresses []net.IP) ([]byte, []byte, error) {

	if len(DNSNames) == 0 {
		DNSNames = []string{"localhost"}
	}

	if len(IPAddresses) == 0 {
		IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	}
	// read ca certificate
	commonName := DNSNames[0]
	ca, err := ReadCertificateFromFile(caPath)
	if err != nil {
		caPath = pathToSaveCertificate + "/ca/ca.crt"
		ca, err = ReadCertificateFromFile(caPath)
		if err != nil {
			os.MkdirAll(filepath.Clean(pathToSaveCertificate+"/ca"), 0755)
			err = GenerateCACertificate(filepath.Clean(pathToSaveCertificate+"/ca"), "ca", commonName, orgName, country, location)
			if err != nil {
				return nil, nil, err
			}
			ca, err = ReadCertificateFromFile(caPath)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	caPrivKey, err := GetPrivateKeyFromFile(strings.ReplaceAll(caPath, ".crt", ".key"))
	if err != nil {
		return nil, nil, err
	}

	certSerial, err := GenerateCertSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: certSerial,
		Subject: pkix.Name{
			Organization:  []string{orgName},
			Country:       []string{country},
			Province:      []string{location},
			Locality:      []string{location},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    commonName,
		},
		IPAddresses:  IPAddresses,
		DNSNames:     DNSNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	err = os.WriteFile(filepath.Clean(pathToSaveCertificate+"/"+commonName+".crt"), certPEM.Bytes(), 0644)
	if err != nil {
		return nil, nil, err
	}

	err = os.WriteFile(filepath.Clean(pathToSaveCertificate+"/"+commonName+".key"), certPrivKeyPEM.Bytes(), 0644)
	if err != nil {
		return nil, nil, err
	}

	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}

// CertSetup sets up TLS configurations for server and client communication.
// It takes the paths to the certificate (.crt file and .key file must be placed in the same folder)
// and CA (Certificate Authority) files as input - .crt file.
// It returns the TLS configurations for the server and client, along with any error encountered.
func CertSetup(certPath, caPath string) (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {

	certPEM, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return nil, nil, err
	}

	certPrivKeyPEM, err := os.ReadFile(filepath.Clean(strings.ReplaceAll(certPath, ".crt", ".key")))
	if err != nil {
		return nil, nil, err
	}

	caPEM, err := os.ReadFile(filepath.Clean(caPath))
	if err != nil {
		return nil, nil, err
	}

	serverCert, err := tls.X509KeyPair(certPEM, certPrivKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEM)
	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}

	return
}

func GenerateCertificateWithCASignV2(pathToSaveCertificate, caPem, caPrKeyPem, orgName, country, location string, DNSNames []string, IPAddresses []net.IP) ([]byte, []byte, error) {

	if len(DNSNames) == 0 {
		DNSNames = []string{"localhost"}
	}

	if len(IPAddresses) == 0 {
		IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	}

	commonName := DNSNames[0]

	// read ca certificate
	ca, err := ReadCertificateFromByte([]byte(caPem))
	if err != nil {
		return nil, nil, err
	}

	// read ca pk
	caPrivKey, err := GetPrivateKeyFromByte([]byte(caPrKeyPem))
	if err != nil {
		return nil, nil, err
	}

	// get serial for sertificate
	certSerial, err := GenerateCertSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	subjectKeyId, err := GenerateKey(5)
	if err != nil {
		return nil, nil, err
	}

	// set up certificate
	cert := &x509.Certificate{
		SerialNumber: certSerial,
		Subject: pkix.Name{
			Organization:  []string{orgName},
			Country:       []string{country},
			Province:      []string{location},
			Locality:      []string{location},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    commonName,
		},
		IPAddresses: IPAddresses,
		DNSNames:    DNSNames,
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		// SubjectKeyId: []byte{1, 2, 3, 4, 6},
		SubjectKeyId: subjectKeyId,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, nil, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		return nil, nil, err
	}

	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}
