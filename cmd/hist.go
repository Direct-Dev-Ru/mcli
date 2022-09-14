/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var histCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	// config, _ := cmd.Flags().GetString("config")

	fmt.Println("Hello I am Super CLI. hist command.")
	fmt.Println("Args are:", args)

}

// histCmd represents the hist command
var histCmd = &cobra.Command{
	Use:   "hist",
	Short: "lists your bash command history",
	Long: `Bash has history file
this command can list view its saved commands to screen. For example:
supercli hist view
supercli hist --filter docker`,
	// Run: histCmdRunFunc,
}

func init() {
	rootCmd.AddCommand(histCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and view subcommands, e.g.:
	// histCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is cviewed directly, e.g.:
	// histCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
