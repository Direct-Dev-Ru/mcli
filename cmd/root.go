/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var Ilogger, Elogger zerolog.Logger
var cfgFile string

type InputData struct {
	inputSlice  []string
	joinedInput string
}

var Input InputData = InputData{inputSlice: []string{}, joinedInput: ""}

type runFunc func(cmd *cobra.Command, args []string)

var rootCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	config, _ := cmd.Flags().GetString("config")

	fmt.Println("Hello from Multy CLI. Config is " + config)

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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", os.Getenv("HOME")+"/.scli.yaml", "specify config file - default $HOME/.scli.yaml")

	// Cobra also supports local flags, which will only run
	// when this action is cviewed directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	// configFile, _ := rootCmd.Flags().GetString("config")
	// if configFile != "" {
	// 	fmt.Println("configFile:", configFile)
	// }

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
