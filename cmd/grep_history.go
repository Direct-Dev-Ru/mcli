/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type bashHistEntry struct {
	commandNumber int
	commandText   string
	commandTime   time.Time
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
		Elogger.Fatal().Msgf("grep history: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currentCommand string
	var currentTs string
	var tm time.Time
	// var timeIsOn bool = false
	bashCommands := make([]bashHistEntry, 100)
	iteration := 1
	for scanner.Scan() {
		currentCommand = scanner.Text()
		if strings.HasPrefix(currentCommand, "#") {
			// timeIsOn = true
			currentTs = strings.TrimPrefix(currentCommand, "#")
			i, err := strconv.ParseInt(currentTs, 10, 64)
			if err != nil {
				Elogger.Fatal().Msgf("grep history: %w", err)
			}
			tm = time.Unix(i, 0)
		} else {
			currentCommand = strings.TrimSpace(currentCommand)
			isDuplicated := false
			for k, v := range bashCommands {
				if currentCommand == v.commandText {
					bashCommands[k].commandNumber = iteration
					bashCommands[k].commandTime = tm
					isDuplicated = true
					break
				}
			}
			if !isDuplicated {
				bashCommands = append(bashCommands,
					bashHistEntry{commandNumber: iteration, commandText: currentCommand, commandTime: tm})
			}
			iteration++
		}

	}
	if err := scanner.Err(); err != nil {
		Elogger.Fatal().Msgf("grep history: %w", err)
	}
	return bashCommands
}

func historyRun(cmd *cobra.Command, args []string) {

	filter, _ := cmd.Flags().GetString("filter")
	Ilogger.Trace().Msg("Filter is: " + filter)
	var emptyTm time.Time
	for _, entry := range readHistory() {
		numCmd, cmd, tm := entry.commandNumber, entry.commandText, entry.commandTime
		if strings.Contains(cmd, filter) {
			if tm.UnixNano() > emptyTm.UnixNano() {
				fmt.Println(numCmd, cmd, " | ", tm.Local().String(), " | ")
			} else {
				fmt.Println(numCmd, cmd)
			}
		}
	}
}

// historyCmd represents the view command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "grep terminal history",
	Long: `	prints terminal history and optionally filters it. 
			for example: mcli grep history --filter docker
`,
	Run: historyRun,
}

func init() {
	grepCmd.AddCommand(historyCmd)
	historyCmd.Flags().StringP("filter", "f", "", "filter for output")
}
