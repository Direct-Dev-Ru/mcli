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

var ColorReset string = "\033[0m"

var ColorRed string = "\033[31m"
var ColorGreen string = "\033[32m"
var ColorYellow string = "\033[33m"
var ColorBlue string = "\033[34m"
var ColorPurple string = "\033[35m"
var ColorCyan string = "\033[36m"
var ColorWhite string = "\033[37m"

func ToggleColors(showColor bool) {
	if !showColor {
		ColorRed = ""
		ColorGreen = ""
		ColorYellow = ""
		ColorBlue = ""
		ColorPurple = ""
		ColorCyan = ""
		ColorWhite = ""
	} else {
		ColorRed = "\033[31m"
		ColorGreen = "\033[32m"
		ColorYellow = "\033[33m"
		ColorBlue = "\033[34m"
		ColorPurple = "\033[35m"
		ColorCyan = "\033[36m"
		ColorWhite = "\033[37m"

	}
}

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
