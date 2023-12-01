package cmd

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func ToggleColors(showColor bool) {
	if !showColor {
		ColorRed = ""
		ColorGreen = ""
		ColorYellow = ""
		ColorBlue = ""
		ColorPurple = ""
		ColorCyan = ""
		ColorWhite = ""
		ColorReset = ""
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

func IsCommandNotInPipeInputFromFile() bool {
	return GlobalMap["IS_COMMAND_IN_PIPE"] == "CommandNotInPipeInputFromFile"
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
			return configPath, rootPath, nil
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
	os.Setenv("RootPath", GlobalMap["RootPath"])
	GlobalMap["DefaultConfigPath"] = StandartPath(cPath)
	os.Setenv("DefaultConfigPath", GlobalMap["DefaultConfigPath"])
	// println(GlobalMap["RootPath"])
	// println(GlobalMap["DefaultConfigPath"])

	// // generate rootkey for internal secrets
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

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "specify path to config file *.yaml")
	rootCmd.PersistentFlags().StringVar(&InputDataFromFile, "stdin-from-file", "", "specify path to file emulating data for stdin")
	rootCmd.Flags().StringP("root-args", "a", "", "args for root command")

}
