/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type bashHistEntry struct {
	commandText string
	commandTime time.Time
}

func readHistory() []bashHistEntry {
	var filepath string
	switch GlobalMap["OS"] {
	case "windows":
		filepath = os.Getenv("USERPROFILE") + "/AppData/Roaming/Microsoft/Windows/PowerShell/PSReadline/ConsoleHost_history.txt"
	case "darwin":
		filepath = os.Getenv("HOME") + "/.bash_history"
	case "linux":
		filepath = os.Getenv("HOME") + "/.bash_history"
	default:
		filepath = os.Getenv("HOME") + "/.bash_history"
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currentCommand string
	var currentTs string
	var tm time.Time
	// var timeIsOn bool = false
	bashCommands := make([]bashHistEntry, 100)
	fmt.Println("time", tm)
	for scanner.Scan() {
		currentCommand = scanner.Text()
		if strings.HasPrefix(currentCommand, "#") {
			// timeIsOn = true
			currentTs = strings.TrimPrefix(currentCommand, "#")
			i, err := strconv.ParseInt(currentTs, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			tm = time.Unix(i, 0)
		} else {
			currentCommand = strings.TrimSpace(currentCommand)
			bashCommands = append(bashCommands, bashHistEntry{commandText: currentCommand, commandTime: tm})
		}

	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return bashCommands
}

// viewCmd represents the view command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "grep terminal history",
	Long: `	prints terminal history and optionally filters it. 
			for example: mcli grep history --filter docker
`,
	Run: func(cmd *cobra.Command, args []string) {
		// contents, _ := ioutil.ReadFile(os.Getenv("HOME") + "/.bash_history")
		filter, _ := cmd.Flags().GetString("filter")
		Ilogger.Trace().Msg("Filter is :" + filter)
		for ind, entry := range readHistory() {
			cmd, tm := entry.commandText, entry.commandTime
			if strings.Contains(cmd, filter) {
				fmt.Println(ind, cmd, " | ", tm.Local().String(), " | ")
			}
		}
	},
}

func init() {
	grepCmd.AddCommand(historyCmd)
	historyCmd.Flags().StringP("filter", "f", "", "filter for output")
}
