/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func set(newData *SomeData, wc chan *SomeData) {
	wc <- newData
}
func get(rc chan *SomeData) *SomeData {
	return <-rc
}

func monitor(rc chan *SomeData, wc chan *SomeData, db *AccumData) {
	var someData *SomeData
	defer fmt.Println("close monitor")
	for {
		select {
		case newData := <-wc:
			someData = newData
			db.Lock()
			db.data[strconv.Itoa(newData.payload)] = *newData
			db.Unlock()
			// fmt.Printf("%d \n", someData.payload)
		case rc <- someData:
		}
	}

}

var Config ConfigData = ConfigData{}
var rootCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	config, _ := cmd.Flags().GetString("config")
	rootArgs, _ := cmd.Flags().GetString("root-args")
	if len(config) == 0 {
		config = GlobalMap["DefaultConfigPath"]
	}
	Ilogger.Info().Msg("Hello from Multy CLI. Config file = " + config)

	if len(args) == 0 {
		args = strings.Fields(rootArgs)
	}
	n, err := strconv.Atoi("3")
	if len(args) > 0 {
		n, err = strconv.Atoi(args[0])
	}

	if err != nil {
		Elogger.Error().Msg("mcli: " + err.Error())
		n, _ = strconv.Atoi("3")
	}

	var readData = make(chan *SomeData)
	var writeData = make(chan *SomeData)
	var accuData *AccumData = &AccumData{
		data: make(map[string]SomeData),
	}

	//rand.Seed(time.Now().UnixNano())
	go monitor(readData, writeData, accuData)

	var w sync.WaitGroup

	for r := 0; r < n; r++ {
		w.Add(1)
		go func() {
			defer w.Done()
			set(&SomeData{payload: rand.Intn(10 * n)}, writeData)
		}()
	}
	w.Wait()

	Ilogger.Trace().Msg(fmt.Sprintf("mcli: Last value : %v\n", get(readData).payload))
	Ilogger.Trace().Msg(fmt.Sprintf("mcli: data : %v\n", accuData.data))

	// closure variables - danger in gorutines
	// for i := 1; i < 21; i++ {
	// 	go func(i int) {
	// 		fmt.Print(i, " ")
	// 	}(i)
	// }
	// time.Sleep(2 * time.Second)
	// fmt.Println()
}

// rootCmd represents the base command when running without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcli",
	Short: "cli tool for some operations in Linux and Windows",
	Long: `Yes there is an standart tools is 
But self made is more clearer and more manageble
`,
	Run: rootCmdRunFunc,
}

// Execute adds view child commands to the root command and sets flags appropriately.
// This is cviewed by main.main(). It only needs to happen once to the rootCmd.
func Execute(loggers []zerolog.Logger) {
	Ilogger, Elogger = loggers[0], loggers[1]
	// Elogger.Error().Msg("Some Test Error")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	var err error

	// Check if piped to StdIn

	info, _ := os.Stdin.Stat()
	GlobalMap["IS_COMMAND_IN_PIPE"] = "CommandNotInPipe"
	if ((info.Mode()&os.ModeNamedPipe) == os.ModeNamedPipe || info.Size() > 0) || len(InputDataFromFile) > 0 {
		var r *bufio.Reader
		var inputSlice []string = []string{}
		if len(InputDataFromFile) > 0 {
			file, err := os.Open(InputDataFromFile)
			if err != nil {
				Elogger.Fatal().Msgf("Input data file %s do not exists or not accessible", InputDataFromFile)
			}
			defer file.Close()
			r = bufio.NewReader(file)
			GlobalMap["IS_COMMAND_IN_PIPE"] = "CommandNotInPipeInputFromFile"
		} else {
			GlobalMap["IS_COMMAND_IN_PIPE"] = "CommandInPipe"
			r = bufio.NewReader(os.Stdin)
		}
		for {
			input, err := r.ReadString('\n')
			if input != "" {
				inputSlice = append(inputSlice, input)
			}
			if err != nil && err == io.EOF {
				break
			}
		}
		Input.InputSlice = inputSlice
	}

	// process root parameters
	configFile, _ := rootCmd.Flags().GetString("config")

	RedisHost, _ = rootCmd.Flags().GetString("redis-host")
	isRedisHostSet := rootCmd.Flags().Lookup("redis-host").Changed
	if !isRedisHostSet && len(os.Getenv("REDIS_HOST")) > 0 {
		RedisHost = os.Getenv("REDIS_HOST")
	}
	if isRedisHostSet || len(os.Getenv("REDIS_HOST")) == 0 {
		os.Setenv("REDIS_HOST", RedisHost)
	}

	RedisPort, _ = rootCmd.Flags().GetString("redis-port")
	isRedisPortSet := rootCmd.Flags().Lookup("redis-port").Changed
	if !isRedisPortSet && len(os.Getenv("REDIS_PORT")) > 0 {
		RedisHost = os.Getenv("REDIS_PORT")
	}
	if isRedisPortSet || len(os.Getenv("REDIS_PORT")) == 0 {
		os.Setenv("REDIS_PORT", RedisPort)
	}

	RedisPwd, _ = rootCmd.Flags().GetString("redis-password")
	isRedisPwdSet := rootCmd.Flags().Lookup("redis-password").Changed
	if !isRedisPwdSet && len(os.Getenv("REDIS_PWD")) > 0 {
		RedisHost = os.Getenv("REDIS_PWD")
	}
	if isRedisPwdSet || len(os.Getenv("REDIS_PWD")) == 0 {
		os.Setenv("REDIS_PWD", RedisPwd)
	}

	// read config file
	if len(configFile) == 0 {
		configFile = GlobalMap["DefaultConfigPath"]
	}
	if configFile != "" {
		Ilogger.Trace().Msg(fmt.Sprint("parsing config file:", configFile))

		if _, err := os.Stat(configFile); err == nil {
			configContent, err := os.ReadFile(configFile)
			configContentString := string(configContent)

			templateRegExp := regexp.MustCompile(`{{\$.+?}}`)
			allVarsEntries := mcli_utils.RemoveDuplicatesStr(templateRegExp.FindAllString(configContentString, -1))
			for _, varEntry := range allVarsEntries {
				// fmt.Println(varEntry)
				if strings.HasSuffix(varEntry, "$}}") {

					mapkey := strings.ReplaceAll(varEntry, "{{$", "")
					mapkey = strings.ReplaceAll(mapkey, "$}}", "")
					configContentString = strings.ReplaceAll(configContentString, varEntry, GlobalMap[mapkey])
				}
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
			Ilogger.Trace().Msg("config file " + configFile + " does not exist")
		} else {
			Elogger.Trace().Msg("config file detect error " + err.Error())
		}
	}

	Config.Cache = mcli_utils.NewCCache(0, 0, nil)

	// read or create key for internal secrets
	var rootKeySecretStorePath = filepath.Dir(Config.Common.InternalKeyFilePath)
	var rootSecretStore_key = Config.Common.InternalKeyFilePath

	if len(Config.Common.InternalKeyFilePath) == 0 {
		rootKeySecretStorePath = filepath.Join(GlobalMap["HomeDir"], ".mcli", "root")
		rootSecretStore_key = filepath.Join(rootKeySecretStorePath, "rootkey.key")
		Config.Common.InternalKeyFilePath = rootSecretStore_key
	}
	_, _, err = mcli_utils.IsExistsAndCreate(rootKeySecretStorePath, true, false)
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

	// fmt.Println("-------------")
	// fmt.Println("Global Map :", GlobalMap)
	Ilogger.Trace().Msgf("Global Map: %v", GlobalMap)
	// fmt.Println("-------------")
	// fmt.Println("Global Cache :", Config.Cache)
	Ilogger.Trace().Msgf("Global Cache: %v", Config.Cache)
}
