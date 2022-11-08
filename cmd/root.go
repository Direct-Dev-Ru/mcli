/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type CommonData struct {
	Timeout      int    `yaml:"timeout"`
	OutputFile   string `yaml:"output-file"`
	OutputFormat string `yaml:"output-format"`
}

type ConfigData struct {
	Common struct {
		Timeout      int    `yaml:"timeout"`
		OutputFile   string `yaml:"output-file"`
		OutputFormat string `yaml:"output-format"`
	}

	Http struct {
		Server struct {
			Port         string `yaml:"port"`
			StaticPath   string `yaml:"static-path"`
			StaticPrefix string `yaml:"static-prefix"`
		}
	}
}

type InputData struct {
	inputSlice  []string
	joinedInput string
}

var Ilogger, Elogger zerolog.Logger
var ConfigPath string
var RootPath string
var Input InputData = InputData{inputSlice: []string{}, joinedInput: ""}

// var Config map[string]interface{}
var Config ConfigData = ConfigData{}

type runFunc func(cmd *cobra.Command, args []string)

var rootCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	config, _ := cmd.Flags().GetString("config")
	Ilogger.Info().Msg("Hello from Multy CLI. Config is " + config)
}

// rootCmd represents the base command when cviewed without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcli",
	Short: "cli for some operations in Linux",
	Long: `Yes there is an aliases is ...
But standalone executable module sometimes 
is more helpful than .bashrc file`,

	Run: rootCmdRunFunc,
}

// Execute adds view child commands to the root command and sets flags appropriately.
// This is cviewed by main.main(). It only needs to happen once to the rootCmd.
func Execute(loggers []zerolog.Logger) {
	Ilogger, Elogger = loggers[0], loggers[1]
	// Elogger.Error().Msg("Some Test Error")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cobra.OnInitialize(initConfig)
	_, callerPath, _, _ := runtime.Caller(0)
	RootPath = path.Dir(path.Dir(callerPath))
	// fmt.Fprintln(os.Stdout, RootPath)

	// rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", os.Getenv("HOME")+"/.mcli.yaml", "specify config file - default $HOME/.mcli.yaml")
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", RootPath+"/.mcli.yaml",
		"specify config file - default "+RootPath+"/.mcli.yaml")

	// Cobra also supports local flags, which will only run
	// when this action is cviewed directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	configFile, _ := rootCmd.Flags().GetString("config")
	if configFile != "" {
		Ilogger.Trace().Msg(fmt.Sprint("parsing config file:", configFile))

		if _, err := os.Stat(configFile); err == nil {
			configContent, err := os.ReadFile(configFile)

			if err == nil {
				err = yaml.Unmarshal(configContent, &Config)
			}
			// fmt.Println("Configuration content :", string(configContent))
			Ilogger.Trace().Msg(fmt.Sprint("Configuration struct :", Config))
		} else if errors.Is(err, os.ErrNotExist) {
			Ilogger.Trace().Msg("config file " + configFile + " does not exist")
		} else {
			Elogger.Trace().Msg("config file detect error " + err.Error())
		}
	}

	// Check if piped to StdIn
	info, _ := os.Stdin.Stat()

	if (info.Mode()&os.ModeNamedPipe) == os.ModeNamedPipe || info.Size() > 0 {

		var inputSlice []string = []string{}
		r := bufio.NewReader(os.Stdin)
		for {
			input, err := r.ReadString('\n')
			if input != "" {
				inputSlice = append(inputSlice, input)
			}
			if err != nil && err == io.EOF {
				break
			}
		}

		// fmt.Println(len(inputSlice))
		// for i, v := range inputSlice {
		// 	fmt.Print(i, " ", v)
		// }

		joinedInput := strings.Join(inputSlice, " ")
		joinedInput = strings.ReplaceAll(joinedInput, "\r\n", "")
		joinedInput = strings.ReplaceAll(joinedInput, "\n", "")
		Input.inputSlice = inputSlice
		Input.joinedInput = joinedInput

		// fmt.Println(Input)
	}

	// fmt.Println(info.Mode(), info.Name(), info.Size())
}
