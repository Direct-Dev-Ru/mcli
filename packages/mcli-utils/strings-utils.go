package mcliutils

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in SliceStringByPositions", r)
		}
	}()
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

var cyrAlfabet []string = []string{"а", "б", "в", "г", "д", "е", "ё", "ж", "з", "и", "й", "к", "л", "м", "н", "о", "п", "р", "с", "т", "у", "ф", "х", "ц", "ч", "ш", "щ", "ь", "ъ", "ы", "э", "ю", "я"}
var latAlfabet []string = []string{"a", "b", "v", "g", "d", "e", "yo", "zh", "z", "i", "y", "k", "l", "m", "n", "o", "p", "r", "s", "t", "u", "f", "h", "c", "ch", "sh", "sch", "", "", "y", "e", "yu", "ya"}
var cyrBigAlfabet []string = []string{"А", "Б", "В", "Г", "Д", "Е", "Ё", "Ж", "З", "И", "Й", "К", "Л", "М", "Н", "О", "П", "Р", "С", "Т", "У", "Ф", "Х", "Ц", "Ч", "Ш", "Щ", "Ь", "Ъ", "Ы", "Э", "Ю", "Я"}
var latBigAlfabet []string = []string{"A", "B", "V", "G", "D", "E", "Yo", "Zh", "Z", "I", "Y", "K", "L", "M", "N", "O", "P", "R", "S", "T", "U", "F", "H", "C", "Ch", "Sh", "Sch", "", "", "Y", "E", "Yu", "Ya"}

func TranslitToLatFromCyr(stringToConvert string) string {

	output := ""
	cyr := append(cyrAlfabet, cyrBigAlfabet...)
	lat := append(latAlfabet, latBigAlfabet...)
	for _, r := range stringToConvert {
		s := string(r)
		if ind := slices.Index(cyr, s); ind >= 0 {
			output += lat[ind]
		}
	}
	return string(output)
}

func FindSubstrings(text, pattern string) ([]string, bool) {
	if pattern == "" {
		pattern = `{{(.*?)}}`
	}
	re := regexp.MustCompile(pattern)
	found := re.FindAllStringSubmatch(text, -1)

	if len(found) > 0 {
		substrings := make([]string, len(found))
		for i, f := range found {
			substrings[i] = f[0]
		}
		return substrings, true
	}

	return nil, false
}

func IsStringEmpty(s string) bool {
	return len(s) == 0
}

func IsStringNotEmpty(s string) bool {
	return !(len(s) == 0)
}

// hasIdenticalSymbols checks if all characters in a string are identical
func HasIdenticalSymbols(s string) bool {
	if len(s) <= 1 {
		return true
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	firstSymbol := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != firstSymbol {
			return false
		}
	}
	return true
}

func PadLeft(input string, length int, padChar rune) string {
	if len(input) >= length {
		return input
	}

	padding := strings.Repeat(string(padChar), length-len(input))
	return padding + input
}

func PadRight(input string, length int, padChar rune) string {
	if len(input) >= length {
		return input
	}

	padding := strings.Repeat(string(padChar), length-len(input))
	return input + padding
}
