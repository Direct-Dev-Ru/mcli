package mclifilesystem

import (
	"encoding/csv"
	"errors"
	"os"
	"strings"
)

type FilterRow func(current []string) bool

// ReadCsv reads csv from path parameter separated by comma parameter separated
// keyIndex in csv fields - for map keys
// filter - func to filter rows: func(current []string) bool (or nil)
func ReadCsv(path string, comma rune, filter FilterRow) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("cannot open CSV file:" + err.Error())
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = comma
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, errors.New("cannot open CSV file:" + err.Error())
	}
	result := make([][]string, 0, len(rows))

	for _, row := range rows {

		if len(row) == 1 {
			row = strings.Fields(row[0])
		}
		if len(row) >= 2 {
			if filter != nil {
				if filter(row) {
					result = append(result, row)
				}
				continue
			}
		}
	}
	return result, nil
}
