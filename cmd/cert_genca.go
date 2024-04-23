/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_secrets "mcli/packages/mcli-secrets"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// gencaCmd represents the genca command
var gencaCmd = &cobra.Command{
	Use:   "genca",
	Short: "Command to generate CA",
	Long: `This command made for CA (.crt and .key) files generate. 
	It can take path to store generated files and some certificate parameters:
		- Org
		- Country
		- Location
	`,
	Run: func(cmd *cobra.Command, args []string) {
		caStorePath, _ := GetStringParam("crt-store-path", cmd, os.Getenv("MCLI_CA_STORE_PATH"))

		orgName, _ := GetStringParam("org-name", cmd, os.Getenv("MCLI_CA_ORG_NAME"))
		country, _ := GetStringParam("country", cmd, os.Getenv("MCLI_CA_COUNTRY"))
		location, _ := GetStringParam("location", cmd, os.Getenv("MCLI_CA_LOCATION"))
		commonName, _ := GetStringParam("common-name", cmd, os.Getenv("MCLI_CA_CN"))

		envEncrypt, _ := strconv.ParseBool(os.Getenv("MCLI_CA_ENCRYPT_KEY"))
		encrypt, _ := GetBoolParam("encrypt", cmd, envEncrypt)

		caCrt, caKey, err := mcli_crypto.GenerateCACertificateV2(commonName, orgName, country, location)
		if err != nil {
			Elogger.Fatal().Msgf("error generating ca certificate %v", err)
		}

		storeInRedis := false
		redisEncKey := ""
		_, _, _, _ = storeInRedis, caCrt, caKey, redisEncKey

		if strings.HasPrefix(caStorePath, "redis://") {
			storeInRedis = true

			if CommonRedisStore == nil {
				Elogger.Fatal().Msgf("storing in redis is not possible: CommonRedisStore is nil")
				return
			}

			// getting encryption redis key
			internalSecretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile, exportCypher, nil)
			if err := internalSecretStore.FillStore(Config.Common.InternalVaultPath, Config.Common.InternalKeyFilePath); err != nil {
				if encrypt {
					Elogger.Fatal().Msgf("error fills secret store %v", err)
				}
			}
			redisEncKeySecret, ok := internalSecretStore.GetSecretPlainMap()["RedisEncKey"]
			if ok {
				redisEncKey = redisEncKeySecret.Secret
			} else {
				if encrypt {
					Elogger.Fatal().Msgf("error getting encryption key")
				}
			}
			CommonRedisStore.SetEncrypt(encrypt, []byte(redisEncKey), cypher)
		}

	},
}

func init() {
	certCmd.AddCommand(gencaCmd)

	// gencaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	gencaCmd.Flags().StringP("ca-store-path", "p", "", "path to save ca and key files (may be redis://...)")
	gencaCmd.Flags().StringP("common-name", "N", "", "CN of CA certificate")
	gencaCmd.Flags().StringP("org-name", "O", "", "organization name")
	gencaCmd.Flags().StringP("country", "C", "", "country")
	gencaCmd.Flags().StringP("location", "L", "", "location")
	gencaCmd.Flags().BoolP("encrypt", "e", false, "encrypt private key file/record")
}
