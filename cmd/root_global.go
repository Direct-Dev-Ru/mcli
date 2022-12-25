package cmd

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type TerminalParams struct {
	TermWidth  int
	TermHeight int
	IsTerminal bool
}

type InputData struct {
	// slice of input lines separatated by \n
	InputSlice []string
	// map then input is a table
	InputMap   map[string][]string
	InputTable []map[string]string
}

type OutputData struct {
	OutputSlice []string
	OutputMap   map[string]interface{}
	OutputTable []map[string]string
}

func (d InputData) GetJoinedString(strJoin string, removeLineBreaks bool) (string, error) {
	joinedInput := strings.Join(d.InputSlice, strJoin)

	if removeLineBreaks {
		joinedInput = strings.ReplaceAll(joinedInput, "\r\n", "")
		joinedInput = strings.ReplaceAll(joinedInput, "\n", "")
	}
	return joinedInput, nil
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
var TermWidth, TermHeight int = 0, 0
var IsTerminal bool = false
var OS string
var WgGlb sync.WaitGroup
var ConfigPath string
var RootPath string

var GlobalMap map[string]string = make(map[string]string)

var Version string = "0.1.0"
var Input InputData = InputData{InputSlice: []string{},
	InputMap:   make(map[string][]string),
	InputTable: make([]map[string]string, 0),
}

// https://habr.com/ru/company/macloud/blog/558316/
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

func StandartPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func IsCommandInPipe() bool {
	return GlobalMap["IS_COMMAND_IN_PIPE"] == "CommandInPipe"
}

func GetRootAndDefaultConfigPaths() (configPath string, rootPath string, err error) {
	var execPath string
	execPath, err = os.Executable()
	if err != nil {
		return "", "", err
	}

	execDirPath := filepath.Dir(execPath)
	rootPath = execDirPath
	var configPathCandidate string
	configPath = ""
	// if runs as script e.g.: go run .
	if strings.Contains(execPath, "go-build") {
		_, callerPath, _, _ := runtime.Caller(0)
		rootPath = path.Dir(path.Dir(callerPath))
		configPathCandidate = rootPath + "/.mcli.yaml"
		_, err = os.Stat(configPathCandidate)
		if err == nil && len(configPath) == 0 {
			configPath = configPathCandidate
		}
	}

	// check if the .mcli.yaml file is in the root dir, in the same dir as executable file
	configPathCandidate = execDirPath + "/.mcli.yaml"
	_, err = os.Stat(configPathCandidate)

	if err == nil && len(configPath) == 0 {
		configPath = configPathCandidate
		return configPath, rootPath, nil
	}

	OS = runtime.GOOS
	var homeDir string
	switch OS {
	case "windows":
		homeDir = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	default:
		homeDir = os.Getenv("HOME")
	}
	// check if the .mcli.yaml file is in the home dir
	configPathCandidate = homeDir + "/.mcli/config/.mcli.yaml"

	_, err = os.Stat(configPathCandidate)
	if err == nil && len(configPath) == 0 {
		configPath = configPathCandidate
		return configPath, rootPath, nil
	}

	return "", rootPath, nil

}

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
	GlobalMap["DefaultConfigPath"] = StandartPath(cPath)

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "specify path to config file *.yaml")

	rootCmd.Flags().StringP("root-args", "a", "", "args for root command")

}
