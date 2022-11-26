package cmd

import (
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type InputData struct {
	inputSlice  []string
	joinedInput string
}

type runFunc func(cmd *cobra.Command, args []string)

type SomeData struct {
	payload int
	// err     error
}
type AccumData struct {
	sync.Mutex
	data map[string]SomeData
}

// Global Vars
var Ilogger, Elogger zerolog.Logger
var OS string
var ConfigPath string
var RootPath string
var GlobalMap map[string]string = make(map[string]string)

var Version string = "1.0.9"
var Input InputData = InputData{inputSlice: []string{}, joinedInput: ""}

func init() {
	cobra.OnInitialize(initConfig)

	_, callerPath, _, _ := runtime.Caller(0)
	RootPath = path.Dir(path.Dir(callerPath))
	GlobalMap["RootPath"] = RootPath
	OS = runtime.GOOS
	GlobalMap["OS"] = OS

	switch OS {
	case "windows":
		GlobalMap["HomeDir"] = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		GlobalMap["LineBreak"] = "\r\n"
	case "darwin":
		GlobalMap["LineBreak"] = "\n"
		GlobalMap["HomeDir"] = os.Getenv("HOME")
	case "linux":
		GlobalMap["HomeDir"] = os.Getenv("HOME")
		GlobalMap["LineBreak"] = "\n"
	default:
		GlobalMap["LineBreak"] = "\n"
		GlobalMap["HomeDir"] = os.Getenv("HOME")
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", RootPath+"/.mcli.yaml",
		"specify config file - default "+RootPath+"/.mcli.yaml")

	rootCmd.Flags().StringP("root-args", "a", "", "args for root command")

}
