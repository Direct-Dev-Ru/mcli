/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	mcli_utils "mcli/packages/mcli-utils"
)

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "top level command for secrets managment",
	Long: `Top level command for secrets managment.
	View subcommnds to get more info ([binaryname] secrets --help)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Secrets config:")
		fmt.Println(mcli_utils.PrettyJsonEncodeToString(Config.Secrets))
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
}
