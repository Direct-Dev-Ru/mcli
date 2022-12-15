package mclifilesystem

import (
	"bytes"
	"compress/gzip"
	"io"
)

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

func UnZipData(data []byte) ([]byte, error) {

	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return nil, err
	}
	s, err := io.ReadAll(r)
	return s, err
}
