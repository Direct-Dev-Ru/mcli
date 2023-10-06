package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetStringParam(param string, cmd *cobra.Command, fromConfig string) (string, error) {

	paramValue, err := cmd.Flags().GetString(param)
	isParamSet := cmd.Flags().Lookup(param).Changed
	if !isParamSet && len(fromConfig) > 0 {
		paramValue = fromConfig
	}
	return paramValue, err
}

func IsPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsPathExistsAndCreate(pathParam string, create, asFile bool) (bool, string, error) {

	_, err := os.Stat(pathParam)
	if err == nil {
		// path exists - returning now
		fileInfo, _ := os.Stat(pathParam)
		if fileInfo.IsDir() {
			return true, "directory", nil
		} else {
			return true, "file", nil
		}
	}

	if os.IsNotExist(err) && create {

		pathToAnalyse := strings.TrimSpace(pathParam)
		pathToAnalyse = strings.ReplaceAll(pathToAnalyse, `\`, "/")
		itemType := "directory"
		fileName := path.Base(pathToAnalyse)
		pathToCreate := pathToAnalyse
		if asFile {
			withoutLastPart := path.Dir(pathToAnalyse)
			pathToCreate = withoutLastPart
		}

		// create directory
		err := os.MkdirAll(pathToCreate, os.ModePerm)
		if err != nil {
			return false, "", err
		}
		if asFile {
			itemType = "file"
			// create empty file
			err = os.WriteFile(filepath.Join(pathToCreate, fileName), []byte{}, os.ModePerm)
			if err != nil {
				return false, "", err
			}
		}
		return true, itemType, nil
	}
	return false, "", nil
}

// TODO: Integrate this function into all import commands ( secrets, redis and so on)
func processInput(inputArgs ...string) (map[string]interface{}, error) {
	inputFile := inputArgs[0]

	if !IsCommandInPipe() && len(inputFile) > 0 {
		exist, _, _ := mcli_utils.IsExistsAndCreate(inputFile, false, false)
		if !exist {
			Elogger.Fatal().Msg("no input file provided for command or input file doesn't exists")
		}
	}

	var inData []string
	var err error
	if IsCommandInPipe() && len(Input.InputSlice) > 0 {
		// Ilogger.Trace().Msg("input from pipe ...")
		inData = make([]string, 0, len(Input.InputSlice))
		for _, inputLine := range Input.InputSlice {
			currentLine := strings.TrimSpace(strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], ""))

			if len(currentLine) == 0 {
				continue
			}
			inData = append(inData, currentLine)
		}
	} else {
		inputSlice, err := mcli_fs.ReadFileLines(inputFile)
		if err != nil {
			Elogger.Fatal().Msgf("error reading input file: %v", err)
		}
		for _, inputLine := range inputSlice {
			currentLine := strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], "")
			currentLine = strings.TrimSpace(currentLine)
			if len(currentLine) == 0 {
				continue
			}
			inData = append(inData, currentLine)
		}

	}
	var inputRecords map[string]interface{} = make(map[string]interface{}, 0)

	if len(inData) > 0 {
		// fmt.Println("input data: ", strings.Join(inData, " "))
		keyPrefix := ""

		// detect json or not
		a1 := strings.HasPrefix(inData[0], "{")
		a2 := strings.HasSuffix(inData[len(inData)-1], "}")
		b1 := strings.HasPrefix(inData[0], "[")
		b2 := strings.HasSuffix(inData[len(inData)-1], "]")

		if a1 && a2 {
			dec := json.NewDecoder(strings.NewReader(strings.Join(inData, " ")))
			for dec.More() {
				var entry map[string]interface{}
				err := dec.Decode(&entry)
				if err != nil {
					Elogger.Fatal().Msgf("error decoding input json sequence: %v", err)
				}
				entry, _ = processEntry(entry, keyPrefix)
				inputRecords[entry["_Key"].(string)] = entry
			}
		} else if b1 && b2 {
			// input is a json array string
			inputArray := make([]map[string]interface{}, 0)
			err = json.NewDecoder(strings.NewReader(strings.Join(inData, " "))).Decode(&inputArray)
			if err != nil {
				Elogger.Fatal().Msgf("error decoding input json array: %v", err)
			}
			for _, entry := range inputArray {
				entry, _ = processEntry(entry, keyPrefix)
				inputRecords[entry["_Key"].(string)] = entry
			}
		} else {
			// yaml input or another looks like plain input - separated by "---"
			allInputData := strings.Join(inData, "\n")
			re := regexp.MustCompile(`---\s*`)
			// Split the text into blocks
			blocks := re.Split(allInputData, -1)
			// Extract and print the random text blocks
			if len(blocks) > 0 {
				for _, block := range blocks {
					var entry map[string]interface{} = make(map[string]interface{}, 0)
					err := yaml.Unmarshal([]byte(block), &entry)
					if err != nil {
						Elogger.Fatal().Msgf("error decoding yaml block: %v", err)
					}
					entry, _ = processEntry(entry, keyPrefix)
					inputRecords[entry["_Key"].(string)] = entry

				}
			} else {
				Elogger.Fatal().Msg("no data blocks found")
			}
		}

		// fmt.Printf("Before: %v\n\n", mcli_utils.PrettyPrintMap(inputRecords))

		for key, record := range inputRecords {
			if recordMap, ok := record.(map[string]interface{}); ok {
				// to store previously resolved fields
				resolvedMap := make(map[string]string, 0)
				// replace {{Key}} with key
				for innerKey, val := range recordMap {
					if sValue, ok := val.(string); ok {
						sValue = strings.ReplaceAll(sValue, "{{Key}}", key)
						recordMap[innerKey] = sValue
						resValue, _ := resolveValue(innerKey, sValue, recordMap, resolvedMap)
						resValue, _ = evaluateValue(resValue)
						recordMap[innerKey] = resValue
					}
				}
				// fmt.Printf("\nresolved: %v\n", resolvedMap)
			}
		}
		// fmt.Printf("\nAfter: %v\n", mcli_utils.PrettyPrintMap(inputRecords))
	}

	return inputRecords, nil
}

func processEntry(entry map[string]interface{}, globalPrefix string) (map[string]interface{}, error) {
	key, ok := entry["Key"]

	sKey, ok2 := key.(string)
	if !(ok && ok2) {
		Elogger.Fatal().Msgf("error decoding input : %v", "Key field do not exists or is not of string type")
	}

	sKey, _ = resolveValue("Key", sKey, entry, nil)
	entry["_Key"] = sKey
	delete(entry, "Key")

	iPrefix, ok := entry["Key-Prefix"]
	Prefix, ok2 := iPrefix.(string)
	if ok && ok2 && len(Prefix) != 0 {
		entry["_Key-Prefix"] = Prefix
	} else {
		entry["_Key-Prefix"] = globalPrefix
	}
	delete(entry, "Key-Prefix")

	return entry, nil
}

func getHttpRequest(partialPath string) ([]byte, error) {
	partialPath = strings.TrimSpace(partialPath)

	if strings.HasPrefix(partialPath, "http") {
		// Make the HTTP GET request
		response, err := http.Get(partialPath)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		// Check the response status code
		if response.StatusCode != http.StatusOK {
			err = fmt.Errorf("unexpected status code: %v", response.StatusCode)
			return nil, err
		}
		// Read the response body into a byte slice
		httpdata, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return httpdata, nil
	}
	return nil, fmt.Errorf("unexpected url")
}

func getFullPath(partialPath string) (string, error) {
	partialPath = strings.TrimSpace(partialPath)
	rootPath := GlobalMap["RootPath"]
	if len(partialPath) == 0 {
		return "", nil
	}

	if strings.HasPrefix(partialPath, "http") || strings.HasPrefix(partialPath, "//") {
		return partialPath, nil
	}
	if GlobalMap["OS"] == "linux" {
		if strings.HasPrefix(partialPath, "/") {
			return partialPath, nil
		}
		if strings.HasPrefix(partialPath, "./") {
			partialPath = strings.TrimPrefix(partialPath, `./`)
			return filepath.Join(rootPath, partialPath), nil
		}
		return filepath.Join(rootPath, partialPath), nil
	}
	if GlobalMap["OS"] == "windows" {
		partialPath = strings.ReplaceAll(partialPath, `/`, `\`)
		partialPath = strings.TrimPrefix(partialPath, `\`)

		if strings.HasPrefix(partialPath, `C:\`) || strings.HasPrefix(partialPath, `D:\`) ||
			strings.HasPrefix(partialPath, `E:\`) || strings.HasPrefix(partialPath, `F:\`) {
			//TODO: Rewrite for regexp
			return partialPath, nil
		}
		return filepath.Join(rootPath, partialPath), nil
	}
	return "", fmt.Errorf("partial path format doesn't support")
}
