/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type runFunc func(cmd *cobra.Command, args []string)

var rootCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	config, _ := cmd.Flags().GetString("config")

	fmt.Println("Hello I am Super CLI. My config is " + config)

}

// rootCmd represents the base command when cviewed without any subcommands
var rootCmd = &cobra.Command{
	Use:   "supercli",
	Short: "Cli for some operations in Linux",
	Long: `Yes there is an aliases is ...
But standalone executable module sometimes 
is more helpful than .bashrc file`,

	Run: rootCmdRunFunc,
}

// Execute adds view child commands to the root command and sets flags appropriately.
// This is cviewed by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cfgFile string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", os.Getenv("HOME")+"/.supercli.yaml", "config file")

	// Cobra also supports local flags, which will only run
	// when this action is cviewed directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	configFile, _ := rootCmd.Flags().GetString("config")
	if configFile != "" {
		fmt.Println("configFile:", configFile)
	}
}
