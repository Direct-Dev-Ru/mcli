package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"regexp"
	"runtime"
	"strings"

	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_interface "mcli/packages/mcli-interface"
	mcli_secrets "mcli/packages/mcli-secrets"
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
func ProcessInput(inputArgs ...string) (map[string]interface{}, error) {
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

func GetHttpRequest(partialPath string) ([]byte, error) {
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

// LoadHttpPlugins loads HTTP plugins from a shared object (.so) file.
//
// It takes two parameters:
//   - module: The path to the shared object file containing the HTTP plugins.
//     If empty, it defaults to "plugins/http_default_plugins/http_plugins_compiled/http_default_handlers.so".
//   - name:   The name of the symbol (exported function or variable) to be looked up in the shared object.
//     This symbol is expected to implement the mcli_interface.HandlerFuncsPlugin interface.
//
// The function returns a map of string to http.HandlerFunc, representing the loaded HTTP handler functions.
// If successful, the map and nil error are returned. If any error occurs during the loading process, the function
// returns nil map and an error describing the issue.
func LoadHttpPlugins(module string, name string) (map[string]http.HandlerFunc, error) {
	// Check if the operating system is Linux
	if runtime.GOOS != "linux" {
		return nil, errors.New("current platform not supported, plugins are only supported on Linux")
	}

	if module == "" {
		module = "plugins/http_default_plugins/http_plugins_compiled/http_default_handlers.so"
	}

	// 1. open the so file to load the symbols
	plug, err := plugin.Open(module)
	if err != nil {
		return nil, err
	}

	// 2. look up a symbol (an exported function or variable)
	// in this case, variable Greeter
	symVariable, err := plug.Lookup(name)
	if err != nil {
		return nil, err
	}

	// 3. Assert that loaded symbol is of a desired type
	// in this case interface type Greeter (defined above)
	var handlerFuncs mcli_interface.HandlerFuncsPlugin
	handlerFuncs, ok := symVariable.(mcli_interface.HandlerFuncsPlugin)
	if !ok {
		return nil, err
	}

	// handlerFuncs.GetHandlerFuncs()
	mapHandlerFuncs := handlerFuncs.GetHandlerFuncsV2("HTTP_ECHO")

	return mapHandlerFuncs, nil
}

func LoadPlugin(modulePath string, objectName string) (plugin.Symbol, error) {
	// Check if the operating system is Linux
	if runtime.GOOS != "linux" {
		return nil, errors.New("current platform not supported plugins, plugins are only supported on Linux")
	}

	if modulePath == "" {
		modulePath = "plugins/default_plugin.so"
	}

	// 1. open the so file to load the symbols
	plug, err := plugin.Open(modulePath)
	if err != nil {
		return nil, err
	}

	// 2. look up a symbol (an exported function or variable)
	// in this case, variable Greeter
	symVariable, err := plug.Lookup(objectName)
	if err != nil {
		return nil, err
	}

	// 3. Assert that loaded symbol is of a desired type
	// in this case interface type Greeter (defined above)
	// var handlerFuncs mcli_interface.HandlerFuncsPlugin
	// handlerFuncs, ok := symVariable.(mcli_interface.HandlerFuncsPlugin)
	// if !ok {
	// 	return nil, err
	// }

	// handlerFuncs.GetHandlerFuncs()
	// mapHandlerFuncs := handlerFuncs.GetHandlerFuncsV2("HTTP_ECHO")

	return symVariable, nil
}

func InitInternalSecreVault(cfg *ConfigData) {
	// TODO: add store this secret values in redis db
	// read or create key for internal secrets
	var rootKeySecretStorePath = filepath.Dir(cfg.Common.InternalKeyFilePath)
	var rootSecretStore_key = cfg.Common.InternalKeyFilePath

	if len(cfg.Common.InternalKeyFilePath) == 0 {
		rootKeySecretStorePath = filepath.Join(GlobalMap["HomeDir"], ".mcli", "root")
		rootSecretStore_key = filepath.Join(rootKeySecretStorePath, "rootkey.key")
		cfg.Common.InternalKeyFilePath = rootSecretStore_key
	}

	_, _, err := mcli_utils.IsExistsAndCreate(rootKeySecretStorePath, true, false)
	if err != nil {
		Elogger.Fatal().Msgf("root secret store error - path do not exists: %s", err.Error())
	}
	ok, _, _ := mcli_utils.IsExistsAndCreate(rootSecretStore_key, false, false)

	if !ok {
		err = mcli_secrets.SaveKeyToFilePlain(rootSecretStore_key, mcli_secrets.GenKey(1024))
		if err != nil {
			Elogger.Fatal().Msgf("root secret store error - save rootSecretStore_key error: %s", err.Error())
		}
	}
	// read root secret from file
	rootInternalSecret, err := mcli_secrets.LoadKeyFromFilePlain(rootSecretStore_key)
	if err != nil {
		Elogger.Fatal().Msgf("root secret store error - load rootSecretStore_key error: %s", err.Error())
	}
	_, err = Config.Cache.Set("RootInternalSecret", rootInternalSecret)
	if err != nil {
		Elogger.Fatal().Msgf("root secret store in cache error: %s", err.Error())
	}
	GlobalMap["RootSecretKeyPath"] = rootSecretStore_key

	// paths to internal secret vault
	var internalSecretVaultBasePath = filepath.Dir(Config.Common.InternalVaultPath)
	var internalSecretVaultPath = Config.Common.InternalVaultPath
	if len(Config.Common.InternalVaultPath) == 0 {
		internalSecretVaultBasePath = filepath.Join(GlobalMap["RootPath"], "internal-secrets")
		internalSecretVaultPath = filepath.Join(internalSecretVaultBasePath, "internal.vault")
		Config.Common.InternalVaultPath = internalSecretVaultPath
	}
	_, _, err = mcli_utils.IsExistsAndCreate(internalSecretVaultBasePath, true, false)
	if err != nil {
		Elogger.Fatal().Msgf("internal vault secret store error : %v", err.Error())
	}
	GlobalMap["RootSecretVaultPath"] = internalSecretVaultPath

	for gKey, gValue := range GlobalMap {
		_, err = Config.Cache.Set(gKey, gValue)
		if err != nil {
			Elogger.Fatal().Msg("set cache value error " + err.Error())
		}
	}
}

func ReadConfigFile(configFile string) (err error) {

	if len(configFile) == 0 {
		configFile = GlobalMap["DefaultConfigPath"]
	}

	if configFile != "" {
		Ilogger.Trace().Msg(fmt.Sprint("parsing config file:", configFile))

		if _, err := os.Stat(configFile); err == nil {
			configContent, err := os.ReadFile(configFile)
			if err != nil {
				Elogger.Err(err).Msg("config file " + configFile + " does not exist")
				return err
			}
			configContentString := string(configContent)

			templateRegExp, err := regexp.Compile(`{{\$.+?}}`)
			if err != nil {
				Elogger.Err(err).Msg("config file " + configFile + " does not exist")
				return err
			}
			allVarsEntries := mcli_utils.RemoveDuplicatesStr(templateRegExp.FindAllString(configContentString, -1))
			for _, varEntry := range allVarsEntries {
				// if template entry end on $}} - we replace it from GlobalMap
				if strings.HasSuffix(varEntry, "$}}") {
					mapkey := strings.ReplaceAll(varEntry, "{{$", "")
					mapkey = strings.ReplaceAll(mapkey, "$}}", "")
					configContentString = strings.ReplaceAll(configContentString, varEntry, GlobalMap[mapkey])
				}
				// if template entry end on }} - we replace it from env
				if strings.HasSuffix(varEntry, "}}") && !strings.HasSuffix(varEntry, "$}}") {
					osEnv := strings.ReplaceAll(varEntry, "{{$", "")
					osEnv = strings.ReplaceAll(osEnv, "}}", "")
					configContentString = strings.ReplaceAll(configContentString, varEntry, os.Getenv(osEnv))
				}
			}

			if err == nil {
				err = yaml.Unmarshal([]byte(configContentString), &Config)
				if err != nil {
					Elogger.Fatal().Msg(err.Error())
				}
			}
			// fmt.Println("Configuration content :", string(configContent))
			Ilogger.Trace().Msg(fmt.Sprintf("Configuration struct : %+v", Config))
		} else if errors.Is(err, os.ErrNotExist) {
			Elogger.Err(err).Msg("config file " + configFile + " does not exist")
			return err
		} else {
			Elogger.Err(err).Msg("config file detect error " + err.Error())
			return err
		}
	}
	return
}

func Deprecated_ProcessInputParameter(paramName, envName string, command *cobra.Command) (string, error) {
	resultValue, _ := command.Flags().GetString(paramName)
	isParamSet := command.Flags().Lookup(paramName).Changed
	if !isParamSet && len(os.Getenv(envName)) > 0 {
		resultValue = os.Getenv(envName)
	}
	if isParamSet || len(os.Getenv(envName)) == 0 {
		os.Setenv(envName, resultValue)
	}
	return resultValue, nil
}

// ProcessInputParameter handles input parameters for a Cobra command,
// setting environment variables based on command line flags and vice versa.
//
// It retrieves the parameter value from the command flags, and if the flag
// is not set and an environment variable with the specified name exists,
// it uses the environment variable value. If the flag is set or no environment
// variable is found, it sets the environment variable with the flag value.
//
// Parameters:
//   - paramName: The name of the command line flag.
//   - envName: The name of the environment variable.
//   - command: The Cobra command to extract the flag value from.
//
// Returns:
//   - The resulting parameter value (from flag or environment variable).
//   - An error if there is an issue with flag retrieval or setting the environment variable.
func ProcessCommandParameter(paramName, envName string, command *cobra.Command) (string, error) {
	flag := command.Flags().Lookup(paramName)
	if flag == nil {
		return "", errors.New("flag not found")
	}

	resultValue, err := command.Flags().GetString(paramName)
	if err != nil {
		return "", err
	}

	if !flag.Changed && len(os.Getenv(envName)) > 0 {
		resultValue = os.Getenv(envName)
	}

	if flag.Changed || len(os.Getenv(envName)) == 0 {
		os.Setenv(envName, resultValue)
	}

	return resultValue, nil
}
