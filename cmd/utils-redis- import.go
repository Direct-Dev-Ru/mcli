/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_error "mcli/packages/mcli-error"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_redis "mcli/packages/mcli-redis"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_store "mcli/packages/mcli-store"
	mcli_utils "mcli/packages/mcli-utils"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cypher mcli_secrets.SecretsCypher = mcli_crypto.AesCypher
var Err mcli_error.CommonError
var F map[string]func(string) (string, error) = make(map[string]func(string) (string, error), 0)

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

	redisHost, _ := cmd.Flags().GetString("redis-host")
	// isRedisHostSet := cmd.Flags().Lookup("redis-host").Changed

	redisPwd, _ := cmd.Flags().GetString("redis-pwd")
	isRedisPwdSet := cmd.Flags().Lookup("redis-pwd").Changed
	if (len(redisPwd) == 0 || redisPwd == "echo ${REDIS_PWD}") && !isRedisPwdSet {
		redisPwd = os.Getenv("REDIS_PWD")
	}

	inputFile, _ := cmd.Flags().GetString("input-file")
	isInputFileSet := cmd.Flags().Lookup("input-file").Changed

	encrypt, _ := cmd.Flags().GetBool("encrypt")
	isEncryptSet := cmd.Flags().Lookup("encrypt").Changed
	if !isEncryptSet {
		encrypt = false
	}

	if !IsCommandInPipe() && len(inputFile) > 0 && isInputFileSet {
		exist, _, _ := mcli_utils.IsExistsAndCreate(inputFile, false)
		if !exist {
			Elogger.Fatal().Msg("noinput provided for command or input file doesn't exists")
		}
	}
	keyPrefix, _ := cmd.Flags().GetString("key-prefix")

	_ = redisHost

	var inData []string
	if IsCommandInPipe() && len(Input.InputSlice) > 0 {
		// Ilogger.Trace().Msg("input from pipe ...")
		inData = make([]string, 0, len(Input.InputSlice))
		for _, inputLine := range Input.InputSlice {
			currentLine := strings.ReplaceAll(inputLine, GlobalMap["LineBreak"], "")
			currentLine = strings.TrimSpace(currentLine)
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

	if len(inData) > 0 {
		// fmt.Println("input data: ", strings.Join(inData, " "))

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

		// ok now we are ready to import records to kv store
		var kvStore mcli_store.KVStorer

		kvStore, err = mcli_redis.NewRedisStore(redisHost, redisPwd, "")
		if err != nil {
			Elogger.Fatal().Msg(err.Error())
		}
		if encrypt {
			enckey, err := mcli_secrets.LoadKeyFromFilePlain(Config.Common.InternalKeyFilePath)
			if err != nil {
				Elogger.Fatal().Msgf("load rootSecretStore_key error: %s", err.Error())
			}
			kvStore.SetEcrypt(encrypt, enckey, cypher)
		}
		keyToCheck := ""
		for key, record := range inputRecords {
			keyToCheck = key
			if recordMap, ok := record.(map[string]interface{}); ok {
				prefix, ok := recordMap["_Key-Prefix"]
				if sPrefix, ok2 := prefix.(string); ok && ok2 {
					err = kvStore.SetRecord(key, record, sPrefix)
					keyToCheck = fmt.Sprintf("%s:%s", sPrefix, key)
				} else {
					err = kvStore.SetRecord(key, record)
				}
				if err != nil {
					Elogger.Fatal().Msgf("save record error: %s", err.Error())
				}
			}
		}

		_, err, ok := kvStore.GetRecord(keyToCheck)
		if !ok || err != nil {
			Elogger.Fatal().Msgf("check record error: %s", err.Error())
		}
		Ilogger.Info().Msg("import successful")
		os.Exit(0)
	}
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

func evaluateValue(sValue string) (string, error) {

	if len(sValue) == 0 {
		return "", errors.New("empty value")
	}
	if !strings.HasPrefix(sValue, "F.") {
		return sValue, nil
	}

	re := regexp.MustCompile(`F\.(.*?)\((.*?)\)`)
	found := re.FindAllStringSubmatch(sValue, -1)

	if len(found) > 0 {
		for _, f := range found {
			if len(f) == 3 {
				funcName := f[1]
				funcParam := f[2]
				if function, ok := F[funcName]; ok {
					result, err := function(funcParam)
					if err != nil {
						return sValue, err
					}
					sValue = result
				}
			}
		}

	}
	return sValue, nil
}

func resolveValue(sKey, sValue string, sourceMap map[string]interface{}, resolved map[string]string) (string, error) {
	if sourceMap == nil {
		return "", errors.New("empty source map")
	}
	if resolved == nil {
		resolved = make(map[string]string, 0)
	}
	sValue = strings.TrimSpace(sValue)
	matches, ok := mcli_utils.FindSubstrings(sValue, `{{(.*?)}}`)
	if !ok {
		return sValue, nil
	}
	for _, match := range matches {
		keyMatch := strings.ReplaceAll(strings.ReplaceAll(match, "{", ""), "}", "")
		// fmt.Println(match, keyMatch, sourceMap)
		if resolvedMatch, ok := resolved[match]; ok {
			// fmt.Println("get from resolved before for ", sKey, match, sValue, resolvedMatch)
			sValue = strings.ReplaceAll(sValue, match, resolvedMatch)
			// fmt.Println("get from resolved after for ", sKey, match, sValue, resolvedMatch)
		} else {
			if resolvedMatch, ok := sourceMap[keyMatch]; ok {
				if sVal, ok := resolvedMatch.(string); ok {
					_, ok := mcli_utils.FindSubstrings(sVal, `{{(.*?)}}`)
					if strings.Contains(sVal, fmt.Sprintf("{{%s}}", sKey)) {
						return "", fmt.Errorf("deadlock ring resolution of key %s", sKey)
					}
					if ok {
						sVal, _ = resolveValue(keyMatch, sVal, sourceMap, resolved)
						sValue = strings.ReplaceAll(sValue, match, sVal)
					} else {
						sValue = strings.ReplaceAll(sValue, match, sVal)
					}
					// fmt.Println("set to resolved for ", sKey, match, sVal, resolvedMatch)
					resolved[match] = sVal
				}
			}
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
	F["HASH"] = mcli_crypto.HashPassword
	F["UPPER"] = func(value string) (string, error) {
		return strings.ToUpper(value), nil
	}
	F["LOWER"] = func(value string) (string, error) {
		return strings.ToLower(value), nil
	}
	F["GENPWD"] = func(value string) (string, error) {
		pwdlen := int64(12)
		var err error
		if value != "" {
			pwdlen, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				pwdlen = int64(math.Max(float64(len(value)), float64(12)))
			}
		}
		return mcli_crypto.GeneratePassword(int(pwdlen))
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	redisImportCmd.PersistentFlags().String("redis-host", "localhost:6379", "host:port to connect to REDIS DB")
	redisImportCmd.PersistentFlags().String("redis-pwd", "echo ${REDIS_PWD}", "Password for REDIS DB")
	redisImportCmd.PersistentFlags().String("input-file", "", "file with data for import")
	redisImportCmd.PersistentFlags().String("key-prefix", "", "prefix to add to key then store in redis database")
	redisImportCmd.Flags().BoolP("encrypt", "e", false, "encrypt records")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// redisImportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
