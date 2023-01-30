/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// utilsCmd represents the utils command
var utilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Set of commands to help make different things such as convert, compress, transform and so on",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("utils commsnd called")
	},
}

func init() {
	rootCmd.AddCommand(utilsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// utilsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// utilsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
