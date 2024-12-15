/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Working with Dns",
	Long: `
Working with Dns
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dns subtool config:")
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dnsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dnsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
