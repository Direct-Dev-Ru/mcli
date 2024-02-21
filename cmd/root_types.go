package cmd

import (
	"fmt"
	mcliutils "mcli/packages/mcli-utils"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/spf13/cobra"
)

// --------Parametrable--------- //

type Parametrable interface {
	String() (string, bool)
	Int() (int, bool)
	Bool() (bool, bool)
}

type GenParameter struct {
	Value interface{}
}

func (v GenParameter) String() (ret string, ok bool) {
	ret, ok = v.Value.(string)
	return ret, ok

}
func (v GenParameter) Int() (ret int, ok bool) {
	ret, ok = v.Value.(int)
	return ret, ok
}
func (v GenParameter) Bool() (ret bool, ok bool) {
	ret, ok = v.Value.(bool)
	return ret, ok
}

// ---------

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
	OutputMap   map[string][]string
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

func (d *InputData) SplitByOneOrMoreSpaces(skipLines int, removeOneSymbolLines bool,
	useFirstLineAsHeader bool, useOneSpaceAsSplitter bool) (map[string][]string, error) {

	var inputLines [][]rune = make([][]rune, 0)
	// convert input lines to [][]rune
	if skipLines < 0 {
		skipLines = 0
	}
	for i := skipLines; i < len(d.InputSlice); i++ {
		inputLine := d.InputSlice[i]
		if mcliutils.HasIdenticalSymbols(inputLine) {
			continue
		}
		inputLines = append(inputLines, []rune(inputLine))
	}

	var fields []string
	var positions []int
	inField := true

	maxLengthLine := 0
	maxLineIdx := 0
	for idx, line := range inputLines {
		if len(line) > maxLengthLine {
			maxLengthLine = len(line)
			maxLineIdx = idx
		}
	}
	// fmt.Println(maxLineIdx)
	// fmt.Println(inputLines[maxLineIdx])
	baseLine := inputLines[maxLineIdx]

	// fmt.Println(baseLine, len(baseLine), maxLineIdx, maxLengthLine)
	// fmt.Println(string(baseLine), len(string(baseLine)), maxLineIdx, maxLengthLine)

	prevPos := 0
	isColumnBorderCandidate := false
	isColumnBorder := false
	fieldCandidate := ""
	for pos, char := range baseLine {

		if unicode.IsSpace(char) {
			if inField {
				// fmt.Println(pos)
				if !isColumnBorderCandidate {
					isColumnBorderCandidate = true
					for _, line := range inputLines {
						if len(line) > pos && !unicode.IsSpace(line[pos]) {
							isColumnBorderCandidate = false
							break
						}
					}
					if isColumnBorderCandidate {
						continue
					}
				}
				if isColumnBorderCandidate {
					isColumnBorder = true

					for _, line := range inputLines {
						if !unicode.IsSpace(line[pos]) {
							isColumnBorder = false
							break
						}
					}
					if !isColumnBorder {
						continue
					}

				}
				// fmt.Println(isColumnBorder, inField, prevPos)

				if isColumnBorder && inField {
					fieldCandidate = ""
					if useFirstLineAsHeader {
						fieldCandidate = strings.TrimSpace(string(inputLines[0][prevPos : pos+1]))
					}

					if len(fieldCandidate) > 0 || !useFirstLineAsHeader {
						// fmt.Println(fieldCandidate, inField)
						inField = false
						if useFirstLineAsHeader {
							fields = append(fields, fieldCandidate)
						} else {
							fields = append(fields, "Column"+mcliutils.PadLeft(strconv.Itoa(len(fields)+1), 2, '0'))
						}
						positions = append(positions, pos)
						prevPos = pos
						isColumnBorder = false
						isColumnBorderCandidate = false
					}
				}
			} else {
				for _, line := range inputLines {
					if !unicode.IsSpace(line[pos]) {
						inField = true
						break
					}
				}
				continue
			}
		} else {
			isColumnBorderCandidate = false
			isColumnBorder = false
			inField = true
		}
	}
	// fmt.Println(positions[len(positions)-1])

	if useFirstLineAsHeader {
		fieldCandidate = strings.TrimSpace(string(inputLines[0][positions[len(positions)-1]+1:]))
	}
	if len(fieldCandidate) > 0 || !useFirstLineAsHeader {
		// fmt.Println(fieldCandidate, inField)
		if useFirstLineAsHeader {
			fields = append(fields, fieldCandidate)
		} else {
			fields = append(fields, "Column"+mcliutils.PadLeft(strconv.Itoa(len(fields)+1), 2, '0'))
		}
		// positions = append(positions, 6500000)
	}

	table := make(map[string][]string, 0)

	// Initialize the map with empty slices
	for _, col := range fields {
		table[col] = make([]string, 0)
	}

	for lnum, line := range inputLines {
		// fmt.Println("-----------------------", lnum, "--------------------")
		// fmt.Println(string(line))
		// markerString := createStringWithCaretPositions(positions, len(line))
		// fmt.Println(markerString)
		if useFirstLineAsHeader && lnum == 0 {
			continue
		}

		prev := 0
		fieldValue := ""
		for fnum, pos := range positions {
			fieldValue = strings.TrimSpace(string(line[prev:pos]))
			prev = pos
			// fmt.Println(fnum, headers[fnum], "\t\t", fieldValue)
			table[fields[fnum]] = append(table[fields[fnum]], fieldValue)
		}
		fieldValue = strings.TrimSpace(string(line[prev:]))
		// fmt.Println(len(positions), headers[len(positions)], "\t\t", fieldValue)
		table[fields[len(positions)]] = append(table[fields[len(positions)]], fieldValue)

	}
	d.InputMap = table
	return table, nil
}

func (d *InputData) PrintAsTable(columnDivider string) {
	mcliutils.PrintAsTable(d.InputMap, columnDivider)
}

func (d *OutputData) PrintAsTable(columnDivider string) {
	mcliutils.PrintAsTable(d.OutputMap, columnDivider)
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
