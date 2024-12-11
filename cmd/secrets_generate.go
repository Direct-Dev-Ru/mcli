/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_type "mcli/packages/mcli-type"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/spf13/cobra"
)

type InputSecretEntry struct {
	Name        string `json:"Name"`
	Login       string `json:"Login"`
	Description string `json:"Description"`
	Secret      string `json:"Secret"`
}

func GenerateSecret(cmd *cobra.Command, args []string) {
	var LineBreak string = GlobalMap["LineBreak"]

	var useWords, obfuscate, translit, ruToEn bool
	var dictPath, vaultPath, keyFilePath string
	var minLength, maxLength int

	useWords, _ = cmd.Flags().GetBool("use-words")
	isUseWordsSet := cmd.Flags().Lookup("use-words").Changed
	if !isUseWordsSet {
		useWords = Config.Secrets.Common.UseWords
	}
	obfuscate, _ = cmd.Flags().GetBool("obfuscate")
	isObfuscateSet := cmd.Flags().Lookup("obfuscate").Changed
	if !isObfuscateSet {
		obfuscate = Config.Secrets.Common.Obfuscate
	}
	translit, _ = cmd.Flags().GetBool("translit")
	isTranslitSet := cmd.Flags().Lookup("translit").Changed
	if !isTranslitSet {
		translit = false
	}
	ruToEn, _ = cmd.Flags().GetBool("qwerty")
	isRuToEnSet := cmd.Flags().Lookup("qwerty").Changed
	if !isRuToEnSet {
		ruToEn = false
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
	// fmt.Println("keyFilePath: ", keyFilePath, Config.Secrets.Common.KeyFilePath)
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

	var runesReplaces []mcli_crypto.ReplaceEntry = make([]mcli_crypto.ReplaceEntry, 3)
	if obfuscate {
		runesReplaces[0] = mcli_crypto.ReplaceEntry{OriginRune: 'a', ReplaceRune: '@', Number: 1000}
		runesReplaces[1] = mcli_crypto.ReplaceEntry{OriginRune: 'O', ReplaceRune: '0', Number: 1000}
		runesReplaces[2] = mcli_crypto.ReplaceEntry{OriginRune: 'i', ReplaceRune: '1', Number: 1000}
	}

	var knvp mcli_type.KeyAndVaultProvider
	var err error
	knvp, err = mcli_secrets.NewDefaultKeyAndVaultProvider(vaultPath, keyFilePath)
	if err != nil {
		Elogger.Fatal().Msgf("get default knv provider fault: %v", err)
	}
	// init secret store
	secretStore := mcli_secrets.NewSecretsEntriesV2(mcli_fs.GetFile, mcli_fs.SetFile,
		mcli_crypto.AesCypher, nil, knvp)

	var phrase string
	// fill secret store
	if err := secretStore.FillStoreV2(); err != nil {
		Elogger.Fatal().Msg(err.Error())
	}
	// if err := secretStore.FillStore(vaultPath, keyFilePath); err != nil {
	// 	Elogger.Fatal().Msg(err.Error())
	// }
	if os.Getenv("DEBUG") == "true" {
		fmt.Println(secretStore.Secrets)
	}

	if IsCommandInPipe() && len(Input.InputSlice) == 0 {
		Elogger.Fatal().Msg("input from pipe is empty: provide not empty pipe or run without pipe")
	}
	// process if stdin contains stream of objects to generate secrets
	if IsCommandInPipe() && len(Input.InputSlice) > 0 {
		Ilogger.Trace().Msg("input from pipe ...")
		// input should be blocks of data if plain input:
		// Name: some_secret_name
		// Login: some_secret_login
		// Secret: secret_phrase (optional - shoul be base64 encoded)
		// Description: some_secret_description
		// or if input is json
		// {"Name": "some_secret_name_1", "Login": "some_secret_login_1",
		// "Description": "some_secret_description_1", "Secret": "secret_phrase_1" (optional)}
		// {"Name": "some_secret_name_2", "Login": "some_secret_login_2",
		// "Description": "some_secret_description_2", "Secret": "secret_phrase_2" (optional)}
		inData := make([]string, 0, len(Input.InputSlice))
		for _, inputLine := range Input.InputSlice {
			currentLine := strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], "")
			currentLine = strings.TrimSpace(currentLine)
			if len(currentLine) == 0 {
				continue
			}
			inData = append(inData, currentLine)
		}

		// detect json or not
		a1 := strings.HasPrefix(inData[0], "{")
		a2 := strings.HasSuffix(inData[len(inData)-1], "}")
		b1 := strings.HasPrefix(inData[0], "[")
		b2 := strings.HasSuffix(inData[len(inData)-1], "]")
		var inputSecrets []InputSecretEntry = make([]InputSecretEntry, 0, 2)
		if a1 && a2 {
			dec := json.NewDecoder(strings.NewReader(strings.Join(inData, " ")))
			for dec.More() {
				var entry InputSecretEntry
				err := dec.Decode(&entry)
				if err != nil {
					Elogger.Fatal().Msgf("error decoding input json sequence: %v", err)
				}
				inputSecrets = append(inputSecrets, entry)
			}
		} else if b1 && b2 {
			// input is a json array string

			err = json.NewDecoder(strings.NewReader(strings.Join(inData, " "))).Decode(&inputSecrets)
			if err != nil {
				Elogger.Fatal().Msgf("error decoding input json array: %v", err)
			}
		} else {
			// plain input
			fmt.Println("plain input:")
			current := InputSecretEntry{Name: "#@IniT@#", Secret: "generate"}
			for _, v := range inData {

				splitted := strings.SplitN(v, ":", 2)
				// fmt.Println(splitted, len(splitted))
				if len(splitted) != 2 {
					Elogger.Fatal().Msg("error decoding plain input: wrong format")
				}
				splitted[0] = strings.TrimSpace(splitted[0])
				splitted[0] = strings.Trim(splitted[0], `"`)
				splitted[1] = strings.TrimSpace(splitted[1])
				splitted[1] = strings.Trim(splitted[1], `"`)
				switch splitted[0] {
				case "Name":
					if current.Name != "#@IniT@#" {
						inputSecrets = append(inputSecrets, current)
					}
					current = InputSecretEntry{Secret: "generate"}
					current.Name = splitted[1]
				case "Login":
					current.Login = splitted[1]
				case "Description":
					current.Description = splitted[1]
				case "Secret":
					current.Secret = splitted[1]
				}

				// fmt.Printf("%v(len=%v):%+v\n", k, len(v), []rune(v))
			}
			fmt.Printf("%v", inputSecrets)
		}
		for i, sec := range inputSecrets {
			ssecret := sec.Secret
			if len(ssecret) == 0 || strings.ToLower(ssecret) == "generate" {
				inputSecrets[i].Secret = "generate"
				continue
			} else {
				r, err := regexp.Compile(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`)
				if err != nil {
					inputSecrets[i].Secret = "generate"
					continue
				}
				if matched := r.MatchString(ssecret); !matched {
					inputSecrets[i].Secret = "generate"
					continue
				}
				inputSecrets[i].Secret, _ = mcli_crypto.Base64Decode(ssecret)
			}
		}

		// now we are ready to generate secrets and save it to secret vault
		for _, sec := range inputSecrets {
			WgGlb.Add(1)
			go func(sec InputSecretEntry) {
				defer WgGlb.Done()
				name := sec.Name
				login := sec.Login
				descr := sec.Description
				phrase := sec.Secret

				if phrase == "generate" {
					if useWords {
						phrase, err = mcli_crypto.GeneratePassPhrase(dictPath, runesReplaces, translit, ruToEn)
					} else {
						phrase, err = mcli_crypto.GeneratePassword(mcli_utils.Random(minLength, maxLength))
					}
					if err != nil {
						Elogger.Fatal().Msg(err.Error())
					}
				}

				secretEntry, err := secretStore.NewEntry(name, login, descr)
				if err != nil {
					Elogger.Fatal().Msgf("generate: new secret entry error: %v", err)
				}
				secretEntry.SetSecret(phrase, true, false)

				Ilogger.Trace().Msgf("[%+v] %+v\n", phrase, secretEntry)
				// fmt.Printf("[%+v] %+v\n", phrase, secretEntry)

				err = secretStore.AddEntry(secretEntry)
				if err != nil {
					Elogger.Fatal().Msgf("generate: add secret entry error: %v", err)
				}
			}(sec)
		}
		WgGlb.Wait()
		err = secretStore.Save("", "")
		if err != nil {
			Elogger.Fatal().Msgf("generate: save secret error: %v", err)
		}
		return

	} else {
		// manual enter secret parameters
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
				phrase, err = mcli_crypto.GeneratePassPhrase(dictPath, runesReplaces, translit, ruToEn)
			} else {
				phrase, err = mcli_crypto.GeneratePassword(mcli_utils.Random(minLength, maxLength))
			}
			if err != nil {
				Elogger.Fatal().Msg(err.Error())
			}
			if len(phrase) > maxLength && !useWords {
				phrase = phrase[:maxLength]
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

				secretStore.SaveV2()
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
	}
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "tool for password phrase generation",
	Long: ` This subcommand helps generate russians words based passphrase and optionally store it in vault
			For example: mcli secrets generate --use-words 
`,
	Run: GenerateSecret,
}

func init() {
	secretsCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("dict-path", "d", "", "path to csv word dictionary")
	generateCmd.Flags().StringP("vault-path", "v", GlobalMap["HomeDir"]+"/.mcli/secrets/defvault", "path to vault")
	generateCmd.Flags().StringP("keyfile-path", "k", "", "path to file to get access key")
	generateCmd.Flags().BoolP("use-words", "w", false, "generate by russian words toggle")
	generateCmd.Flags().BoolP("obfuscate", "o", false, "obfuscate phrase by replacing a=@, l=1 and so on")
	generateCmd.Flags().BoolP("translit", "t", false, "translit password phrase from ru to en")
	generateCmd.Flags().BoolP("qwerty", "q", false, "translit password phrase from ru layout to en layout (йцук to qwer)")
	generateCmd.Flags().IntP("min-length", "m", 8, "min length of password (affected then use-words eq false )")
	generateCmd.Flags().IntP("max-length", "x", 24, "max length of password (affected then use-words eq false )")
}
