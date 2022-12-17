package mclifilesystem

import (
	"os"
)

// https://www.devdungeon.com/content/working-files-go

type SetFileContentType func([]string) ([]string, error)

var SetFile SetFileContentType = func(in []string) ([]string, error) {
	return in, nil
}

func (sc SetFileContentType) TarToFile(source, target string) error {
	// fmt.Println(sc([]string{source, target}))
	return TarToFile(source, target)
}

func (sc SetFileContentType) SetContent(filePath string, content []byte) (int, error) {
	err := os.WriteFile(filePath, content, 0644)
	return len(content), err
}
