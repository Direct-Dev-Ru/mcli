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
	"regexp"
	"strconv"
	"strings"
	"sync"

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

// rootCmd represents the base command when cviewed without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcli",
	Short: "cli for some operations in Linux",
	Long: `Yes there is an aliases is ...
But standalone executable module sometimes 
is more helpful than .bashrc file`,

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

	// Check if piped to StdIn
	info, _ := os.Stdin.Stat()
	GlobalMap["IS_COMMAND_IN_PIPE"] = "CommandNotInPipe"
	if (info.Mode()&os.ModeNamedPipe) == os.ModeNamedPipe || info.Size() > 0 {
		GlobalMap["IS_COMMAND_IN_PIPE"] = "CommandInPipe"
		var inputSlice []string = []string{}
		r := bufio.NewReader(os.Stdin)
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
		// Input.DivideInputSlice("||", ' ')
	}

	// read config
	configFile, _ := rootCmd.Flags().GetString("config")
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

	// fmt.Println(info.Mode(), info.Name(), info.Size())
}
