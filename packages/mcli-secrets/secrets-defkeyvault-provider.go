package mclisecrets

import (
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_utils "mcli/packages/mcli-utils"
	"os"
)

type DefaultKeyAndVaultProvider struct {
	key       []byte
	keyPath   string
	vaultPath string
	cypher    mcli_crypto.AesCypherType
}

func NewDefaultKeyAndVaultProvider(vaultPath, keyPath string) (*DefaultKeyAndVaultProvider, error) {
	result := DefaultKeyAndVaultProvider{cypher: mcli_crypto.AesCypher}

	result.vaultPath = vaultPath

	keyContent, err := mcli_crypto.AesCypher.GetKey(keyPath, false)
	if err != nil {
		return nil, fmt.Errorf("error reading keyfile: %w", err)
	}
	result.key = keyContent
	result.keyPath = keyPath

	return &result, nil
}

func (dkvp DefaultKeyAndVaultProvider) GetKey() ([]byte, error) {
	return dkvp.key, nil
}

func (dkvp DefaultKeyAndVaultProvider) GetKeyPath() (string, error) {
	return dkvp.keyPath, nil
}

func (dkvp DefaultKeyAndVaultProvider) GetVaultPath() (string, error) {
	return dkvp.vaultPath, nil
}

func (dkvp DefaultKeyAndVaultProvider) GetVault() ([]byte, error) {
	_, _, err := mcli_utils.IsExistsAndCreate(dkvp.vaultPath, true, true)
	if err != nil {
		return nil, fmt.Errorf("error detecting or creating vault file: %w", err)
	}
	content, err := os.ReadFile(dkvp.vaultPath)
	if err != nil {
		return nil, fmt.Errorf("error reading vaultfile: %w", err)
	}

	return content, nil
}

func (dkvp DefaultKeyAndVaultProvider) SetVault(data []byte) error {
	_, _, err := mcli_utils.IsExistsAndCreate(dkvp.vaultPath, true, false)
	if err != nil {
		return fmt.Errorf("error detecting or creating vault file containing folder: %w", err)
	}
	err = os.WriteFile(dkvp.vaultPath, data, 0700)
	if err != nil {
		return fmt.Errorf("error writing vault file: %w", err)
	}
	return nil
}
