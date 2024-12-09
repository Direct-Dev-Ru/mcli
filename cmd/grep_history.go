/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/eiannone/keyboard"
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
		if checkUnixTime && strings.HasPrefix(currentCommand, "#") &&
			!strings.HasPrefix(currentCommand, "#!") {
			// timeIsOn = true
			currentTs = strings.TrimPrefix(currentCommand, "#")
			i, err := strconv.ParseInt(currentTs, 10, 64)
			if err != nil {
				// Elogger.Fatal().Msgf("grep history: %v [%v]", err, currentCommand)
				tm = time.Unix(0, 0)
			} else {
				tm = time.Unix(i, 0)
			}
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
		// Elogger.Fatal().Msgf("grep history: %v", err)
		return nil, err
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
	_ = emptyTm
	historyEntries, err := readHistory(fileToGrep)
	if err != nil {
		Elogger.Fatal().Msgf("grep history: %v", err)
	}
	// Sort entries
	sort.Slice(historyEntries, func(i, j int) bool {
		return historyEntries[i].commandNumber < historyEntries[j].commandNumber
	})
	// Remove duplicates
	seen := make(map[string]struct{})
	deDupEntries := []bashHistEntry{}
	for _, entry := range historyEntries {
		if _, exists := seen[entry.commandText]; !exists {
			seen[entry.commandText] = struct{}{}
			deDupEntries = append(deDupEntries, entry)
		}
	}
	filteredEntries := filterByRegex(deDupEntries, filter)
	if len(filteredEntries) == 0 {
		fmt.Println("No commands matched the filter.")
		return
	}

	// for _, entry := range filteredEntries {
	// 	numCmd, cmdText, tm := entry.commandNumber, entry.commandText, entry.commandTime
	// 	if tm.UnixNano() > emptyTm.UnixNano() {
	// 		fmt.Fprintf(cmd.OutOrStdout(), "%d %s | %v | \n", numCmd, cmdText, tm)
	// 	} else {
	// 		fmt.Fprintf(cmd.OutOrStdout(), "%d %s \n", numCmd, cmdText)
	// 	}
	// }
	// Display and navigate
	navigateAndCopy(filteredEntries)
}

func navigateAndCopy(entries []bashHistEntry) {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	index := len(entries) - 1 // Start from the last entry

	for {
		// Clear screen and display the current command
		fmt.Print("\033[H\033[2J") // ANSI escape codes to clear the terminal
		fmt.Printf("Command %d/%d: %s\n", index+1, len(entries), entries[index].commandText)
		fmt.Println("\nNavigate with 'w' (up), 's' (down). Press 'c' to copy to clipboard, 'q' to quit.")

		// Capture key press
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("Error reading key: %v\n", err)
			return
		}

		switch char {
		case 'w': // Up
			if index > 0 {
				index--
			}
		case 's': // Down
			if index < len(entries)-1 {
				index++
			}
		case 'c': // Copy
			clipboard.WriteAll(entries[index].commandText)
			fmt.Println("Copied to clipboard!")
		case 'q': // Quit
			fmt.Println("Goodbye!")
			return
		}

		// Handle special keys (if needed)
		if key == keyboard.KeyEsc {
			fmt.Println("Exiting...")
			return
		}
	}
}

func _navigateAndCopy(entries []bashHistEntry) {
	index := len(entries) - 1 // Start from the last entry

	for {
		// Clear screen and display the current command
		fmt.Print("\033[H\033[2J") // ANSI escape codes to clear the terminal
		fmt.Printf("Command %d/%d: %s\n", index+1, len(entries), entries[index].commandText)
		fmt.Println("\nNavigate with 'w' (up), 's' (down). Press 'c' to copy to clipboard, 'q' to quit.")

		// Get user input
		var input string
		fmt.Scan(&input)

		switch strings.ToLower(input) {
		case "w": // Up
			if index > 0 {
				index--
			}
		case "s": // Down
			if index < len(entries)-1 {
				index++
			}
		case "c": // Copy
			clipboard.WriteAll(entries[index].commandText)
			fmt.Println("Copied to clipboard!")
		case "q": // Quit
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid input, try again.")
		}
	}
}

func filterByRegex(entries []bashHistEntry, regex string) []bashHistEntry {
	re, err := regexp.Compile(regex)
	if err != nil {
		fmt.Printf("Invalid regex: %v\n", err)
		return entries // Return unfiltered entries if regex is invalid
	}

	filtered := []bashHistEntry{}
	for _, entry := range entries {
		if re.MatchString(entry.commandText) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
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
