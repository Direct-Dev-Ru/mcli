/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "just test command for debuging",
	Long:  `useful only when DEBUG variable set to true`,
	Run: func(cmd *cobra.Command, args []string) {
		ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
		if ENV_DEBUG != "true" {
			os.Exit(1)
		}
		// refString := "{{$RootPath$}}/pwdgen/{{$HOME}}/freqrnc2011.csv"
		// templateRegExp := regexp.MustCompile(`{{\$.+?}}`)

		// all := templateRegExp.FindAllString(refString, -1)
		// fmt.Println("All: ")
		// for _, val := range all {
		// 	fmt.Println(val)
		// }

		cmdSh := &exec.Cmd{
			Path:   "./script.sh",
			Args:   []string{"./script.sh"},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		// run command
		if err := cmdSh.Run(); err != nil {
			fmt.Println("Error:", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
