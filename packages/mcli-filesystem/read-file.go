package mclifilesystem

import (
	"errors"
	"io"
	"os"
)

// https://www.devdungeon.com/content/working-files-go

type GetFileContentType func(string) ([]byte, error)

var GetFile GetFileContentType = func(string) ([]byte, error) {
	return nil, nil
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

// // make a buffer to keep chunks that are read
// buf := make([]byte, 1024)
// for {
// 	// read a chunk
// 	n, err := fi.Read(buf)
// 	if err != nil && err != io.EOF {
// 		panic(err)
// 	}
// 	if n == 0 {
// 		break
// 	}

// 	// write a chunk
// 	if _, err := fo.Write(buf[:n]); err != nil {
// 		panic(err)
// 	}
// }
