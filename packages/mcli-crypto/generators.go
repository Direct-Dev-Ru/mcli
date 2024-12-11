package mclicrypto

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_utils "mcli/packages/mcli-utils"
	"os"
	"strings"
	"unicode"
)

//go:embed defaultWords.csv
var embeddedDefaultWordsFile []byte

func GetDictionary(path string) (map[string][]string, error) {
	filter := func(current []string) bool {
		return len(current[0]) > 3 && !strings.Contains(current[0], "-")
	}
	rawDictionary, err := mcli_fs.ReadCsv(path, '\t', filter)
	if err != nil {
		return nil, err
	}
	dict := make(map[string][]string)

	for _, row := range rawDictionary {
		dict[row[1]] = append(dict[row[1]], row[0])
	}
	return dict, nil
}

func GetDictionaryFromBytes(fileData []byte) (map[string][]string, error) {
	if len(fileData) == 0 {
		fileData = embeddedDefaultWordsFile
	}
	content := string(fileData)
	lines := strings.Split(content, "\n")

	passDict := make(map[string][]string)

	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 || strings.Contains(fields[0], "-") {
			continue
		}

		key := fields[1]
		value := fields[0]

		passDict[key] = append(passDict[key], value)
	}
	return passDict, nil
}

type ReplaceEntry struct {
	OriginRune  rune
	ReplaceRune rune
	Number      int
}

// Generates passphrase from list of words given on path wordsListPath
// replaces - slice with replace structs
func GeneratePassPhrase(wordsListPath string, replaces []ReplaceEntry, translit, rutoenQwert bool) (string, error) {
	if len(wordsListPath) == 0 {
		wordsListPath = "./pwdgen/freqrnc2011.csv"
	}
	var passDict map[string][]string
	if _, err := os.Stat(wordsListPath); os.IsNotExist(err) {
		passDict, err = GetDictionaryFromBytes(embeddedDefaultWordsFile)
		if err != nil {
			return "", err
		}
	} else {
		passDict, err = GetDictionary(wordsListPath)
		if err != nil {
			return "", err
		}
	}
	words := map[string]string{"s": "", "v": "", "adv": "", "a": ""}
	for t := range words {
		list, ok := passDict[t]

		if ok && len(list) > 0 {
			words[t] = list[mcli_utils.Random(0, len(list)-1)]
		}
	}
	a := []rune(words["a"])
	s := []rune(words["s"])
	adv := []rune(words["adv"])
	v := []rune(words["v"])
	n := mcli_utils.Random(1, len(passDict["a"]))

	phrase := fmt.Sprintf("%s%s%s%s",
		string(append([]rune{unicode.ToUpper(a[0])}, a[1:]...)),
		string(append([]rune{unicode.ToUpper(s[0])}, s[1:]...)),
		string(append([]rune{unicode.ToUpper(v[0])}, v[1:]...)),
		string(append([]rune{unicode.ToUpper(adv[0])}, adv[1:]...)),
	)
	if translit {
		phrase = mcli_utils.TranslitToLatFromCyr(phrase)
	}
	if rutoenQwert {
		phrase = mcli_utils.RuToEnKeyboardLayout(phrase)
	}
	phrase = fmt.Sprintf("%s%v", phrase, n)

	// Process replacements
	if len(replaces) > 0 {
		for _, r := range replaces {
			phrase = strings.Replace(phrase, string(r.OriginRune), string(r.ReplaceRune), r.Number)
		}
	}

	return phrase, nil
}

func GenerateBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GeneratePassword(length int) (string, error) {
	if length == 0 {
		length = 12
	}
	b, err := GenerateBytes(length)
	p := Base64Encode(string(b))
	p = strings.ReplaceAll(p, "=", "")
	r := []rune(p)
	if len(r) > length {
		p = string(r[:length])
	}
	return p, err
}
