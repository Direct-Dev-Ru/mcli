/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_redis "mcli/packages/mcli-redis"
	mcli_secrets "mcli/packages/mcli-secrets"
	mcli_utils "mcli/packages/mcli-utils"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func Set(newData *SomeData, wc chan *SomeData) {
	wc <- newData
}

func Get(rc chan *SomeData) *SomeData {
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
			Set(&SomeData{payload: rand.Intn(10 * n)}, writeData)
		}()
	}
	w.Wait()

	// Ilogger.Trace().Msg(fmt.Sprintf("mcli: Last value : %v\n", Get(readData).payload))
	// Ilogger.Trace().Msg(fmt.Sprintf("mcli: data : %v\n", accuData.data))

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
	Short: "This is set of cli tool for some operations in Linux and Windows(not tested properly)",
	Long: `	Yes there is an Unix pattern: one thing one programm
		    But this is just top level link of other things
	`,
	Run: rootCmdRunFunc,
}

// Execute adds view child commands to the root command and sets flags appropriately.
// This is cviewed by main.main(). It only needs to happen once to the rootCmd.
func Execute(loggers []zerolog.Logger) {
	Ilogger, Elogger = loggers[0], loggers[1]

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	if Ctx != nil {
		defer func() {
			Ilogger.Trace().Msg("Ctx context now will be canceled - extra handling done")
			CtxCancel()
			time.Sleep(1 * time.Second)
		}()
	}

	if CommonRedisStore != nil {
		defer func() {
			Ilogger.Trace().Msg("Closing Common Redis Pool")
			CommonRedisStore.RedisPool.Close()
		}()
	}
}

func initConfig() {

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

	configFile, _ := rootCmd.Flags().GetString("config")
	// read config file
	ReadConfigFile(configFile)

	// now Config var is filled

	if Config.Common.AppName == "" {
		Config.Common.AppName = "default_mcli"
	}

	var err error
	IsRedis, _ = ProcessBoolCommandParameter("is-redis", Config.Common.RedisRequire, rootCmd)
	if IsRedis {

		// RedisHost, _ = ProcessCommandParameter("redis-host", "REDIS_HOST", rootCmd)
		err = TProcessCommandParameter[string](&RedisHost, "redis-host", "REDIS_HOST", rootCmd)
		if err != nil {
			Elogger.Debug().Err(err)
		}
		// RedisPort, _ = ProcessCommandParameter("redis-port", "REDIS_PORT", rootCmd)
		err = TProcessCommandParameter[string](&RedisPort, "redis-port", "REDIS_PORT", rootCmd)
		if err != nil {
			Elogger.Debug().Err(err)
		}
		// RedisPwd, _ = ProcessCommandParameter("redis-password", "REDIS_PWD", rootCmd)
		err = TProcessCommandParameter[string](&RedisPwd, "redis-password", "REDIS_PWD", rootCmd)
		if err != nil {
			Elogger.Debug().Err(err)
		}
		err = TProcessCommandParameter[int](&RedisDb, "redis-db", "REDIS_DB", rootCmd)
		if err != nil {
			Elogger.Debug().Err(err)
		}

		if Config.Common.RedisHost == "" {
			Config.Common.RedisHost = fmt.Sprintf("%s:%s", RedisHost, RedisPort)
		} else if Config.Common.RedisHost == ":" {
			Config.Common.RedisHost = fmt.Sprintf("%s:%s", "localhost", "6379")
		}

		if Config.Common.RedisPwd == "" {
			Config.Common.RedisPwd = RedisPwd
		}

		// common redis connection init
		CommonRedisStore, err = mcli_redis.NewRedisStore("rediscommon_"+Config.Common.AppName, Config.Common.RedisHost, Config.Common.RedisPwd,
			Config.Common.AppName, Config.Common.RedisDatabaseNo)
		if Config.Common.RedisRequire && err != nil {
			Elogger.Fatal().Msg(fmt.Sprintf("error init redis store: %v\n", err.Error()))
		}

		_, err = CommonRedisStore.RedisPool.Get().Do("PING")
		if Config.Common.RedisRequire && err != nil {
			Elogger.Fatal().Msgf("redis connection error: %v", err.Error())
		}
		if err == nil {
			Ilogger.Trace().Msg("Ping Pong to common Redis server is successful")
		}

		// getting or generating enc key for encryption records in redis database
		internalSecretStore := mcli_secrets.NewSecretsEntries(mcli_fs.GetFile, mcli_fs.SetFile, mcli_crypto.AesCypher, nil)

		if err := internalSecretStore.FillStore(Config.Common.InternalVaultPath, Config.Common.InternalKeyFilePath); err != nil {
			Elogger.Fatal().Msgf("error filling secret store %v", err)
		}
		secretMap := internalSecretStore.GetSecretPlainMap()
		redisEncKey := ""
		redisEncKeySecret, ok := secretMap["RedisEncKey"]

		if ok {
			redisEncKey = redisEncKeySecret.Secret
			_ = redisEncKey
			// Ilogger.Trace().Msgf("redis encryption key have been retrived from store: %s", fmt.Sprintf("%x", redisEncKey))
		} else {
			redisEncKey = string(mcli_secrets.GenKey(64))

			redisSecretKeyEntry, err := internalSecretStore.NewEntry("RedisEncKey", "RedisEncKey", "Key fo redis records encryption")
			if err != nil {
				Elogger.Fatal().Msgf("redisEncKey new entry creation error: %v", err)
			}
			redisSecretKeyEntry.SetSecret(fmt.Sprintf("%x", redisEncKey), true, false)
			// Ilogger.Trace().Msgf("redis encryption key have been generated: %s", fmt.Sprintf("%x", redisEncKey))

			internalSecretStore.AddEntry(redisSecretKeyEntry)
			if err != nil {
				Elogger.Fatal().Msgf("secret store add entry error: %v", err)
			}

			internalSecretStore.Save(Config.Common.InternalVaultPath, Config.Common.InternalKeyFilePath)
			if err != nil {
				Elogger.Fatal().Msgf("secret store save error: %v", err)
			}
		}

		// if CommonRedisStore != nil {
		// 	defer func() {
		// 		fmt.Println("defer: close common redis pool")
		// 		CommonRedisStore.RedisPool.Close()
		// 	}()
		// }

		//end of common redis connection init
	}

	Ctx, CtxCancel = context.WithCancel(context.Background())

	Config.Cache = mcli_utils.NewCCache(0, 0, nil, Ctx, Notify)
	GlobalCache = *mcli_utils.NewCCache(0, 0, nil, Ctx, Notify)

	Config.Cache.Set("Ctx", nil, 0, Ctx)
	Config.Cache.Set("CtxCancel", nil, 0, CtxCancel)

	InitInternalSecreVault(&Config)

	// TODO:Hide passwords

	if IsVerbose {
		Ilogger.Trace().Msgf("Global Map: %v", GlobalMap)

		Ilogger.Trace().Msgf("Global Cache: %v", Config.Cache)
	}
}
