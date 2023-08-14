/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_utils "mcli/packages/mcli-utils"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

func getFilterTokens(filter string) [][]string {
	orFilterMembers := strings.Split(filter, "-or-")
	filterStructure := make([][]string, 0)
	for _, orM := range orFilterMembers {
		currenAndSlice := make([]string, 0)
		currentMember := strings.TrimSpace(orM)
		andFilterMembers := strings.Split(currentMember, "-and-")
		for _, andM := range andFilterMembers {
			currenAndSlice = append(currenAndSlice, strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(andM), "{"), "}"))
		}
		filterStructure = append(filterStructure, currenAndSlice)
	}
	return filterStructure
}

func filterTable(table []map[string]string, tokens [][]string) []map[string]string {
	resTable := make([]map[string]string, 0, 1)

	for _, row := range table {
		if len(tokens) == 0 {
			resTable = append(resTable, row)
			continue
		}
		resOr := false

		for _, orM := range tokens {
			currentAndRes := true
			for _, andM := range orM {
				var re = regexp.MustCompile(`^?(?P<col>[A-Z0-9А-Я]+):(?P<val>.+)`)
				tokenSubMatch := re.FindStringSubmatch(andM)
				if len(tokenSubMatch) > 0 {
					var colName, val string = "", ""
					if index := re.SubexpIndex("col"); index > 0 {
						colName = tokenSubMatch[index]
					}
					if index := re.SubexpIndex("val"); index > 0 {
						val = tokenSubMatch[index]
					}
					if len(val) > 0 && len(colName) > 0 {
						colValue, ok := row[colName]
						if ok {
							valRe := regexp.MustCompile(val)
							currentAndRes = currentAndRes && valRe.MatchString(colValue)
							continue
						}
						currentAndRes = currentAndRes && false
					}
				} else {
					// Column not specified - check all columns -> true if match in any column
					currentRowMatch := false
					for _, col := range row {
						valRe := regexp.MustCompile(andM)
						currentRowMatch = currentRowMatch || valRe.MatchString(col)
					}
					currentAndRes = currentAndRes && currentRowMatch
				}
			}
			resOr = resOr || currentAndRes
		}
		if resOr {
			resTable = append(resTable, row)
		}
	}
	return resTable
}

func filterString(line string, tokens [][]string) bool {
	if len(tokens) == 0 {
		return true
	}
	resOr := false
	for _, orM := range tokens {
		currentAndRes := true
		for _, andM := range orM {
			valRe := regexp.MustCompile(andM)
			currentMatch := valRe.MatchString(line)
			currentAndRes = currentAndRes && currentMatch
			if !currentAndRes {
				break
			}
		}
		resOr = resOr || currentAndRes
	}
	return resOr
}

type outWalker struct {
	filepath    string
	lineContent string
	lineNumber  int
}

func ProcessOneInputFile(filepath string, filterTokens [][]string, fnameFilterTokens [][]string, fs os.FileInfo) []outWalker {
	result := make([]outWalker, 0, 10)
	fmt.Println(filepath, filterTokens)

	file, err := os.Open(filepath)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currentLine string
	var lineNumber int = 1
	for scanner.Scan() {
		currentLine = scanner.Text()
		if filterString(currentLine, filterTokens) {
			result = append(result, outWalker{filepath: filepath, lineContent: currentLine, lineNumber: lineNumber})
		}
		lineNumber++
	}

	return result
}

func ProcessSourceParameter(source string) []string {
	pathCandidates := make([]string, 0, 5)
	pathMembers := make([]string, 0, len(pathCandidates))

	for _, path := range strings.Fields(source) {
		if path = strings.TrimSpace(path); len(path) > 0 {
			pathCandidates = append(pathCandidates, path)
		}
	}
	sort.Strings(pathCandidates)

	for k, p := range pathCandidates {
		var prev string
		if k > 0 {
			prev = pathCandidates[k-1]
		} else {
			pathMembers = append(pathMembers, p)
		}
		if !strings.Contains(p, prev) {
			pathMembers = append(pathMembers, p)
		}
	}
	return pathMembers
}

func ListSourceByWalk(source, filter string) (result []outWalker, err error) {

	paths := ProcessSourceParameter(source)
	filterTokens := getFilterTokens(filter)
	//  TODO: implement filering by filename and dirname
	// fileNameFilterTokens := getFilterTokens(fnameFilter)

	resultCh := make(chan []outWalker, 100)
	// loop through paths and run goroutines
	for _, path := range paths {

		fs, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		// path is file
		if !fs.IsDir() {
			WgGlb.Add(1)
			go func(path, filter string) {
				defer WgGlb.Done()
				resultCh <- ProcessOneInputFile(path, filterTokens, nil, fs)
			}(path, filter)
		} else {
			// path is Dir
			filepath.Walk(path, func(wPath string, info os.FileInfo, err error) error {
				// if the same path
				if wPath == path {
					return nil
				}
				// If current path is Dir - do nothing
				if info.IsDir() {
					_ = fmt.Sprintf("[%s]\n", wPath)
				}
				// if we got file, we take its full path and
				if wPath != path && !info.IsDir() {
					fullFilePath := wPath

					WgGlb.Add(1)
					go func(path, filter string) {
						defer WgGlb.Done()
						resultCh <- ProcessOneInputFile(path, filterTokens, nil, fs)
					}(fullFilePath, filter)
				}
				return nil
			})
		}
	}
	// waits for all goroutines to finish
	go func() {
		WgGlb.Wait()
		close(resultCh)
	}()

	for v := range resultCh {
		result = append(result, v...)
	}

	return result, nil
}

var grepCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	var inputType, outputType, filter, source, dest, out_cols string
	var showColor, isNoHeaders, isShowInputFromPipe bool = false, false, false

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

	out_cols, _ = cmd.Flags().GetString("out-cols")
	out_cols = strings.ToUpper(out_cols)
	outColumns := strings.Split(out_cols, ",")
	isNoHeaders, _ = cmd.Flags().GetBool("no-headers")
	isShowInputFromPipe, _ = cmd.Flags().GetBool("viewpipe")

	Ilogger.Trace().MsgFunc(func() string {
		return fmt.Sprintf("%#v %#v %#v", filter, source, dest)
	})

	OutPut := OutputData{OutputSlice: make([]string, 0, 10)}

	// Process if input passed througth pipe
	var InputForGrep InputData

	if IsCommandInPipe() {
		if v, e := Input.DivideInputSlice("||", ' '); e != nil {
			InputForGrep = InputData{InputSlice: Input.InputSlice,
				InputMap:   make(map[string][]string),
				InputTable: make([]map[string]string, 0),
			}
		} else {
			InputForGrep = InputData{InputSlice: v,
				InputMap:   make(map[string][]string),
				InputTable: make([]map[string]string, 0),
			}
		}
		if isShowInputFromPipe {
			fmt.Println("----------------start input from pipe passed to command--------------------")
			fmt.Println("")
			for _, line := range Input.InputSlice {
				fmt.Print(line)
			}
			fmt.Println("")
			fmt.Println("----------------end input from pipe passed to command--------------------")
		}
		// fmt.Println(InputForGrep.InputSlice)

		if len(InputForGrep.InputSlice) > 0 {
			switch inputType {
			case "table":
				var headers []string = make([]string, 0, 5)
				var headersPositions []int = make([]int, 0, 5)
				// var headersBorders []border = make([]border, 0, 5)
				var isHeadersSet bool = false
				for _, inputLine := range InputForGrep.InputSlice {
					currentLine := strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], "")
					if len(currentLine) == 0 {
						continue
					}
					// splits by two or more spaces or one or more tabs
					// splitRX := regexp.MustCompile(`([ ]{2,})|([\t]{1,})`)
					splitRX := regexp.MustCompile(`([|]{2,})|([\t]{1,})`)

					if !isHeadersSet {
						hs := splitRX.Split(currentLine, -1)
						//fmt.Println(hs)
						for _, h := range hs {
							InputForGrep.InputMap[h] = make([]string, 0, len(InputForGrep.InputSlice)-1)
							headers = append(headers, strings.TrimSpace(strings.ToUpper(h)))
							headersPositions = append(headersPositions, strings.Index(currentLine, h))
						}
						isHeadersSet = true
						continue
					}
					var row []string

					row = splitRX.Split(currentLine, -1)
					for i, f := range row {
						row[i] = strings.TrimSpace(f)
					}
					if len(row) != len(headers) {
						row = mcli_utils.SliceStringByPositions(currentLine, headersPositions)
					}

					currentRowMap := make(map[string]string, 0)

					for k, h := range headers {
						if k < len(row) && len(h) > 0 {
							InputForGrep.InputMap[h] = append(InputForGrep.InputMap[h], row[k])
							currentRowMap[h] = row[k]
						}
					}

					InputForGrep.InputTable = append(InputForGrep.InputTable, currentRowMap)

				}
			case "json":
			case "plain":
				for _, inputLine := range InputForGrep.InputSlice {
					currentLine := strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], "")
					if len(currentLine) == 0 {
						continue
					}
					hashSum := hex.EncodeToString(mcli_crypto.SHA_256(currentLine))
					InputForGrep.InputMap[hashSum] = []string{currentLine}
				}
				for k, v := range InputForGrep.InputMap {
					fmt.Printf("%s: %v\n", k, v)
				}
			}
		}
	} else {
		// read input data given through parameters ( files or dirs )
		// InputForGrep.InputTable = append(InputForGrep.InputTable, map[string]string{"k": "kkk"})

		switch inputType {
		case "plain":
			if source == "/input" {
				Elogger.Fatal().Msg("source (paths to file(s) or dir(s)) does not specified")
			}
			result, err := ListSourceByWalk(source, filter)
			if err != nil {
				Elogger.Error().Msgf("while processing [%s] got an error:  %v", source, err.Error())
			}
			for _, v := range result {
				fmt.Println(v)
			}
		}
	}

	// Process filtering of data - if plain then filtering was maid during reading of files
	switch inputType {
	case "table":
		OutPut.OutputTable = filterTable(InputForGrep.InputTable, getFilterTokens(filter))
		outJson, _ := mcli_utils.PrettyJsonEncodeToString(OutPut.OutputTable)
		fmt.Println("outJson :", outJson)

	case "json":
	default:
	}

	// Output results

	printFunction := func(header, value string) {
		fmt.Printf("%s:%s\n", header, value)
	}
	if isNoHeaders {
		printFunction = func(header, value string) {
			fmt.Printf("%s\n", value)
		}
	}

	switch {
	case outputType == "plai n":
		if len(outColumns) > 0 {
			for k, v := range InputForGrep.InputMap {
				if slices.Contains(outColumns, k) {
					for _, v2 := range v {
						printFunction(k, v2)
					}
				}
			}
		}
	}

}

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "analog of grep command",
	Long: `Linux has grep command. This command is analog - it can get input from pipe or file or source can be specified 
		througth param --source. For example: mcli grep --source ./myfile1 ./mydir2 --filter Hello
		--filter is a regular expression if starts with regexp: --filter regexp:^Hello
	`,
	Run: grepCmdRunFunc,
}

func init() {
	rootCmd.AddCommand(grepCmd)

	grepCmd.Flags().StringP("input-type", "i", "plain", "how parse input: as plain or json or table")
	grepCmd.Flags().BoolP("viewpipe", "v", false, "set if you want see input pipe passed to command")
	grepCmd.Flags().StringP("output-type", "o", "plain", "how format output: as plain or json or table")
	grepCmd.Flags().StringP("out-cols", "l", "", "output columns when table outputs: if omits  - all columns printed")
	grepCmd.Flags().BoolP("no-headers", "n", false, "then outputs as table omit headers or not")
	grepCmd.Flags().StringP("source", "s", "/input", "is input data from pipe - /input value or should be specified througth --source")
	grepCmd.Flags().StringP("dest", "d", "/stdout", "is output data print to stdout /stdout or to file. default - stdout")
	grepCmd.Flags().StringP("filter", "f", "", "filter expression - if starts from regexp: it will be regexp search")
	grepCmd.Flags().BoolP("color", "c", false, "show with colorzzz")
}
