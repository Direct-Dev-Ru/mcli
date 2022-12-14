package mcliutils

import "strings"

func RemoveDuplicatesStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func SubString(str string, start, end int) string {
	ln := len(str)
	if ln == 0 || start > end {
		return ""
	}
	if end > ln {
		end = ln
	}
	if start < 0 {
		start = 0
	}
	runeSlice := []rune(str)
	return string(runeSlice[start:end])

}

func SubStringFind(str, start, end string) string {
	ln := len(str)
	if ln == 0 {
		return ""
	}
	var nStart, nEnd int
	nStart, nEnd = 0, ln
	if len(start) > 0 {
		nStart = strings.Index(str, start)
		if nStart == -1 {
			nStart = ln
		}
	}

	if len(end) > 0 {
		if start == end {
			nEnd = strings.LastIndex(str, end)
		} else {
			nEnd = strings.Index(str, end)
		}
		if nEnd == -1 {
			nEnd = ln
		}
	}
	return SubString(str, nStart+1, nEnd)
}

func SliceStringByPositions(stringToSlice string, positions []int) []string {
	row := make([]string, 0, len(positions))
	runeSlice := []rune(stringToSlice)
	for p := 0; p < len(positions); p++ {
		var cell string = ""
		if p == 0 {
			cell = string(runeSlice[:positions[p+1]])
		}
		if p == len(positions)-1 {
			cell = string(runeSlice[positions[p]:])
		}
		if p < len(positions)-1 && p > 0 {
			cell = string(runeSlice[positions[p]:positions[p+1]])
		}
		row = append(row, strings.TrimSpace(cell))
	}
	return row
}
