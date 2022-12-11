/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"fmt"
	mcli_utils "mcli/packages/mcli-utils"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var grepCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	var inputType, outputType, filter, source, dest string

	var showColor, isRegExp bool = false, false
	showColor, _ = cmd.Flags().GetBool("color")
	ToggleColors(showColor)

	inputType, _ = cmd.Flags().GetString("input-type")
	isInputTypeSet := cmd.Flags().Lookup("input-type").Changed
	if !isInputTypeSet && false { //len(Config.Grep. ...) > 0 {
		inputType = "" //Config.Grep. ...
	}
	inputTypes := []string{"plain", "json", "json-pretty", "table"}
	ok := slices.Contains(inputTypes, inputType)
	if !ok {
		inputType = "plain"
	}

	filter, _ = cmd.Flags().GetString("filter")
	isFilterSet := cmd.Flags().Lookup("filter").Changed
	if !isFilterSet && false { //len(Config.Grep. ...) > 0 {
		filter = "" //Config.Grep. ...
	}

	source, _ = cmd.Flags().GetString("source")
	isSourceSet := cmd.Flags().Lookup("source").Changed
	if !isSourceSet && false { //len(Config.Grep. ...) > 0 {
		source = "" //Config.Grep. ...
	}

	dest, _ = cmd.Flags().GetString("dest")
	isDestSet := cmd.Flags().Lookup("dest").Changed
	if !isDestSet && false { //len(Config.Grep. ...) > 0 {
		dest = "" //Config.Grep. ...
	}

	outputType, _ = cmd.Flags().GetString("output-type")
	outputType = strings.ToLower(outputType)
	isOutputTypeSet := cmd.Flags().Lookup("output-type").Changed
	if !isOutputTypeSet && false { //len(Config.Grep. ...) > 0 {
		outputType = "" //Config.Grep. ...
	}
	outTypes := []string{"plain", "json", "json-pretty", "table"}
	ok = slices.Contains(outTypes, outputType)
	if !ok {
		outputType = "plain"
	}
	Ilogger.Trace().MsgFunc(func() string {
		return fmt.Sprintf("%v %v %v %v", isRegExp, filter, source, dest)
	})

	// Process if input througth pipe entered
	if IsCommandInPipe() {

		if len(Input.InputSlice) > 0 {
			var headers []string = make([]string, 0, 5)
			var isHeadersSet bool = false
			for _, inputLine := range Input.InputSlice {

				currentLine := strings.TrimSpace(strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], ""))
				if len(currentLine) == 0 {
					continue
				}
				// check for input type
				switch inputType {
				case "table":

					// splits by two or more spaces or one ore more tabs
					splitRX := regexp.MustCompile(`([ ]{2,})|([\t]{1,})`)
					if !isHeadersSet {
						hs := splitRX.Split(currentLine, -1)
						for _, h := range hs {
							Input.InputMap[h] = make([]string, 0, len(Input.InputSlice)-1)
							headers = append(headers, h)
						}
						isHeadersSet = true
						continue
					}
					row := splitRX.Split(currentLine, -1)
					currentRowMap := make(map[string]string, 0)
					for k, h := range headers {
						Input.InputMap[h] = append(Input.InputMap[h], row[k])
						currentRowMap[h] = row[k]
					}
					Input.InputTableSlice = append(Input.InputTableSlice, currentRowMap)
				case "json":
				default:

				}

			}
			// for k, v := range Input.InputMap {
			// 	fmt.Println(k, v)
			// }

			outString, _ := mcli_utils.PrettyJsonEncodeToString(Input.InputTableSlice)
			fmt.Println(outString)
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
