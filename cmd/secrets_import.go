/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_secrets "mcli/packages/mcli-secrets"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

// exportCmd represents the export command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets from specified file asking in terminal for crypting key",
	Long: `You can import secrets from separate file. 
	For example: mcli secrets import secret-1 secret-2 ... secret-n d=/ask -e=/home/username/export.vault
	You will be asked for secret key to print it in terminal to decrypt imported secrets ... 
	Or you can use stdin for key definition
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var LineBreak string = GlobalMap["LineBreak"]
		var vaultPath, source_path, keyFilePath, importkey_path string

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

		importkey_path, _ = cmd.Flags().GetString("importkey-path")
		source_path, _ = cmd.Flags().GetString("source-path")
		_, _ = importkey_path, source_path

		var key []byte
		var err error

		// destination vault
		secretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)

		if err := secretStore.FillStore(vaultPath, keyFilePath); err != nil {
			Elogger.Fatal().Msg(err.Error())
		}

		// source store - to import from
		sourceStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)

		switch importkey_path {
		case "/input":
			key, err = mcli_crypto.GetKeyFromString(Input.joinedInput)
		case "/ask":
			reader := bufio.NewReader(os.Stdin)
			fmt.Print(ColorGreen + "your key: " + ColorReset)
			keyString, _ := reader.ReadString('\n')
			if keyString == LineBreak {
				fmt.Print(ColorYellow + "your key: " + ColorReset)
				keyString, _ = reader.ReadString('\n')
				if keyString == LineBreak {
					fmt.Println(ColorRed + "Empty key provided ... Good luck" + ColorReset)
					Elogger.Fatal().Msg("mcli secrets import: key is empty from keyboard")
				}
			}
			key, err = mcli_crypto.GetKeyFromString(keyString)
		default:
			key, err = sourceStore.Cypher.GetKey(importkey_path, false)
		}
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		encContent, err := os.ReadFile(source_path)
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		err = sourceStore.GetFromEncContent(encContent, key)
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}

		fmt.Println(secretStore.Secrets)
		fmt.Println(sourceStore.Secrets)

		if len(args) > 0 {
			filteredSecrets := make([]mcli_secrets.SecretEntry, 0, len(args))
			for _, s := range secretStore.Secrets {
				if slices.Contains(args, s.Name) {
					filteredSecrets = append(filteredSecrets, s)
				}
			}
			secretStore.Secrets = filteredSecrets
		}

	},
}

func init() {
	secretsCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("vault-path", "v", GlobalMap["HomeDir"]+"/.mcli/secrets/defvault", "path to destination secret vault")
	importCmd.Flags().StringP("keyfile-path", "k", "", "path to file to get access key of destination secret vault")

	importCmd.Flags().StringP("source-path", "s", "", "path to source secret vault")
	importCmd.Flags().StringP("importkey-path", "i", "/input", "path to file to get access key of source secret vault (or /input or /ask")
}
