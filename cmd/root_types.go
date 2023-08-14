package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

type TerminalParams struct {
	TermWidth  int
	TermHeight int
	IsTerminal bool
}

type InputData struct {
	// slice of input lines separatated by \n
	InputSlice []string
	// map then input is a table
	InputMap   map[string][]string
	InputTable []map[string]string
}

type OutputData struct {
	OutputSlice []string
	OutputMap   map[string]interface{}
	OutputTable []map[string]string
}

func (d InputData) DivideInputSlice(divider string, columnBorder rune) ([]string, error) {
	if len(divider) == 0 {
		divider = "|"
	}
	if len(d.InputSlice) == 0 {
		return nil, fmt.Errorf("input slice is null - nothing to do")
	}
	lenghts := make([]int, len(d.InputSlice))
	outSlice := make([]string, len(d.InputSlice))
	runeSlices := make([][]rune, 0, len(d.InputSlice))
	// first go through the input slice and calc lenths of strings
	maxIdx := 0
	maxLength := 0
	runeIdx := 0
	for _, line := range d.InputSlice {
		lineTrimmed := strings.TrimSpace(line)
		if !(len(lineTrimmed) == 0 || lineTrimmed == "\n" || lineTrimmed == "\r\n") {
			runeSlices = append(runeSlices, []rune(line))
			lenghts[runeIdx] = len(runeSlices[runeIdx])
			if maxLength < lenghts[runeIdx] {
				maxIdx = runeIdx
				maxLength = lenghts[runeIdx]
			}
			runeIdx++
		}
	}
	// for _, v := range runeSlices {
	// 	fmt.Println(len(v), ":", string(v))
	// 	for _, rune := range v {
	// 		fmt.Print(rune, "'", string(rune), "' ")
	// 	}
	// 	fmt.Println("-----")

	// }
	// go char by char and if current char is space - check if it is space in all lines and replace with |
	lineToIterate := runeSlices[maxIdx]
	// fmt.Println("---------------line to iterate---------------")
	// fmt.Println(string(lineToIterate))
	// fmt.Println(lineToIterate)
	// fmt.Println("---------------end line to iterate---------------")
	flDivided := false
	_ = flDivided
	flColAllSpaces := false
	runeColCheck1 := columnBorder
	//runeColCheck2 := '\t'
	for ind, rune := range lineToIterate {
		// fmt.Printf("rune: %v(%c) ::: flColAllSpaces: %v ::: flDivided: %v\n", rune, rune, flColAllSpaces, flDivided)

		if rune == runeColCheck1 {
			// now we check that in every line space is
			flColAllSpaces = true
			for _, runeSlice := range runeSlices {
				if ind < len(runeSlice) {
					if runeSlice[ind] != runeColCheck1 {
						flColAllSpaces = false
						flDivided = false
						break
					}
				}
			}
			fillWith := ""
			if flColAllSpaces && flDivided {
				flDivided = false
			}

			if flColAllSpaces {
				fillWith = divider
				flDivided = true
				for lInd, runeSlice := range runeSlices {
					if ind < len(runeSlice) {
						outSlice[lInd] = outSlice[lInd] + fillWith
					}
				}
			} else {
				for lInd, runeSlice := range runeSlices {
					if ind < len(runeSlice) {
						outSlice[lInd] = outSlice[lInd] + string(runeSlice[ind])
					}
				}
			}

		} else {
			for lInd, runeSlice := range runeSlices {
				if ind < len(runeSlice) {
					outSlice[lInd] = outSlice[lInd] + string(runeSlice[ind])
				}
			}
		}
		// fmt.Printf("rune: %v(%c) ::: flColAllSpaces: %v ::: flDivided: %v\n", rune, rune, flColAllSpaces, flDivided)
		// fmt.Println("--------------------------------------------------------------------------")
	}
	// fmt.Println(d.InputSlice)
	// fmt.Println(runeSlices)

	// fmt.Println(outSlice)
	return outSlice, nil
}

func (d InputData) GetJoinedString(strJoin string, removeLineBreaks bool) (string, error) {
	joinedInput := strings.Join(d.InputSlice, strJoin)

	if removeLineBreaks {
		joinedInput = strings.ReplaceAll(joinedInput, "\r\n", "")
		joinedInput = strings.ReplaceAll(joinedInput, "\n", "")
	}
	return joinedInput, nil
}

type runFunc func(cmd *cobra.Command, args []string)

type SomeData struct {
	payload int
	// err     error
}
type AccumData struct {
	sync.Mutex
	data map[string]SomeData
}
