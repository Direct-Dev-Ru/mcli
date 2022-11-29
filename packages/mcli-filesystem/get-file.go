package mclifilesystem

// https://www.devdungeon.com/content/working-files-go

import (
	"errors"
	"os"
)

type getFileHandler func(string) (*os.File, func(), error)

var getFileForRW getFileHandler = func(filePath string) (*os.File, func(), error) {
	var (
		file *os.File = nil
		err  error    = nil
	)

	if _, err = os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(filePath)
		// file, err = os.OpenFile(filePath, os.O_CREATE, 0666)
	} else {
		file, err = os.OpenFile(filePath, os.O_RDWR, 0666)
	}
	if err != nil {
		return nil, nil, err
	}
	return file, func() {
		file.Close()
	}, nil
}

var getFileForR getFileHandler = func(filePath string) (*os.File, func(), error) {
	var (
		file *os.File = nil
		err  error    = nil
	)

	if _, err = os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return nil, nil, errors.New("file does not exist")
	} else {
		file, err = os.OpenFile(filePath, os.O_RDONLY, 0644)
	}
	if err != nil {
		return nil, nil, err
	}
	return file, func() {
		file.Close()
	}, nil
}
