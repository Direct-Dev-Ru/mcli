/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/spf13/cobra"
)

func readCsv(path string) (map[string][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("cannot open CSV file:" + err.Error())
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, errors.New("cannot open CSV file:" + err.Error())
	}
	result := make(map[string][]string)

	for _, row := range rows {

		if len(row) == 1 {
			row = strings.Fields(row[0])
		}

		if len(row) >= 2 {
			s := strings.TrimSpace(row[0])
			if len(s) > 3 && !strings.Contains(s, "-") {
				result[row[1]] = append(result[row[1]], strings.TrimSpace(row[0]))
			}
		}
	}

	return result, nil
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Toolza for password phrase generation",
	Long: `This command helps generate russians words based passphrase and optionally store it in vault
For example: mcli secrets generate --use-words 
`,
	Run: func(cmd *cobra.Command, args []string) {
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
		fmt.Println(vaultPath, isVaultPathSet, Config.Secrets.Common.VaultPath)
		if !isVaultPathSet && len(Config.Secrets.Common.VaultPath) > 0 {
			vaultPath = Config.Secrets.Common.VaultPath
		}
		keyFilePath, _ = cmd.Flags().GetString("keyfile-path")
		isKeyFilePathSet := cmd.Flags().Lookup("keyfile-path").Changed
		if !isKeyFilePathSet && len(Config.Secrets.Common.KeyFilePath) > 0 {
			keyFilePath = Config.Secrets.Common.KeyFilePath
		}

		Ilogger.Trace().Bool("use-words", useWords).Str("dict-path", dictPath).Str("vault-path", vaultPath).
			Str("keyfile-path", keyFilePath).Send()

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

		var runesReplaces []mcli_crypto.ReplaceEntry = make([]mcli_crypto.ReplaceEntry, 3)
		runesReplaces[0] = mcli_crypto.ReplaceEntry{OriginRune: 'а', ReplaceRune: '@', Number: 1}
		runesReplaces[1] = mcli_crypto.ReplaceEntry{OriginRune: 'А', ReplaceRune: '@', Number: 1}
		runesReplaces[2] = mcli_crypto.ReplaceEntry{OriginRune: 'О', ReplaceRune: '0', Number: 1}

		var phrase string
		var err error

		theKey, err := mcli_crypto.GetKeyFromFile(keyFilePath)
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		theKeyString := mcli_crypto.SHA_256(string(theKey))
		// fmt.Println(phrase)
		fmt.Println(theKeyString, len(theKeyString))

		for {
			// reading from stdin
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Please enter secret name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSuffix(name, "\n")
			fmt.Print("Please enter secret login: ")
			login, _ := reader.ReadString('\n')
			login = strings.TrimSuffix(login, "\n")

			fmt.Print("Please enter secret description: ")
			descr, _ := reader.ReadString('\n')
			descr = strings.TrimSuffix(descr, "\n")

			if useWords {
				phrase, err = mcli_crypto.GeneratePassPhrase(dictPath, runesReplaces)
			} else {
				phrase, err = mcli_crypto.GeneratePassword(mcli_utils.Random(minLength, maxLength))
			}
			if err != nil {
				Elogger.Fatal().Msg(err.Error())
			}

			fmt.Print("Suggested secret " + phrase + " (enter another if you want): ")
			ownSecret, _ := reader.ReadString('\n')
			if ownSecret != "\n" {
				fmt.Print("Retype your own secret: ")
				reOwnSecret, _ := reader.ReadString('\n')
				if ownSecret != reOwnSecret {
					fmt.Print("Secrets dont match (enter your secret): ")
					ownSecret, _ = reader.ReadString('\n')
					fmt.Print("Retype your own secret: ")
					reOwnSecret, _ = reader.ReadString('\n')
					if ownSecret != reOwnSecret {
						fmt.Println("Secrets dont match - Good buy !!! (please take a lesson of keyboard using )")
						Elogger.Fatal().Msg("mcli secrets: secrets dont match - Good buy !!! (please take a lesson of keyboard using )")
					}
				}
				if len(ownSecret) < minLength {
					fmt.Println("Secret too slow - Good buy !!! (please generate in you mind more complex secret )")
					Elogger.Fatal().Msg("mcli secrets: secret too easy - Good buy !!! (please generate in you mind more complex secret )")
				}
				phrase = strings.TrimSuffix(ownSecret, "\n")
			}
			nowTime := time.Now()

			secretEntry := mcli_crypto.SecretEntry{Name: name, Description: descr, Login: login, Secret: phrase, CreatedAt: nowTime}

			fmt.Print("Store Secret Enrty to Secret Vault? Enter yes or no: ")
			cmd, _ := reader.ReadString('\n')
			cmd = strings.TrimSuffix(cmd, "\n")
			cmd = strings.ToLower(cmd)
			// if cmd = "y" || cmd = "yes" || "д" || cmd = "да" || cmd = "да"
			if strings.Contains("y да yes д ", cmd+" ") {

				fmt.Println(secretEntry)
				os.Setenv("SECRETS_SECRET", phrase)

				// shellCmd := exec.Command("echo", "SECRETS_SECRET="+phrase)
				// str, err := shellCmd.Output()
				// fmt.Println(string(str), err)
			}

			fmt.Print("Enter (q)uit to quit generation or any key to continue: ")
			cmd, _ = reader.ReadString('\n')
			// runeCmd := []rune(cmd)

			if !(cmd == "quit\n" || cmd == "q\n") {
				continue
			} else {
				break
			}
		}

		// encByteArray, err := mcli_crypto.Encrypt(theKeyString, []byte(phrase))
		// encString := hex.EncodeToString(encByteArray)
		// fmt.Println(encByteArray)
		// fmt.Println(encString)
		// decByteArray, _ := hex.DecodeString(encString)
		// decByteArray, _ = mcli_crypto.Decrypt(theKeyString, decByteArray)
		// decString := string(decByteArray)
		// fmt.Println(decByteArray)
		// fmt.Println(decString)

	},
}

func init() {
	secretsCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("dict-path", "d", "", "path to csv word dictionary")
	generateCmd.Flags().StringP("vault-path", "v", os.Getenv("HOME")+"/.mcli/secrets/defvault", "path to vault")
	generateCmd.Flags().StringP("keyfile-path", "k", "/dev/random", "path to file to get vault key or /dev/random to get random key")
	generateCmd.Flags().BoolP("use-words", "w", false, "generate by russian words toggle")
	generateCmd.Flags().IntP("min-length", "m", 8, "min length of password (affected then use-words eq false )")
	generateCmd.Flags().IntP("max-length", "x", 24, "max length of password (affected then use-words eq false )")
}
