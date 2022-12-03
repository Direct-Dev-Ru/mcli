/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

var grepCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	// var port, staticPath, staticPrefix string
	// var timeout int64 = 0

	// port, _ = cmd.Flags().GetString("port")
	// isPortSet := cmd.Flags().Lookup("port").Changed

	// Ilogger.Trace().Msg("Args are: " + strings.Join(args, " "))
	// Ilogger.Trace().Msg("\nInput are: " + fmt.Sprintf("%v", Input.inputSlice))
	inputMap := make(map[int][]string)
	if len(Input.InputSlice) > 0 {
		for i, inputline := range Input.InputSlice {
			// splits by two
			splitRX := regexp.MustCompile(`([ ]{2,})|([\t]{1,})`)

			inputMap[i] = splitRX.Split(inputline, -1)
			// fmt.Println(len(inputMap[i]))

			for _, v := range inputMap[i] {
				fmt.Println(v)
			}

		}
	}

}

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "grep command analog",
	Long: `Linux has grep command
	this command can grep input from pipe or file. For example:
		mcli grep --file ./myfile --filter docker 
	`,
	Run: grepCmdRunFunc,
}

func init() {
	rootCmd.AddCommand(grepCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and view subcommands, e.g.:
	// grepCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is cviewed directly, e.g.:
	// grepCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	grepCmd.Flags().StringP("input-parse", "i", "plaintext", "how parse input: as plaintext or json or table")
}
