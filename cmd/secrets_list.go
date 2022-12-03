/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_utils "mcli/packages/mcli-utils"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Subcommand for view secrets in a vault",
	Long: `	This command lists stored secrets ia a vault. It can be filtered and displayed either in plain output or in
			or json or table view.
			For example: mcli secrets list -f "{name:secret-001}" -o plain
			mcli secrets list secret-001 -o json
`,
	Run: func(cmd *cobra.Command, args []string) {
		// var LineBreak string = GlobalMap["LineBreak"]
		// fmt.Println(args)
		var vaultPath, keyFilePath, outputType string
		var showSecret, showColor, onlySecret bool = false, false, false
		showSecret, _ = cmd.Flags().GetBool("show-secret")
		onlySecret, _ = cmd.Flags().GetBool("secret-only")
		showColor, _ = cmd.Flags().GetBool("color")

		ToggleColors(showColor)

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
		outputType, _ = cmd.Flags().GetString("output")
		outputType = strings.ToLower(outputType)
		outTypes := []string{"plain", "json", "table"}
		ok := slices.Contains(outTypes, outputType)
		if !ok {
			outputType = "plain"
		}

		secretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile,
			mcli_crypto.AesCypher, nil)

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
		// Ilogger.Trace().Msg("storeContent: " + fmt.Sprintf("%v", secretStore.Secrets))

		switch {
		case outputType == "table":
			// TODO: well yes this cod is ugly (((
			headerTitles := []string{"Name:", "Login:", "Secret:", "Description:", "CreatedAt:"}
			columnCount := len(headerTitles)
			maxLengths := make([]int, 0, columnCount)
			for _, header := range headerTitles {
				maxLengths = append(maxLengths, len(header))
			}

			// check max length of runes values

			for i := 0; i < columnCount; i++ {
				for _, s := range secretStore.Secrets {
					var value []rune
					switch i {
					case 0:
						value = []rune(s.Name)
					case 1:
						value = []rune(s.Login)
					case 2:
						value = []rune(s.Secret)
					case 3:
						value = []rune(s.Description)
					case 4:
						value = []rune(s.CreatedAt.Format("2006-01-02"))
					default:
						value = []rune(s.Name)
					}
					if maxLengths[i] < len(value) {
						maxLengths[i] = len(value)
					}
				}
			}
			stringToPrint := ""

			// printing headers
			var numSpaces int = 3
			for i, v := range headerTitles {
				stringToPrint += v + strings.Repeat(" ", maxLengths[i]-len(v)+numSpaces) + "\t"
			}
			fmt.Println(ColorPurple + stringToPrint + ColorReset)

			for i := 0; i < len(secretStore.Secrets); i++ {
				s := secretStore.Secrets[i]
				stringToPrint := ""
				v := []rune(s.Name)
				stringToPrint += string(v) + strings.Repeat(" ", maxLengths[0]-len(v)+numSpaces) + "\t"
				v = []rune(s.Login)
				stringToPrint += string(v) + strings.Repeat(" ", maxLengths[1]-len(v)+numSpaces) + "\t"
				v = []rune(s.Secret)
				if showSecret {
					stringToPrint += string(v) + strings.Repeat(" ", maxLengths[2]-len(v)+numSpaces) + "\t"
				} else {
					stringToPrint += strings.Repeat("*", maxLengths[2]-len(v)+numSpaces) + strings.Repeat(" ", len(v)) + "\t"
				}
				v = []rune(s.Description)

				stringToPrint += string(v) + strings.Repeat(" ", maxLengths[3]-len(v)+numSpaces) + "\t"
				v = []rune(s.CreatedAt.Format("2006-01-02"))
				stringToPrint += string(v) + strings.Repeat(" ", maxLengths[4]-len(v)+numSpaces) + "\t"

				fmt.Println(ColorCyan + stringToPrint + ColorReset)
			}

		case outputType == "plain":
			for _, v := range secretStore.Secrets {
				if onlySecret {
					fmt.Println(v.Secret)
					continue
				}

				if showSecret {
					fmt.Println(v.Login + ":" + v.Secret)
				} else {
					fmt.Println(v.Login)
				}

			}
		case outputType == "json":
			var copySecrets []mcli_secrets.SecretEntry = make([]mcli_secrets.SecretEntry, len(secretStore.Secrets))
			copy(copySecrets, secretStore.Secrets)
			for i := range copySecrets {
				if !showSecret {
					copySecrets[i].Secret = "***"
				}
			}
			outString, err := mcli_utils.PrettyJsonEncodeToString(copySecrets)
			if err != nil {
				Elogger.Fatal().Msg(err.Error())
			}
			fmt.Println(ColorGreen + outString + ColorReset)
		default:
			reason := "output format not defined"
			Elogger.Fatal().Msg("undefined error: " + fmt.Sprintf("reason: %s", reason))
		}

	},
}

func init() {
	secretsCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("vault-path", "v", GlobalMap["HomeDir"]+"/.mcli/secrets/defvault", "path to vault")
	listCmd.Flags().StringP("keyfile-path", "k", "", "path to file to get access key")
	listCmd.Flags().StringP("output", "o", "table", "output format (default - table, optional - json, plain)")
	listCmd.Flags().BoolP("show-secret", "s", false, "show secrets or not")
	listCmd.Flags().BoolP("color", "c", false, "show with colorzzz")
	listCmd.Flags().BoolP("secret-only", "p", false, "show only secret then plain output - usefull to store into variable: VAR1=$(mcli secrets list secret-001 -p)")
}
