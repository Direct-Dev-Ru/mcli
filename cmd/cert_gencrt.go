/*
Copyright © 2024 ANTON K. <info@direct-dev.ru>
*/
package cmd

import (
	"os"
	"strings"
	"time"

	mcli_crypto "mcli/packages/mcli-crypto"

	"github.com/spf13/cobra"
)

type StoreFormatString struct {
	ValueType string
	Value     string
	TimeStamp time.Time
}

func genCrtRun(cmd *cobra.Command, args []string) {
	crtStorePath, _ := GetStringParam("crt-store-path", cmd, os.Getenv("MCLI_CRT_STORE_PATH"))
	caCrtPath, _ := ProcessCommandParameter("ca-path", os.Getenv("MCLI_CRT_CA_PATH"), cmd)

	// orgName, country, location string, DNSNames []string, IPAddresses []net.IP

	orgName, _ := GetStringParam("org-name", cmd, os.Getenv("MCLI_CRT_ORG_NAME"))
	country, _ := GetStringParam("country", cmd, os.Getenv("MCLI_CRT_COUNTRY"))
	location, _ := GetStringParam("location", cmd, os.Getenv("MCLI_CRT_LOCATION"))
	commonName, _ := GetStringParam("common-name", cmd, os.Getenv("MCLI_CRT_CN"))
	DNSNames := []string{commonName}

	var storeInRedis bool = false
	var redisPrefix string
	if strings.HasPrefix(crtStorePath, "redis://") {
		storeInRedis = true
		redisPrefix = strings.ReplaceAll(crtStorePath, "redis://", "")
	}

	if storeInRedis {
		// we need get CA from redis
	}

	certCrt, certKey, err := mcli_crypto.GenerateCertificateWithCASign(crtStorePath, caCrtPath, orgName, country, location, DNSNames, nil)
	if err != nil {
		Elogger.Fatal().Msgf("error generating certificate %v", err)
	}
	_, _ = certCrt, certKey

	if IsRedis && CommonRedisStore != nil {

		err = CommonRedisStore.SetRecord(commonName+".crt", certCrt, redisPrefix)
		if err != nil {
			Ilogger.Trace().Msgf("certificate do not store in redis database: %v", err)
		}
		err = CommonRedisStore.SetRecord(commonName+".key", certKey, "certificates")
		if err != nil {
			Ilogger.Trace().Msgf("certificate do not store in redis database: %v", err)
		}

		// certFromRedis, err, _ := CommonRedisStore.GetRecord(commonName, "certificates")
		// if err != nil {
		// 	Ilogger.Trace().Msgf("can not get certificate from redis database: %v", err)
		// }
		// fmt.Println(string(certFromRedis))
	}

}

// gencrtCmd represents the gencrt command
var gencrtCmd = &cobra.Command{
	Use:   "gencrt",
	Short: "Command to generate rsa certificate on given path, signed by CA.",
	Long: `Command to generate rsa certificate on given path, signed by CA.
	If CA is not provided? it will be  generated in the same path as crt.
	
		--crt-store-path - it can be sets as /opt/cert for example or file://opt/cert or 
			redis://certificates:domain.com - if redis selected as store, then it will make 2 keys
			certificates:domain.com.crt and certificates:domain.key


	`,
	Run: genCrtRun,
}

func init() {
	certCmd.AddCommand(gencrtCmd)
	// gencrtCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	gencrtCmd.Flags().StringP("crt-store-path", "p", ".", "path to save crt and key files")
	gencrtCmd.Flags().StringP("ca-path", "c", ".", "path to ca crt file(key must be near)")
	gencrtCmd.Flags().StringP("common-name", "N", ".", "common domain name")
	gencrtCmd.Flags().StringP("org-name", "O", ".", "organization name")
	gencrtCmd.Flags().StringP("country", "C", ".", "country")
	gencrtCmd.Flags().StringP("location", "L", ".", "location")

}
