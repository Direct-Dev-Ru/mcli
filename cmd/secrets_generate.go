/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Toolza for password phrase generation",
	Long: `This command helps generate russians words based passphrase and optionally store it in vault
For example: mcli secrets generate --use-words 
`,
	Run: func(cmd *cobra.Command, args []string) {
		var LineBreak string = GlobalMap["LineBreak"]

		var useWords bool
		var dictPath, vaultPath, keyFilePath string
		var minLength, maxLength int

		useWords, _ = cmd.Flags().GetBool("use-words")
		isUseWordsSet := cmd.Flags().Lookup("use-words").Changed
		if !isUseWordsSet {
			useWords = Config.Secrets.Common.UseWords
		}

		dictPath, _ = cmd.Flags().GetString("dict-path")
		isDictPathSet := cmd.Flags().Lookup("dict-path").Changed
		if !isDictPathSet && len(Config.Secrets.Common.DictPath) > 0 {
			dictPath = Config.Secrets.Common.DictPath
		}
		vaultPath, _ = cmd.Flags().GetString("vault-path")
		isVaultPathSet := cmd.Flags().Lookup("vault-path").Changed
		// fmt.Println(vaultPath, isVaultPathSet, Config.Secrets.Common.VaultPath)
		if !isVaultPathSet && len(Config.Secrets.Common.VaultPath) > 0 {
			vaultPath = Config.Secrets.Common.VaultPath
		}
		keyFilePath, _ = cmd.Flags().GetString("keyfile-path")
		isKeyFilePathSet := cmd.Flags().Lookup("keyfile-path").Changed
		if !isKeyFilePathSet && len(Config.Secrets.Common.KeyFilePath) > 0 {
			keyFilePath = Config.Secrets.Common.KeyFilePath
		}

		// Ilogger.Trace().Bool("use-words", useWords).Str("dict-path", dictPath).Str("vault-path", vaultPath).
		// 	Str("keyfile-path", keyFilePath).Send()

		minLength, _ = cmd.Flags().GetInt("min-length")
		isMinLengthSet := cmd.Flags().Lookup("min-length").Changed
		if !isMinLengthSet && Config.Secrets.Common.MinLength > 7 {
			minLength = Config.Secrets.Common.MinLength
		}
		maxLength, _ = cmd.Flags().GetInt("max-length")
		isMaxLengthSet := cmd.Flags().Lookup("max-length").Changed
		if !isMaxLengthSet && Config.Secrets.Common.MaxLength > 11 {
			maxLength = Config.Secrets.Common.MaxLength
		}
		if IsCommanInPipe() {
			Elogger.Fatal().Msg("generate doesn't support pipe yet ...")
		}
		var runesReplaces []mcli_crypto.ReplaceEntry = make([]mcli_crypto.ReplaceEntry, 3)
		runesReplaces[0] = mcli_crypto.ReplaceEntry{OriginRune: 'а', ReplaceRune: '@', Number: 1}
		runesReplaces[1] = mcli_crypto.ReplaceEntry{OriginRune: 'А', ReplaceRune: '@', Number: 1}
		runesReplaces[2] = mcli_crypto.ReplaceEntry{OriginRune: 'О', ReplaceRune: '0', Number: 1}

		secretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)

		var phrase string
		var err error

		if err := secretStore.FillStore(vaultPath, keyFilePath); err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		// Ilogger.Trace().Msg("storeContent: " + fmt.Sprintf("%v", secretStore.Secrets))

		for {
			fmt.Println(ColorGreen + "-------------------Start Generation------------------------------" + ColorReset)
			// reading from stdin
			reader := bufio.NewReader(os.Stdin)
			fmt.Print(ColorGreen + "Please enter secret name: " + ColorReset)
			name, _ := reader.ReadString('\n')
			name = strings.TrimSuffix(name, LineBreak)
			fmt.Print(ColorGreen + "Please enter secret login: " + ColorReset)
			login, _ := reader.ReadString('\n')
			login = strings.TrimSuffix(login, LineBreak)

			fmt.Print(ColorGreen + "Please enter secret description: " + ColorReset)
			descr, _ := reader.ReadString('\n')
			descr = strings.TrimSuffix(descr, LineBreak)

			if useWords {
				phrase, err = mcli_crypto.GeneratePassPhrase(dictPath, runesReplaces)
			} else {
				phrase, err = mcli_crypto.GeneratePassword(mcli_utils.Random(minLength, maxLength))
			}
			if err != nil {
				Elogger.Fatal().Msg(err.Error())
			}

			fmt.Print(ColorYellow + "Is secret " + ColorBlue + phrase + ColorYellow + " is good enougth for you (or enter another if you want): " + ColorReset)
			ownSecret, _ := reader.ReadString('\n')
			if ownSecret != LineBreak {
				fmt.Print(ColorBlue + "Retype your own secret: " + ColorReset)
				reOwnSecret, _ := reader.ReadString('\n')
				if ownSecret != reOwnSecret {
					fmt.Print(ColorRed + "Secrets dont match (enter your secret): " + ColorReset)
					ownSecret, _ = reader.ReadString('\n')
					fmt.Print(ColorRed + "Retype your own secret: " + ColorReset)
					reOwnSecret, _ = reader.ReadString('\n')
					if ownSecret != reOwnSecret {
						fmt.Println(ColorRed + "Secrets dont match - Good buy !!! (please take a lesson of keyboard using )" + ColorReset)
						Elogger.Fatal().Msg("mcli secrets: secrets dont match - Good buy !!! (please take a lesson of keyboard using )")
					}
				}
				if len(ownSecret) < minLength {
					fmt.Println(ColorRed + "Secret too slow - Good buy !!! (please generate in you mind more complex secret )" + ColorReset)
					Elogger.Fatal().Msg("mcli secrets: secret too easy - Good buy !!! (please generate in you mind more complex secret )")
				}
				phrase = strings.TrimSuffix(ownSecret, LineBreak)
			}

			// secretEntry := mcli_secrets.SecretEntry{Name: name, Description: descr,
			// 	Login: login, Secret: phrase, CreatedAt: time.Now()}
			// secretEntry.SetSecret(phrase, []byte{}, false)

			secretEntry, err := secretStore.NewEntry(name, login, descr)
			if err != nil {
				Elogger.Fatal().Msgf("generate: new secret error: %v", err)
			}
			secretEntry.SetSecret(phrase, true, true)

			fmt.Print(ColorYellow + "Store Secret Entry to Secret Vault? Enter yes or no: " + ColorReset)
			cmd, _ := reader.ReadString('\n')
			cmd = strings.TrimSuffix(cmd, LineBreak)
			cmd = strings.ToLower(cmd)

			if strings.Contains("y да yes д ", cmd+" ") {
				secretStore.AddEntry(secretEntry)
				secretStore.Save("", "")
			}

			fmt.Println(ColorGreen + "-----------------------------------------------------------------" + ColorReset)
			fmt.Print(ColorPurple + "Enter (q)uit to quit generation or any key to continue: ")
			cmd, _ = reader.ReadString('\n')
			cmd = strings.TrimSuffix(cmd, LineBreak)
			// runeCmd := []rune(cmd)

			if !(cmd == "quit" || cmd == "q") {
				continue
			} else {
				break
			}
		}

	},
}

func init() {
	secretsCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("dict-path", "d", "", "path to csv word dictionary")
	generateCmd.Flags().StringP("vault-path", "v", GlobalMap["HomeDir"]+"/.mcli/secrets/defvault", "path to vault")
	generateCmd.Flags().StringP("keyfile-path", "k", "", "path to file to get access key")
	generateCmd.Flags().BoolP("use-words", "w", false, "generate by russian words toggle")
	generateCmd.Flags().IntP("min-length", "m", 8, "min length of password (affected then use-words eq false )")
	generateCmd.Flags().IntP("max-length", "x", 24, "max length of password (affected then use-words eq false )")
}
