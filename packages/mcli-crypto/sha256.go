package mclicrypto

import (
	"crypto/sha256"
)

func SHA_256(text string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(text))
	// fmt.Printf("%x\n", algorithm.Sum(nil))
	return algorithm.Sum(nil)
}
