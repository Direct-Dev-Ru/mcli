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
	"golang.org/x/exp/slices"
)

type bashHistEntry struct {
	commandNumber int
	commandText   string
	commandTime   time.Time
}

func readHistory(fileToGrep string) ([]bashHistEntry, error) {
	var filepath string
	var checkUnixTime bool = false

	switch GlobalMap["OS"] {
	case "windows":
		filepath = os.Getenv("USERPROFILE") + "/AppData/Roaming/Microsoft/Windows/PowerShell/PSReadline/ConsoleHost_history.txt"
	case "darwin":
		filepath = os.Getenv("HOME") + "/.bash_history"
	case "linux":
		filepath = os.Getenv("HOME") + "/.bash_history"
		checkUnixTime = true
	default:
		filepath = os.Getenv("HOME") + "/.bash_history"
	}
	if fileToGrep != "" {
		filepath = fileToGrep
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
		// Elogger.Fatal().Msgf("grep history: %v", err)
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
		if checkUnixTime && strings.HasPrefix(currentCommand, "#") {
			// timeIsOn = true
			currentTs = strings.TrimPrefix(currentCommand, "#")
			i, err := strconv.ParseInt(currentTs, 10, 64)
			if err != nil {
				Elogger.Fatal().Msgf("grep history: %v", err)
			}
			tm = time.Unix(i, 0)
		} else {
			currentCommand = strings.TrimSpace(currentCommand)
			isDuplicated := false
			if searchRes := slices.IndexFunc(bashCommands,
				func(c bashHistEntry) bool {
					return currentCommand == c.commandText
				}); searchRes >= 0 {
				bashCommands[searchRes].commandNumber = iteration
				bashCommands[searchRes].commandTime = tm
				isDuplicated = true
			}

			if !isDuplicated {
				bashCommands = append(bashCommands,
					bashHistEntry{commandNumber: iteration, commandText: currentCommand, commandTime: tm})
			}
			iteration++
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
		// Elogger.Fatal().Msgf("grep history: %v", err)
	}
	return bashCommands, nil
}

// historyRun represents run function for grep commsnd history
func historyRun(cmd *cobra.Command, args []string) {
	var fileToGrep string
	if len(args) > 0 {
		fileToGrep = args[0]
	}
	filter, _ := cmd.Flags().GetString("filter")
	// Ilogger.Trace().Msg("Filter is: " + filter)
	var emptyTm time.Time
	historyEntries, err := readHistory(fileToGrep)
	if err != nil {
		Elogger.Fatal().Msgf("grep history: %v", err)
	}
	for _, entry := range historyEntries {
		numCmd, cmdText, tm := entry.commandNumber, entry.commandText, entry.commandTime
		if strings.Contains(cmdText, filter) {
			if tm.UnixNano() > emptyTm.UnixNano() {
				fmt.Fprintf(cmd.OutOrStdout(), "%d %s | %v | \n", numCmd, cmdText, tm)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%d %s \n", numCmd, cmdText)
			}
		}
	}
}

// historyCmd represents the history command
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
