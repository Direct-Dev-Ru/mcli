package mcliredis

import (
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"

	"strings"
)

type RedisKeyAndVaultProvider struct {
	key        []byte
	keyPath    string
	redisStore *RedisStore
	cypher     mcli_crypto.AesCypherType
}

func NewRedisKeyAndVaultProvider(kvStore *RedisStore, keyPath string) (*RedisKeyAndVaultProvider, error) {
	result := RedisKeyAndVaultProvider{cypher: mcli_crypto.AesCypher, redisStore: kvStore}

	if strings.HasPrefix(keyPath, "http") {
		// TODO: write request to http(s) resource for key
		a := 1
		fmt.Println(a)
	} else {
		keyContent, err := mcli_crypto.AesCypher.GetKey(keyPath, false)
		if err != nil {
			return nil, fmt.Errorf("error reading keyfile: %w", err)
		}
		result.key = keyContent
		result.keyPath = keyPath
	}
	return &result, nil
}

func (knvp RedisKeyAndVaultProvider) GetKey() ([]byte, error) {
	return knvp.key, nil
}

func (knvp RedisKeyAndVaultProvider) GetKeyPath() (string, error) {
	return knvp.keyPath, nil
}

func (knvp RedisKeyAndVaultProvider) GetVaultPath() (string, error) {
	return "not implemented", fmt.Errorf("not implemented")
}

func (knvp RedisKeyAndVaultProvider) GetVault() (map[string][]byte, error) {
	rs := knvp.redisStore
	vault, err := rs.GetRecords("*")
	return vault, err
}

func (knvp RedisKeyAndVaultProvider) SetVault(data map[string][]byte) error {
	return nil
}
