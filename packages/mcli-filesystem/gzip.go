package mclifilesystem

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipData - compress data ([]byte) and return []byte with compressed bytes
func ZipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// UnZipData - unzip []byte slice contains compressed data
func UnZipData(data []byte) ([]byte, error) {
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return nil, err
	}
	s, err := io.ReadAll(r)
	return s, err
}

// SetZipContent - compress [content] - []byte and writes it to the [filename]
func (sc SetFileContentType) SetZipContent(filePath string, content []byte) (int, error) {
	// io.reader from []byte
	r := bytes.NewReader(content)

	writer, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	a := gzip.NewWriter(writer)
	defer a.Close()

	n, err := io.Copy(a, r)
	if err != nil {
		return 0, err
	}
	return int(n), err
}

// ZipFile - compress [source] file and writes it to the [destination] file
func (sc SetFileContentType) ZipFile(source, destination string) error {
	r, err := os.Open(source)
	if err != nil {
		return err
	}

	var sourceFilename string
	sourceFilename = filepath.Base(source)
	if len(destination) == 0 {
		destination = filepath.Join(filepath.Dir(source), fmt.Sprintf("%s.gz", sourceFilename))
	}
	w, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer w.Close()

	if sourceFilename == "." {
		sourceFilename = ""
	}
	a := gzip.NewWriter(w)
	a.Name = sourceFilename
	defer a.Close()

	_, err = io.Copy(a, r)
	if err != nil {
		return err
	}
	return nil
}

// UnZipFile - unzip file given by [source] to [destination]. If [destination] is empty then
// it will be equal Dir of [source] + archive.name if sets ( or error)
func (gc GetFileContentType) UnZipFile(source, destination string) error {
	r, err := os.Open(source)
	if err != nil {
		return err
	}
	defer r.Close()
	a, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer a.Close()

	// check if [destination] does not set
	if len(destination) == 0 {
		destFilename := ""
		if len(a.Name) == 0 {
			sourceFileName := filepath.Base(source)
			if !strings.HasSuffix(sourceFileName, ".gz") {
				destFilename = sourceFileName + ".unzip"
			} else {
				destFilename = strings.TrimSuffix(sourceFileName, ".gz")
			}
		} else {
			destFilename = a.Name
		}
		destination = filepath.Join(filepath.Dir(source), destFilename)
		// fmt.Println(destination, destFilename, a.Name)
	}
	w, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, a)
	if err != nil {
		return err
	}
	return nil
}

// GetUnZipContent - unzip file given by filename to []byte slice
func (gc GetFileContentType) GetUnZipContent(filePath string) ([]byte, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	_, err = io.Copy(w, archive)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
