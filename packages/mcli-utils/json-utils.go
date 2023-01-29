package mcliutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/mgo.v2/bson"
)

func PrettyJsonEncode(data interface{}, out io.Writer) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	return nil
}

func PrettyJsonEncodeToString(data interface{}) (string, error) {

	var buffer bytes.Buffer
	err := PrettyJsonEncode(data, &buffer)

	return buffer.String(), err
}

func ConvertJsonToBson(inputFilePath, outputFilePath string) error {
	// Open the input file
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer inputFile.Close()

	// Decode the input file's JSON data
	var data interface{}
	err = json.NewDecoder(inputFile).Decode(&data)
	if err != nil {
		return fmt.Errorf("error decoding input file: %v", err)
	}

	// Marshal the Go object into BSON data
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling json data to bson: %v", err)
	}

	// Write the BSON data to the destination file
	err = os.WriteFile(outputFilePath, bsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing bson data to output file: %v", err)
	}

	// Get the permissions of the source file
	srcInfo, err := os.Stat(inputFilePath)
	if err != nil {
		return fmt.Errorf("error reading stats for input file: %v", err)
	}
	srcMode := srcInfo.Mode()

	err = os.Chmod(outputFilePath, srcMode)
	if err != nil {
		return fmt.Errorf("error chmod for output file: %v", err)
	}
	return nil
}
