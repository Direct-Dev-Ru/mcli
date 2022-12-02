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
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

// removeCmd represents the remove subcommand of secrets command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes one or more secrets from vault",
	Long: `You can remove one or more secrets from secret vault. 
	For example: mcli secrets remove secret-1 secret-2 ... secret-n	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var LineBreak string = GlobalMap["LineBreak"]
		var vaultPath, keyFilePath string
		var confirm bool

		if len(args) == 0 {
			fmt.Println("Specify secret name(s) to remove through argument list, e.g. mcli secrets remove secret-1 secret-2")
			os.Exit(0)
		}
		confirm, _ = cmd.Flags().GetBool("confirm")

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

		secretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)
		if err := secretStore.FillStore(vaultPath, keyFilePath); err != nil {
			Elogger.Fatal().Msg(err.Error())
		}

		newSecretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)

		var removed []string = make([]string, 0, len(args))
		for _, s := range secretStore.Secrets {
			currentName := s.Name
			var remove bool = false

			reader := bufio.NewReader(os.Stdin)

			if slices.Contains(args, currentName) {
				remove = true
				if !confirm {
					remove = false
					fmt.Print(ColorYellow + "Remove Entry " + currentName + " from secret vault? Enter yes or no: " + ColorReset)
					cmd, _ := reader.ReadString('\n')
					cmd = strings.TrimSuffix(cmd, LineBreak)
					cmd = strings.ToLower(cmd)
					if strings.Contains("y да yes д ", cmd+" ") {
						remove = true
					}
				}
			}
			if !remove {
				err := newSecretStore.AddEntry(s)
				if err != nil {
					Elogger.Fatal().Msg(err.Error())
				}
			} else {
				removed = append(removed, currentName)
			}
		}

		err := newSecretStore.Save(vaultPath, keyFilePath)
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		if len(removed) == 0 {
			fmt.Println("no secrets have been removed")
		} else {
			for _, v := range removed {
				fmt.Println(v)
			}
		}
	},
}

func init() {
	secretsCmd.AddCommand(removeCmd)

	removeCmd.Flags().StringP("vault-path", "v", GlobalMap["HomeDir"]+"/.mcli/secrets/defvault", "path to vault")
	removeCmd.Flags().StringP("keyfile-path", "k", "", "path to file to get access key")
	removeCmd.Flags().BoolP("confirm", "y", false, "dont ask confirm on removing entries")
}
