/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_secrets "mcli/packages/mcli-secrets"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets to specified file asking in terminal for crypting key",
	Long: `You can export one or more secrets to separate file. 
	For example: mcli secrets export secret-1 secret-2 ... secret-n -d /home/username/export-vault
	You will be asked for secret key by printing in terminal to encrypt exported secrets ... Please remember this key
	Or you can use stdin for key definition
	`,
	Run: func(cmd *cobra.Command, args []string) {

		var vaultPath, expvault_path, keyFilePath, destkey_path string

		vaultPath, _ = cmd.Flags().GetString("vault-path")
		isVaultPathSet := cmd.Flags().Lookup("vault-path").Changed
		if !isVaultPathSet && len(Config.Secrets.Common.VaultPath) > 0 {
			vaultPath = Config.Secrets.Common.VaultPath
		}
		keyFilePath, _ = cmd.Flags().GetString("keyfile-path")
		isKeyFilePathSet := cmd.Flags().Lookup("keyfile-path").Changed
		if !isKeyFilePathSet && len(Config.Secrets.Common.KeyFilePath) > 0 {
			keyFilePath = Config.Secrets.Common.KeyFilePath
		}

		destkey_path, _ = cmd.Flags().GetString("destkey-path")
		expvault_path, _ = cmd.Flags().GetString("expvault-path")

		var key []byte
		var err error

		secretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)

		switch destkey_path {
		case "/input":
			key, err = mcli_crypto.GetKeyFromString(Input.joinedInput)
		case "/ask":
			key, err = mcli_crypto.GetKeyFromString(Input.joinedInput)
		default:
			key, err = secretStore.Cypher.GetKey(destkey_path, false)
		}
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}

		if err := secretStore.FillStore(vaultPath, keyFilePath); err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		if len(args) > 0 {
			filteredSecrets := make([]mcli_secrets.SecretEntry, 0, len(args))
			for _, s := range secretStore.Secrets {
				if slices.Contains(args, s.Name) {
					filteredSecrets = append(filteredSecrets, s)
				}
			}
			secretStore.Secrets = filteredSecrets
		}
		encData, err := secretStore.GetEncContent(key)
		if expvault_path == "/stdout" || len(expvault_path) == 0 {
			fmt.Println(string(encData))
		} else {
			os.WriteFile(expvault_path, encData, 0600)
			fmt.Println(expvault_path)
		}
	},
}

func init() {
	secretsCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("vault-path", "v", GlobalMap["HomeDir"]+"/.mcli/secrets/defvault", "path to source secret vault")
	exportCmd.Flags().StringP("expvault-path", "e", "/stdout", "path to destination secret vault or /stdout to out into terminal")
	exportCmd.Flags().StringP("keyfile-path", "k", "", "path to file to get access key of source secret vault")
	exportCmd.Flags().StringP("destkey-path", "d", "/input", "path to file to get access key of dest secret vault(or /input or /ask")
}
