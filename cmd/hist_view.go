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

func readBashHistory() []bashHistEntry {
	filepath := os.Getenv("HOME") + "/.bash_history"

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

// for scanner.Scan() {
// 	currentLine = scanner.Text()
// 	if !strings.HasPrefix(currentLine, "#") {
// 		currentLine = strings.TrimSpace(currentLine)
// 		fmt.Println(currentLine, "(", tm, ")")
// 	} else {
// 		currentTs = strings.TrimPrefix(currentLine, "#")
// 		i, err := strconv.ParseInt(currentTs, 10, 64)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		tm = time.Unix(i, 0)
// 	}

// }
// if err := scanner.Err(); err != nil {
// 	log.Fatal(err)
// }

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// contents, _ := ioutil.ReadFile(os.Getenv("HOME") + "/.bash_history")

		filter, _ := cmd.Flags().GetString("filter")
		Ilogger.Trace().Msg("Filter is :" + filter)
		for ind, entry := range readBashHistory() {
			cmd, tm := entry.commandText, entry.commandTime
			if strings.Contains(cmd, filter) {
				fmt.Println(ind, cmd, " | ", tm.Local().String(), " | ")
			}
		}
	},
}

func init() {
	histCmd.AddCommand(viewCmd)
	viewCmd.Flags().String("filter", "", "filter for output")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and view subcommands, e.g.:
	// viewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is cviewed directly, e.g.:
	// viewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
