package cmd

import (
	"encoding/json"
	"fmt"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_redis "mcli/packages/mcli-redis"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_type "mcli/packages/mcli-type"
	mcli_utils "mcli/packages/mcli-utils"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	_ "gopkg.in/mgo.v2/bson"
)

var exportCypher mcli_type.SecretsCypher = mcli_crypto.AesCypher

// var expErr mcli_error.CommonError

func exportFromRedis(cmd *cobra.Command, args []string) {

	defaultRedisHost, defaultRedisPort := strings.Split(Config.Common.RedisHost, ":")[0], strings.Split(Config.Common.RedisHost, ":")[1]
	redisHost, _ := GetStringParam("redis-host", cmd, defaultRedisHost)
	redisPort, _ := GetStringParam("redis-port", cmd, defaultRedisPort)
	redisDb, _ := GetIntParam("redis-db", cmd, Config.Common.RedisDatabaseNo)
	redisPwd, _ := GetStringParam("redis-pwd", cmd, Config.Common.RedisPwd)
	isRedisPwdSet := cmd.Flags().Lookup("redis-pwd").Changed
	if (len(redisPwd) == 0 || redisPwd == "echo ${REDIS_PWD}") && !isRedisPwdSet {
		redisPwd = os.Getenv("REDIS_PWD")
	}

	outDest, _ := cmd.Flags().GetString("out-dest")
	isOutFileSet := cmd.Flags().Lookup("out-dest").Changed

	decrypt, _ := cmd.Flags().GetBool("decrypt")
	isDecryptSet := cmd.Flags().Lookup("decrypt").Changed
	if !isDecryptSet {
		decrypt = false
	}
	append, _ := cmd.Flags().GetBool("append")
	isAppendSet := cmd.Flags().Lookup("append").Changed
	if !isAppendSet {
		append = false
	}
	onlyData, _ := cmd.Flags().GetBool("only-data")
	isOnlyDataSet := cmd.Flags().Lookup("only-data").Changed
	if !isOnlyDataSet {
		onlyData = false
	}

	keysArray, _ := GetStringParam("keys", cmd, "")
	keyPrefix, _ := GetStringParam("key-prefix", cmd, "")

	keys := strings.Split(keysArray, ",")
	if len(keys) == 0 {
		Elogger.Fatal().Msgf("error export keys from redis: %s", "no keys provided to retrieve")
	}

	redisEncKey := ""
	if decrypt {
		// getting encryption redis key
		internalSecretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile, exportCypher, nil)
		if err := internalSecretStore.FillStore(Config.Common.InternalVaultPath, Config.Common.InternalKeyFilePath); err != nil {
			Elogger.Fatal().Msgf("error filling secret store %v", err)
		}
		redisEncKeySecret, ok := internalSecretStore.GetSecretPlainMap()["RedisEncKey"]
		if ok {
			redisEncKey = redisEncKeySecret.Secret
			// Ilogger.Trace().Msgf("redis encryption key have been retrived from store: %s", fmt.Sprintf("%x", redisEncKey))
		}
	}

	// ok now we are ready to export records from kv store
	var kvStore mcli_type.KVStorer
	var err error
	resultHostToConnect := fmt.Sprintf("%s:%s", redisHost, redisPort)

	kvStore, err = mcli_redis.NewRedisStore("redisutils_"+Config.Common.AppName, resultHostToConnect, redisPwd, "", redisDb)
	if err != nil {
		Elogger.Fatal().Msgf("error creating new redis store: %v", err)
	}
	defer kvStore.Close()

	kvStore.SetMarshalling(json.Marshal, json.Unmarshal)
	var wg sync.WaitGroup
	ch := make(chan string)
	for keyIndex, key := range keys {
		key = strings.TrimSpace(key)
		data, err, ok := kvStore.GetRecord(key, keyPrefix)
		if err != nil {
			Elogger.Fatal().Msgf("error retriving key %s, prefix %s: %v", key, keyPrefix, err)
		}
		if ok {
			finalKey := mcli_utils.Iif[string](len(keyPrefix) == 0, key, fmt.Sprintf("%s:%s", keyPrefix, key))
			// if output to file - experiment with goroutine
			if filepath.Ext(outDest) != "" && filepath.Ext(outDest) != ".d" && filepath.Ext(outDest) != ".hash" {
				if keyIndex == 0 {
					wg.Add(1)
					go WriteToOutputFile_Go(outDest, keyIndex, append, onlyData, finalKey, ch, &wg)
				}
				ch <- string(data)
				// err = WriteToOutputFile(outDest, keyIndex, append, onlyData, key, string(data))
				// if err != nil {
				// 	Elogger.Fatal().Msgf("error writing to the file %s: %v", outDest, err)
				// 	return
				// }
			}

			if outDest == "" || outDest == "stdout" {

				PrintRedisExportResult(outDest, finalKey, string(data), false)
			}
		}
	}

	close(ch)
	wg.Wait()

	_, _, _, _, _, _ = decrypt, redisHost, isOutFileSet, keys, keyPrefix, redisEncKey
}

func WriteToOutputFile(outDest string, keyIndex int, append bool, onlyData bool, key string, data string) error {

	var file *os.File
	var err error
	var dirPath string = filepath.Dir(outDest)
	if keyIndex == 0 {
		// Check if the directory exists
		if _, err = os.Stat(dirPath); os.IsNotExist(err) {
			// Directory does not exist, create it
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %v", dirPath, err)
			}
		} else if err != nil {
			return fmt.Errorf("error checking directory %s: %v", dirPath, err)
		}
		if !append {
			os.Remove(outDest)
			file, err = os.OpenFile(outDest, os.O_WRONLY|os.O_CREATE, 0644)
		} else {
			file, err = os.OpenFile(outDest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		}
		if err != nil {
			return fmt.Errorf("error opening/creating file %s: %v", outDest, err)
		}
		defer file.Close()
	} else {
		file, err = os.OpenFile(outDest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("error opening/creating file %s: %v", outDest, err)
		}
		defer file.Close()
	}

	// Write data to the file
	if onlyData {
		_, err = file.WriteString(fmt.Sprintf("%s%s", data, GlobalMap["LineBreak"]))
	} else {
		_, err = file.WriteString(fmt.Sprintf("%s%s%s%s", "["+key+"]", GlobalMap["LineBreak"]+GlobalMap["LineBreak"], data, GlobalMap["LineBreak"]))
	}
	if err != nil {
		return fmt.Errorf("error writing to the file %s: %v", outDest, err)
	}

	return nil
}

func WriteToOutputFile_Go(outDest string, keyIndex int, append bool, onlyData bool, key string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	var file *os.File
	var err error
	var dirPath string = filepath.Dir(outDest)
	if keyIndex == 0 {
		// Check if the directory exists
		if _, err = os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				Elogger.Err(err).Msgf("error creating directory %s: %v", dirPath, err)
				close(ch)
				return
			}
		} else if err != nil {
			Elogger.Err(err).Msgf("error checking directory %s: %v", dirPath, err)
			close(ch)
			return
		}
		if !append {
			os.Remove(outDest)
			file, err = os.OpenFile(outDest, os.O_WRONLY|os.O_CREATE, 0644)
		} else {
			file, err = os.OpenFile(outDest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		}
		if err != nil {
			Elogger.Err(err).Msgf("error opening/creating file %s: %v", outDest, err)
			close(ch)
			return
		}
	} else {
		file, err = os.OpenFile(outDest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			Elogger.Err(err).Msgf("error opening/creating file %s: %v", outDest, err)
			close(ch)
			return
		}

	}
	defer file.Close()

	for data := range ch {
		// Write data to the file
		if onlyData {
			_, err = file.WriteString(fmt.Sprintf("%s%s", data, GlobalMap["LineBreak"]))
		} else {
			_, err = file.WriteString(fmt.Sprintf("%s%s%s%s", "["+key+"]", GlobalMap["LineBreak"]+GlobalMap["LineBreak"], data, GlobalMap["LineBreak"]))
		}
		if err != nil {
			Elogger.Err(err).Msgf("error writing to the file %s: %v", outDest, err)
			close(ch)
			return
		}
	}

}

func PrintRedisExportResult(path, key, data string, plain bool) error {
	path = strings.TrimSpace(path)
	if path == "stdout" || path == "" {
		fmt.Printf("[%s]\n", key)
		fmt.Println("")
		fmt.Printf("%s\n", string(data))
	}
	// if output to file ( path contains extension )
	if false && (filepath.Ext(path) != "" && filepath.Ext(path) != ".d" && filepath.Ext(path) != ".hash") {
		dirPath := filepath.Dir(path)
		// Check if the directory exists
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			// Directory does not exist, create it
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %v", dirPath, err)
			}

		} else if err != nil {
			return fmt.Errorf("error checking directory %s: %v", dirPath, err)
		}
		// Open the file with append mode, create it if it doesn't exist
		// file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("error opening file %s : %v", path, err)
		}
		defer file.Close()
		// Write data to the file
		_, err = file.WriteString(fmt.Sprintf("%s%s%s%s", "["+key+"]", GlobalMap["LineBreak"]+GlobalMap["LineBreak"],
			data, GlobalMap["LineBreak"]+GlobalMap["LineBreak"]))
		if err != nil {
			return fmt.Errorf("error writing to file %s : %v", path, err)
		}
	}
	return nil
}

// redisExportCmd represents the redisExport command
var redisExportCmd = &cobra.Command{
	Use:   "redis-export",
	Short: "Export from redis database",
	Long:  `This command exports records as strings from redis database.`,
	Run:   exportFromRedis,
}

func init() {
	utilsCmd.AddCommand(redisExportCmd)
	redisExportCmd.PersistentFlags().StringP("redis-host", "H", "127.0.0.1", "host to connect")
	redisExportCmd.PersistentFlags().StringP("redis-port", "p", "6379", "port to connect")
	redisExportCmd.PersistentFlags().IntP("redis-db", "D", 1, "redis table to retrieve keys")
	redisExportCmd.PersistentFlags().StringP("redis-pwd", "P", "echo ${REDIS_PWD}", "Password for REDIS")
	redisExportCmd.PersistentFlags().StringP("out-dest", "o", "stdout", "filepath/stdout to output results")
	redisExportCmd.PersistentFlags().StringP("key-prefix", "x", "", "prefix to add to key then store in redis database")
	redisExportCmd.PersistentFlags().StringP("keys", "k", "", "keys (comma separated) to retrieve from redis database")
	redisExportCmd.Flags().BoolP("decrypt", "d", false, "decrypt records before output - ignore errors and prints as is")
	redisExportCmd.Flags().BoolP("append", "a", false, "do not recreate out file(s) if it is exists - append always")
	redisExportCmd.Flags().BoolP("only-data", "O", false, "prints out only data - don't prints out key name")
	// redisExportCmd.PersistentFlags().StringP("encrypt-key-name", "E", "mcli-redis-enc-name", "prefix to add to key then store in redis database")
}
