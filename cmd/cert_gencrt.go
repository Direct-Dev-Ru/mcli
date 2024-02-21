/*
Copyright © 2024 ANTON K. <info@direct-dev.ru>
*/
package cmd

import (
	"os"
	"path/filepath"
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

	var caInRedis, ok bool = false, false
	var doCAGenerate bool = false
	var caRedisPrefix string = "rootCA:"
	var caRedisCrtKey, caRedisPKkey string = "", ""
	var caCrtBytes, caPKBytes []byte
	var err error

	_ = caInRedis
	_ = doCAGenerate
	_ = ok

	if strings.HasPrefix(caCrtPath, "redis://") {
		if CommonRedisStore == nil {
			Elogger.Fatal().Msgf("can not connect to redis database: CommonRedisStore is nil")
			return
		}
		caInRedis = true
		caCrtPath = strings.ReplaceAll(caCrtPath, "redis://", "")
		caParts := strings.Split(caCrtPath, ":")
		if len(caParts) == 1 {
			caRedisCrtKey = caCrtPath
		} else {
			caRedisCrtKey = caParts[len(caParts)-1]
			caRedisPrefix = strings.Join(caParts[:len(caParts)-1], ":")
		}
		if !strings.HasSuffix(caRedisCrtKey, ".crt") {
			caRedisCrtKey = caRedisCrtKey + ".crt"
		}
		caRedisPKkey = strings.ReplaceAll(caRedisCrtKey, ".crt", ".key")

		// get ca from redis
		caCrtBytes, err, ok = CommonRedisStore.GetRecord(caRedisCrtKey, caRedisPrefix)
		if err != nil {
			Elogger.Fatal().Msgf("can not get ca crt from redis database: %v", err)
			return
		}
		_ = ok
		caPKBytes, err, ok = CommonRedisStore.GetRecord(caRedisPKkey, caRedisPrefix)
		if err != nil {
			Elogger.Fatal().Msgf("can not get ca pk from redis database: %v", err)
			return
		}
		_ = ok
	} else {
		// read ca from file system
		caCrtPath = filepath.Clean(caCrtPath)
		caPKPath := strings.ReplaceAll(caCrtPath, ".crt", ".key")

		caCrtBytes, err = os.ReadFile(caCrtPath)
		if err != nil {
			Elogger.Fatal().Msgf("can not read ca crt from file %v %v", caCrtPath, err)
			return
		}
		caPKBytes, err = os.ReadFile(caPKPath)
		if err != nil {
			Elogger.Fatal().Msgf("can not read ca pk from file %v %v", caPKPath, err)
			return
		}
	}

	certCrt, certKey, err := mcli_crypto.GenerateCertificateWithCASignV2(crtStorePath, string(caCrtBytes),
		string(caPKBytes), orgName, country, location, DNSNames, nil)
	if err != nil {
		Elogger.Fatal().Msgf("error generating certificate %v", err)
	}
	_, _ = certCrt, certKey

	var storeInRedis bool = false
	var redisCrtKey, redisCrtPrefix string
	var redisPKKey, redisPKPrefix string
	if strings.HasPrefix(crtStorePath, "redis://") {
		storeInRedis = true

		crtStorePath = strings.ReplaceAll(crtStorePath, "redis://", "")
		redisCrtParts := strings.Split(crtStorePath, ":")
		if len(redisCrtParts) == 1 {
			redisCrtKey = caCrtPath
		} else {
			redisCrtKey = redisCrtParts[len(redisCrtParts)-1]
			redisCrtPrefix = strings.Join(redisCrtParts[:len(redisCrtParts)-1], ":")
			redisPKPrefix = redisCrtPrefix
		}
	} else {
		// store path is file in file system
		crtStorePath = filepath.Clean(crtStorePath)
		redisCrtKey = filepath.Base(crtStorePath)
		redisCrtPrefix = filepath.Dir(crtStorePath)
		redisPKPrefix = filepath.Dir(crtStorePath)
	}
	if !strings.HasSuffix(redisCrtKey, ".crt") {
		redisCrtKey = redisCrtKey + ".crt"
	}
	redisPKKey = strings.ReplaceAll(redisCrtKey, ".crt", ".key")

	_ = storeInRedis

	if storeInRedis {
		if CommonRedisStore == nil {
			Elogger.Fatal().Msgf("can not connect to redis database: CommonRedisStore is nil")
			return
		}
		err = CommonRedisStore.SetRecord(redisCrtKey, certCrt, redisCrtPrefix)
		if err != nil {
			Ilogger.Trace().Msgf("certificate do not stored in redis database: %v", err)
		}
		err = CommonRedisStore.SetRecord(redisPKKey, certKey, redisPKPrefix)
		if err != nil {
			Ilogger.Trace().Msgf("certificate PK do not stored in redis database: %v", err)
		}
	} else {
		err = os.WriteFile(filepath.Join(redisCrtPrefix, redisCrtKey), certCrt, 0644)
		if err != nil {
			Elogger.Fatal().Msgf("can not save certificate to filesystem %v", err)
			return
		}

		err = os.WriteFile(filepath.Join(redisPKPrefix, redisPKKey), certKey, 0644)
		if err != nil {
			Elogger.Fatal().Msgf("can not save certificate's pk to filesystem %v", err)
			return
		}
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
