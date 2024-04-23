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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	_ "gopkg.in/mgo.v2/bson"
)

type ContentWriteMessage struct {
	FolderName string
	FileName   string
	Content    []byte
}

type SimpleMessage struct {
	Key  string
	Data []byte
}

var exportCypher mcli_type.SecretsCypher = mcli_crypto.AesCypher

// var expErr mcli_error.CommonError

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

func WriteToOutputFileWorker(outDest string, keyIndex int, append bool, onlyData bool, ch chan SimpleMessage, wg *sync.WaitGroup) {
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

	for msg := range ch {
		// Write data to the file
		key := msg.Key
		data := msg.Data
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

func PrintRedisExportResult(path, key, data string, onlyData bool) {
	path = strings.TrimSpace(path)
	if path == "stdout" || path == "" {
		if !onlyData {
			fmt.Printf("[%s]\n", key)
			fmt.Println("")
		}
		fmt.Printf("%s\n", string(data))
	}
}

func WriteToFolderWorker(id int, flat bool, msgChan <-chan ContentWriteMessage, errChan chan<- error, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		fmt.Printf("Worker %d stopped\n", id)
		wg.Done()
	}()

	fmt.Printf("Worker %d started\n", id)

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return // message channel closed
			}

			if !flat {
				// Calculate result folder for storing content
				parts := strings.Split(msg.FileName, ":")
				if len(parts) > 1 {
					for i := 0; i < len(parts)-1; i++ {
						msg.FolderName = filepath.Join(msg.FolderName, parts[i])
					}
				}
				msg.FileName = parts[len(parts)-1]
			}

			// Create folder if it doesn't exist
			err := os.MkdirAll(msg.FolderName, 0755)
			if err != nil {
				errChan <- fmt.Errorf("error in worker %d creating folder %s: %v", id, msg.FolderName, err)
				continue
			}

			// Write content to file
			err = os.WriteFile(filepath.Join(msg.FolderName, msg.FileName), msg.Content, 0644)
			// if err != nil || msg.FileName == "certificates:stage:test-domain-com.crt" {
			if err != nil {
				errChan <- fmt.Errorf("error in worker %d writing to file %s/%s: %v", id, msg.FolderName, msg.FileName, err)
				continue
			}
			// fmt.Printf("Worker %d success writing to file %s/%s\n", id, msg.FolderName, msg.FileName)
			Ilogger.Trace().Msgf("Worker %d success writing to file %s/%s\n", id, msg.FolderName, msg.FileName)
		case _, ok := <-stop: // stop signal received
			if !ok {
				Ilogger.Trace().Msgf("Worker %d receives stop signal\n", id)
				return
			}
		default:
			// fmt.Printf("Worker %d : i am waiting \n", id)
			time.Sleep(1 * time.Millisecond)
		}
	}
}

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
	append, _ := cmd.Flags().GetBool("append")
	onlyData, _ := cmd.Flags().GetBool("only-data")
	clearDest, _ := cmd.Flags().GetBool("clear")
	flat, _ := cmd.Flags().GetBool("flat")

	keysArray, _ := GetStringParam("keys", cmd, "")
	keyPrefix, _ := GetStringParam("key-prefix", cmd, "")

	keys := strings.Split(keysArray, ",")
	if len(keys) == 0 {
		Elogger.Fatal().Msgf("error export keys from redis: %s", "no keys provided to retrieve")
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

	redisEncKey := ""
	if true {
		// getting encryption redis key
		internalSecretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile, exportCypher, nil)
		if err := internalSecretStore.FillStore(Config.Common.InternalVaultPath, Config.Common.InternalKeyFilePath); err != nil {
			if decrypt {
				Elogger.Fatal().Msgf("error filling secret store %v", err)
			}
		}
		redisEncKeySecret, ok := internalSecretStore.GetSecretPlainMap()["RedisEncKey"]
		if ok {
			redisEncKey = redisEncKeySecret.Secret
			// Ilogger.Trace().Msgf("redis encryption key have been retrived from store: %s", fmt.Sprintf("%x", redisEncKey))
			if decrypt {
				// setup encryption parameters
				kvStore.SetEncrypt(decrypt, []byte(redisEncKey), cypher)
			}
		}
	}

	kvStore.SetMarshalling(json.Marshal, json.Unmarshal)
	var wg sync.WaitGroup
	ch := make(chan SimpleMessage)
	numWorkers := 3
	msgChan := make(chan ContentWriteMessage, numWorkers)
	errChan := make(chan error, numWorkers)
	stopCh := make(chan struct{})

	// messages := []ContentWriteMessage{
	// 	{FolderName: outDest, FileName: "certificates:stage:test-domain-com.key", Content: []byte("Content for certificates:stage:test-domain-com.key")},
	// 	{FolderName: outDest, FileName: "certificates:ex2.domain.com", Content: []byte("Content for certificates:ex2.domain.com")},
	// 	{FolderName: outDest, FileName: "certificates:stage:test-domain-com.crt", Content: []byte("Content for certificates:stage:test-domain-com.crt")},
	// 	{FolderName: outDest, FileName: "certificates:example.domain.ru", Content: []byte("Content for certificates:example.domain.ru")},
	// 	{FolderName: outDest, FileName: "secret.key", Content: []byte("Content for secret.key")},
	// }

	// _ = messages

	firstKey := true
	outStdOut := outDest == "" || outDest == "stdout"
	for _, key := range keys {
		key = strings.TrimSpace(key)
		// mKey := strings.HasSuffix(key, "*")
		var err error
		var data []byte
		var mData map[string][]byte
		mData, err = kvStore.GetRecords(key, keyPrefix)
		if err != nil {
			// try read record with decryption
			if !decrypt {
				// setup encryption parameters
				kvStore.SetEncrypt(true, []byte(redisEncKey), cypher)
				mData, err = kvStore.GetRecords(key, keyPrefix)
				if err != nil {
					Elogger.Fatal().Msgf("error retriving decrypting key(s) %s, prefix %s: %v", key, keyPrefix, err)
				}
				kvStore.SetEncrypt(false, []byte(redisEncKey), cypher)
			} else {
				Elogger.Fatal().Msgf("error retriving key(s) %s, prefix %s: %v", key, keyPrefix, err)
			}
		}

		// ok := true
		// data := []byte("test data")
		// for keyIndex, key := range messages {
		// finalKey := ""

		if outStdOut {
			for key, data := range mData {
				PrintRedisExportResult(outDest, key, string(data), onlyData)
			}
		}

		// if output to single file
		if !outStdOut && (filepath.Ext(outDest) != "" && filepath.Ext(outDest) != ".d" && filepath.Ext(outDest) != ".hash") {

			if firstKey {
				wg.Add(1)
				go WriteToOutputFileWorker(outDest, mcli_utils.Iif[int](firstKey, 0, 1), append, onlyData, ch, &wg)
			}
			for key, data := range mData {
				// finalKey = mcli_utils.Iif[string](len(keyPrefix) == 0, key, fmt.Sprintf("%s:%s", keyPrefix, strings.ReplaceAll(key, "*", ".multi")))
				ch <- SimpleMessage{Key: key, Data: data}
			}
			firstKey = false

		}

		if !outStdOut && (filepath.Ext(outDest) == "" || filepath.Ext(outDest) == ".d") {
			if firstKey {
				if clearDest {
					err := os.RemoveAll(outDest)
					if err != nil {
						Elogger.Fatal().Msgf("fatal error from worker: %v\n", err)
						return
					}
				}
				runtime.GOMAXPROCS(2)
				// Start worker goroutines
				for i := 0; i < numWorkers; i++ {
					wg.Add(1)
					go WriteToFolderWorker(i, flat, msgChan, errChan, stopCh, &wg)
				}
				firstKey = false
			}
			msgChan <- ContentWriteMessage{FolderName: outDest, FileName: key, Content: data}
		}

		if filepath.Ext(outDest) == ".hash" {
			fmt.Println(".hash saving ...")
		}

	}

	if !outStdOut && (filepath.Ext(outDest) != "" && filepath.Ext(outDest) != ".d" && filepath.Ext(outDest) != ".hash") {
		close(ch)
		wg.Wait()
	}
	if !outStdOut && (filepath.Ext(outDest) == "" || filepath.Ext(outDest) == ".d") {
		close(msgChan)
		// Wait for workers to finish or stop
		go func() {
			wg.Wait()
			close(errChan)
			close(stopCh)
		}()
		// Collect errors from workers
		for err := range errChan {
			Elogger.Fatal().Msgf("fatal error from worker: %v\n", err)
			close(stopCh) // Send stop signal to all workers
		}

	}

	_, _, _, _, _, _ = decrypt, redisHost, isOutFileSet, keys, keyPrefix, redisEncKey
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
	redisExportCmd.PersistentFlags().IntP("redis-db", "D", 1, "number of redis database to export from")
	redisExportCmd.PersistentFlags().StringP("redis-pwd", "P", "echo ${REDIS_PWD}", "Password for REDIS")
	redisExportCmd.PersistentFlags().StringP("out-dest", "o", "stdout", "filepath/stdout to output results")
	redisExportCmd.PersistentFlags().StringP("key-prefix", "x", "", "prefix to add to key then store in redis database")
	redisExportCmd.PersistentFlags().StringP("keys", "k", "", "keys (comma separated) to retrieve from redis database")
	redisExportCmd.Flags().BoolP("decrypt", "d", false, "decrypt records before output - if omits trying force decryption in case of error retriving undecrypted value(s)")
	redisExportCmd.Flags().BoolP("append", "a", false, "do not recreate out file(s) if it is exists - append always")
	redisExportCmd.Flags().BoolP("only-data", "O", false, "prints out only data - don't prints out key name")
	redisExportCmd.Flags().BoolP("clear", "c", false, "clear dest folder before export")
	redisExportCmd.Flags().BoolP("flat", "f", false, "do not transform key parts (:) to subfolders")
	// redisExportCmd.PersistentFlags().StringP("encrypt-key-name", "E", "mcli-redis-enc-name", "prefix to add to key then store in redis database")
}
