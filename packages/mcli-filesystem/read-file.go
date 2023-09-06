package mclifilesystem

import (
	"bufio"
	"errors"
	"io"
	"os"
)

// https://www.devdungeon.com/content/working-files-go

type GetFileContentType func([]string) ([]string, error)

var GetFile GetFileContentType = func(in []string) ([]string, error) {
	return in, nil
}

func (gc GetFileContentType) UntarFromFile(tarball, target string) error {
	return UntarFromFile(tarball, target)
}

func (gc GetFileContentType) GetContent(filePath string) ([]byte, error) {

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	result, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (gc GetFileContentType) GetContentChunks(filePath string) ([]byte, error) {
	f, closer, err := getFileForR(filePath)
	if err != nil {
		return nil, err
	}
	defer closer()

	var result []byte = make([]byte, 0, 1024*1024)
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		result = append(result, buf[:]...)
	}
	return result, nil
}

func ReadFileLines(filepath string) ([]string, error) {
	// Open the file for reading
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string

	// Create a scanner to read lines from the file
	scanner := bufio.NewScanner(file)

	// Read lines one by one

	for scanner.Scan() {
		line := scanner.Text()

		lines = append(lines, line)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
