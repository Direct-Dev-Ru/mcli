/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// gencaCmd represents the genca command
var gencaCmd = &cobra.Command{
	Use:   "genca",
	Short: "Command to generate CA",
	Long: `This command made for CA (.crt and .key) files generate. 
	It can take path to store generated files and some certificate parameters:
		- Org
		- Country
		- Location
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("genca called")
	},
}

func init() {
	certCmd.AddCommand(gencaCmd)

	// gencaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	certCmd.Flags().StringP("store-path", "p", ".", "path to save crt and key files")
	certCmd.Flags().StringP("file-name", "n", "ca", "name of CA certificate and key without extension")
}
