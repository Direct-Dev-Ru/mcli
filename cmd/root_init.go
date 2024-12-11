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

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", cPath, "specify path to config yaml file")

	rootCmd.PersistentFlags().StringVar(&InputDataFromFile, "stdin-from-file", "", "specify path to file emulating data for stdin")
	rootCmd.PersistentFlags().StringVar(&RedisHost, "redis-host", "127.0.0.1", "specify redis host")
	rootCmd.PersistentFlags().IntVar(&RedisDb, "redis-db", 1, "specify redis database number")
	rootCmd.PersistentFlags().StringVar(&RedisPort, "redis-port", "6379", "specify redis port")
	rootCmd.PersistentFlags().StringVar(&RedisPwd, "redis-password", "", "specify redis Pa$$w0rd")
	rootCmd.PersistentFlags().BoolVarP(&IsRedis, "is-redis", "", false, "specify connect to common redis database or not")
	rootCmd.PersistentFlags().BoolVarP(&IsVerbose, "verbose", "V", false, "specify verbose toggle")

	rootCmd.Flags().StringP("root-args", "a", "", "args for root command")

}
