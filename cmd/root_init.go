package cmd

import (
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	cobra.OnInitialize(initConfig)

	// define Terminal Params
	GlobalMap["IsTerminal"] = "false"

	if term.IsTerminal(0) {
		// println("in a term")
		IsTerminal = true
		GlobalMap["IsTerminal"] = "true"
	}
	var err error
	TermWidth, TermHeight, err = term.GetSize(0)
	GlobalMap["TermWidth"] = strconv.Itoa(TermWidth)
	GlobalMap["TermHeight"] = strconv.Itoa(TermHeight)

	if err != nil {
		IsTerminal = false
		GlobalMap["IsTerminal"] = "false"
	}
	// println("width:", TermWidth, "height:", TermHeight)

	OS = runtime.GOOS
	GlobalMap["OS"] = OS

	switch OS {
	case "windows":
		GlobalMap["HomeDir"] = StandartPath(os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH"))
		GlobalMap["LineBreak"] = "\r\n"
	case "darwin":
		GlobalMap["LineBreak"] = "\n"
		GlobalMap["HomeDir"] = StandartPath(os.Getenv("HOME"))
	case "linux":
		GlobalMap["HomeDir"] = StandartPath(os.Getenv("HOME"))
		GlobalMap["LineBreak"] = "\n"
	default:
		GlobalMap["LineBreak"] = "\n"
		GlobalMap["HomeDir"] = StandartPath(os.Getenv("HOME"))
	}

	cPath, rPath, _ := GetRootAndDefaultConfigPaths()

	GlobalMap["RootPath"] = StandartPath(rPath)
	os.Setenv("RootPath", GlobalMap["RootPath"])
	GlobalMap["DefaultConfigPath"] = StandartPath(cPath)
	os.Setenv("DefaultConfigPath", GlobalMap["DefaultConfigPath"])

	// println(GlobalMap["RootPath"])
	// println(GlobalMap["DefaultConfigPath"])

	// generate rootkey for internal secrets
	// rootSecretStorePath := filepath.Join(GlobalMap["HomeDir"], ".mcli", "root", "secret")
	// _, _, err = mcli_utils.IsExistsAndCreate(rootSecretStorePath, true)
	// if err != nil {
	// 	log.Fatalln("root secret store error - path do not exists: ", err)
	// }
	// rootSecretStore_key := filepath.Join(rootSecretStorePath, "rootkey.key")
	// ok, _, _ := mcli_utils.IsExistsAndCreate(rootSecretStore_key, false)
	// if !ok {
	// 	err = mcli_secrets.SaveKeyToFilePlain(rootSecretStore_key, mcli_secrets.GenKey(1024))
	// 	if err != nil {
	// 		log.Fatalln("root secret store error - save rootSecretStore_key error: ", err)
	// 	}
	// }
	// _, err = mcli_secrets.LoadKeyFromFilePlain(rootSecretStore_key)
	// if err != nil {
	// 	log.Fatalln("root secret store error - load rootSecretStore_key error: ", err)
	// }
	// GlobalMap["RootSecretKeyPath"] = rootSecretStore_key

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", cPath, "specify path to config file *.yaml")
	rootCmd.PersistentFlags().StringVar(&InputDataFromFile, "stdin-from-file", "", "specify path to file emulating data for stdin")
	rootCmd.PersistentFlags().StringVar(&RedisHost, "redis-host", "127.0.0.1", "specify redis host")
	rootCmd.PersistentFlags().StringVar(&RedisPort, "redis-port", "6379", "specify redis port")
	rootCmd.PersistentFlags().StringVar(&RedisPwd, "redis-password", "", "specify redis Pa$$w0rd")
	rootCmd.PersistentFlags().BoolVarP(&IsRedis, "is-redis", "", false, "specify connect to common redis database or not")
	rootCmd.PersistentFlags().BoolVarP(&IsVerbose, "verbose", "V", false, "specify verbose toggle")

	rootCmd.Flags().StringP("root-args", "a", "", "args for root command")

}
