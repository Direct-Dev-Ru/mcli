/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var grepCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	// var port, staticPath, staticPrefix string
	// var timeout int64 = 0

	// port, _ = cmd.Flags().GetString("port")
	// isPortSet := cmd.Flags().Lookup("port").Changed

	// Ilogger.Trace().Msg("Args are: " + strings.Join(args, " "))
	// Ilogger.Trace().Msg("\nInput are: " + fmt.Sprintf("%v", Input.inputSlice))

	var inputType, outputType, filter, source string

	var showColor, isRegExp bool = false, false
	showColor, _ = cmd.Flags().GetBool("color")
	ToggleColors(showColor)

	inputType, _ = cmd.Flags().GetString("input-type")
	isInputTypeSet := cmd.Flags().Lookup("input-type").Changed
	if !isInputTypeSet && false { //len(Config.Secrets.Common....) > 0 {
		inputType = "" //Config.Secrets.Common....
	}
	inputTypes := []string{"plain", "json", "table"}
	ok := slices.Contains(inputTypes, inputType)
	if !ok {
		inputType = "plain"
	}

	filter, _ = cmd.Flags().GetString("filter")
	isFilterSet := cmd.Flags().Lookup("").Changed
	if !isFilterSet && false { //len(Config.Secrets.Common....) > 0 {
		filter = "" //Config.Secrets.Common....
	}
	fmt.Println(isRegExp, filter, source)

	outputType, _ = cmd.Flags().GetString("output-type")
	outputType = strings.ToLower(outputType)
	isOutputTypeSet := cmd.Flags().Lookup("output-type").Changed
	if !isOutputTypeSet && false { //len(Config.Secrets.Common....) > 0 {
		outputType = "" //Config.Secrets.Common....
	}

	outTypes := []string{"plain", "json", "table"}
	ok = slices.Contains(outTypes, outputType)
	if !ok {
		outputType = "plain"
	}

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
	Short: "analog of grep command",
	Long: `Linux has grep command - this command can grep input from pipe or file or source can be specified 
		througth param --source. For example: mcli grep --source ./myfile1 ./mydir2 --filter Hello
		--filter is a regular expression if starts with regexp: --filter regexp:^Hello
	`,
	Run: grepCmdRunFunc,
}

func init() {
	rootCmd.AddCommand(grepCmd)

	grepCmd.Flags().StringP("input-type", "i", "plain", "how parse input: as plain or json or table")
	grepCmd.Flags().StringP("output-type", "o", "plain", "how format output: as plain or json or table")
	grepCmd.Flags().StringP("source", "s", "/input", "is input data from pipe - /input value or should be specified througth --source")
	grepCmd.Flags().StringP("dest", "d", "/stdout", "is output data print to stdout /stdout or to file. default - stdout")
	grepCmd.Flags().StringP("filter", "f", "", "filter expression - if starts from regexp: it will be regexp search")
	grepCmd.Flags().BoolP("color", "c", false, "show with colorzzz")
}
