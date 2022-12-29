package mclicrypto

import (
	"crypto/rand"
	"fmt"
	mcli_fs "mcli/packages/mcli-filesystem"
	mcli_utils "mcli/packages/mcli-utils"
	"strings"
	"unicode"
)

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

type ReplaceEntry struct {
	OriginRune  rune
	ReplaceRune rune
	Number      int
}

// Generates passphrase from list of words given on path wordsListPath
// replaces - slice with replace structs
func GeneratePassPhrase(wordsListPath string, replaces []ReplaceEntry) (string, error) {
	if len(wordsListPath) == 0 {
		wordsListPath = "./pwdgen/freqrnc2011.csv"
	}
	passDict, err := GetDictionary(wordsListPath)
	if err != nil {
		return "", err
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

	phrase := fmt.Sprintf("%s%s%s%s",
		string(append([]rune{unicode.ToUpper(a[0])}, a[1:]...)),
		string(append([]rune{unicode.ToUpper(s[0])}, s[1:]...)),
		string(append([]rune{unicode.ToUpper(adv[0])}, adv[1:]...)),
		string(append([]rune{unicode.ToUpper(v[0])}, v[1:]...)))
	phrase = mcli_utils.TranslitToLatFromCyr(phrase)

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
	b, err := GenerateBytes(length)
	p := Base64Encode(string(b))
	p = strings.ReplaceAll(p, "=", "")
	return p, err
}
