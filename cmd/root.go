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
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	mcli_utils "mcli/packages/mcli-utils"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type InputData struct {
	inputSlice  []string
	joinedInput string
}

// Global Vars
var Ilogger, Elogger zerolog.Logger
var ConfigPath string
var RootPath string
var GlobalMap map[string]string = make(map[string]string)

var Version string = "1.0.9"
var Input InputData = InputData{inputSlice: []string{}, joinedInput: ""}

// var Config map[string]interface{}
var Config ConfigData = ConfigData{}

type runFunc func(cmd *cobra.Command, args []string)

type SomeData struct {
	payload int
	// err     error
}
type AccumData struct {
	sync.Mutex
	data map[string]SomeData
}

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

var rootCmdRunFunc runFunc = func(cmd *cobra.Command, args []string) {
	config, _ := cmd.Flags().GetString("config")
	rootArgs, _ := cmd.Flags().GetString("root-args")
	Ilogger.Info().Msg("Hello from Multy CLI. Config is " + config)

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

	rand.Seed(time.Now().UnixNano())
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

func init() {

	cobra.OnInitialize(initConfig)

	_, callerPath, _, _ := runtime.Caller(0)
	RootPath = path.Dir(path.Dir(callerPath))
	GlobalMap["RootPath"] = RootPath

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", RootPath+"/.mcli.yaml",
		"specify config file - default "+RootPath+"/.mcli.yaml")

	rootCmd.Flags().StringP("root-args", "a", "", "args for root command")
}

func initConfig() {

	// Check if piped to StdIn
	info, _ := os.Stdin.Stat()
	if (info.Mode()&os.ModeNamedPipe) == os.ModeNamedPipe || info.Size() > 0 {

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

		// fmt.Println(len(inputSlice))
		// for i, v := range inputSlice {
		// 	fmt.Print(i, " ", v)
		// }

		joinedInput := strings.Join(inputSlice, " ")
		joinedInput = strings.ReplaceAll(joinedInput, "\r\n", "")
		joinedInput = strings.ReplaceAll(joinedInput, "\n", "")
		Input.inputSlice = inputSlice
		Input.joinedInput = joinedInput
		if len(Input.inputSlice) > 0 {
			Ilogger.Trace().Msg(fmt.Sprintf("\n%v\n", Input.inputSlice))
		}
	}

	// read config
	configFile, _ := rootCmd.Flags().GetString("config")
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
				_ = yaml.Unmarshal([]byte(configContentString), &Config)
			}
			// fmt.Println("Configuration content :", string(configContent))
			Ilogger.Trace().Msg(fmt.Sprint("Configuration struct :", Config))
		} else if errors.Is(err, os.ErrNotExist) {
			Ilogger.Trace().Msg("config file " + configFile + " does not exist")
		} else {
			Elogger.Trace().Msg("config file detect error " + err.Error())
		}
	}

	// fmt.Println(info.Mode(), info.Name(), info.Size())
}
