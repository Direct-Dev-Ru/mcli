/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	mcli_error "mcli/packages/mcli-error"
	mcli_utils "mcli/packages/mcli-utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var Err mcli_error.CommonError

// input should be blocks of data if yaml input:
// Key:	KeyName1
// someKey1: SomeValue1
// someKey2: SomeValue2
// ...
// someKeyN: SomeValueN
// ---
// Key:	KeyName2
// someKey1: SomeValue1
// someKey2: SomeValue2
// ...
// someKeyM: SomeValueM
// ---

// or if input is json
// {"Key":	"KeyName1", "someKey1": "SomeValue1", "someKey2": "SomeValue2", ...
// "someKeyM": "SomeValueM"}
// {"Key":	"KeyName2", "someKey1": "SomeValue1", "someKey2": "SomeValue2", ...
// "someKeyN": "SomeValueN"}
func importToRedis(cmd *cobra.Command, args []string) {
	var err error
	redisPwd, _ := cmd.Flags().GetString("dict-path")
	isRedisPwdSet := cmd.Flags().Lookup("redis-pwd").Changed
	if len(redisPwd) == 0 && !isRedisPwdSet {
		redisPwd = os.Getenv("REDIS_PWD")
	}
	_ = redisPwd

	if IsCommandInPipe() && len(Input.InputSlice) > 0 {
		Ilogger.Trace().Msg("input from pipe ...")

		inData := make([]string, 0, len(Input.InputSlice))
		for _, inputLine := range Input.InputSlice {
			currentLine := strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], "")
			currentLine = strings.TrimSpace(currentLine)
			if len(currentLine) == 0 {
				continue
			}
			inData = append(inData, currentLine)
		}

		fmt.Println("input data: ", strings.Join(inData, " "))

		// detect json or not
		a1 := strings.HasPrefix(inData[0], "{")
		a2 := strings.HasSuffix(inData[len(inData)-1], "}")
		b1 := strings.HasPrefix(inData[0], "[")
		b2 := strings.HasSuffix(inData[len(inData)-1], "]")
		var inputRecords map[string]interface{} = make(map[string]interface{}, 0)
		if a1 && a2 {

			dec := json.NewDecoder(strings.NewReader(strings.Join(inData, " ")))
			for dec.More() {
				var entry map[string]interface{}
				err := dec.Decode(&entry)

				if err != nil {
					Elogger.Fatal().Msgf("error decoding input json sequence: %v", err)
				}
				key, ok := entry["Key"]
				fmt.Println(key)
				sKey, ok2 := key.(string)
				if !(ok && ok2) {
					Elogger.Fatal().Msgf("error decoding input json sequence: %v", "Key field do nor exists or is not a string type")
				}
				delete(entry, "Key")
				inputRecords[sKey] = entry
			}
		} else if b1 && b2 {
			// input is a json array string

			err = json.NewDecoder(strings.NewReader(strings.Join(inData, " "))).Decode(&inputRecords)
			if err != nil {
				Elogger.Fatal().Msgf("error decoding input json array: %v", err)
			}
		} else {
			// yaml input or another looks like plain input - separated by "---"
			fmt.Println("plain input:")
			current := InputSecretEntry{Name: "#@IniT@#", Secret: "generate"}
			for _, v := range inData {

				splitted := strings.SplitN(v, ":", 2)
				// fmt.Println(splitted, len(splitted))
				if len(splitted) != 2 {
					Elogger.Fatal().Msg("error decoding plain input: wrong format")
				}
				splitted[0] = strings.TrimSpace(splitted[0])
				splitted[0] = strings.Trim(splitted[0], `"`)
				splitted[1] = strings.TrimSpace(splitted[1])
				splitted[1] = strings.Trim(splitted[1], `"`)
				switch splitted[0] {
				case "Name":
					current = InputSecretEntry{Secret: "generate"}
					current.Name = splitted[1]
				case "Login":
					current.Login = splitted[1]
				case "Description":
					current.Description = splitted[1]
				case "Secret":
					current.Secret = splitted[1]
				}

				// fmt.Printf("%v(len=%v):%+v\n", k, len(v), []rune(v))
			}

		}

		fmt.Printf("Before: %v\n\n", inputRecords)

		for key, record := range inputRecords {
			if recordMap, ok := record.(map[string]interface{}); ok {
				resolvedMap := make(map[string]string, 0)
				for innerKey, val := range recordMap {
					if sValue, ok := val.(string); ok {
						resValue, _ := resolveValue(sValue, recordMap, resolvedMap)
						recordMap[innerKey] = resValue

						sval := strings.TrimSpace(sValue)
						if strings.HasPrefix(sval, "{{") && strings.HasSuffix(sval, "}}") {
							sval = strings.Trim(strings.Trim(sval, "{"), "}")
							if sval == "Key" {
								recordMap[innerKey] = key
							} else {
								if sValToReplace, ok := recordMap[sval]; ok {
									recordMap[innerKey] = sValToReplace
								}
							}
						}
					}
				}
			}
		}
		fmt.Printf("After: %v", inputRecords)
	}
}

func resolveValue(sValue string, sourceMap map[string]interface{}, resolved map[string]string) (string, error) {
	if sourceMap == nil {
		return "", errors.New("empty source map")
	}
	if resolved == nil {
		resolved = make(map[string]string, 0)
	}
	sValue = strings.TrimSpace(sValue)
	matches, ok := mcli_utils.FindSubstrings(sValue)
	if !ok {
		return sValue, nil
	}
	for _, match := range matches {
		if resolvedMatch, ok := resolved[match]; ok {
			sValue = strings.ReplaceAll(sValue, match, resolvedMatch)
		}
	}
	return sValue, nil
}

// redisImportCmd represents the redisImport command
var redisImportCmd = &cobra.Command{
	Use:   "redis-import",
	Short: "Import records to redis database",
	Long: `Import records as strings to redis database. Format can be: yaml or json
	it is recommended to use arrays of json objects.
	`,
	Run: importToRedis,
}

func init() {
	utilsCmd.AddCommand(redisImportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	redisImportCmd.PersistentFlags().String("redis-host", "localhost:6379", "host and port to connect to REDIS DB")
	redisImportCmd.PersistentFlags().String("redis-pwd", "echo ${REDIS_PWD}", "Password for REDIS DB")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// redisImportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
