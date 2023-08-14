package mcliutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
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

func JsonStringToInterface(jsonString string) (interface{}, error) {
	var data interface{}
	decoder := json.NewDecoder(strings.NewReader(jsonString))
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func BsonDataToInterfaceMap(bsonData []byte) (interface{}, error) {
	var data map[string]interface{} = make(map[string]interface{})
	err := bson.Unmarshal(bsonData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ConvertBsonToJson(inputFilePath, outputFilePath string) error {
	// Open the input file
	inputFile, err := os.ReadFile(inputFilePath)
	if err != nil {
		return fmt.Errorf("error reading input bson file: %v", err)
	}

	// Decode the input file's BSON data
	var data interface{}
	err = bson.Unmarshal(inputFile, &data)
	if err != nil {
		return fmt.Errorf("error decoding input bson file: %v", err)
	}

	// Marshal the Go object into JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data to json: %v", err)
	}

	// Write the JSON data to the destination file
	err = os.WriteFile(outputFilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing json data to output file: %v", err)
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
		return fmt.Errorf("error writing bson data to output file: %w", err)
	}

	// Get the permissions of the source file
	srcInfo, err := os.Stat(inputFilePath)
	if err != nil {
		return fmt.Errorf("error reading stats for input file: %w", err)
	}
	srcMode := srcInfo.Mode()

	err = os.Chmod(outputFilePath, srcMode)
	if err != nil {
		return fmt.Errorf("error chmod for output file: %w", err)
	}
	return nil
}

func ConvertJsonToYaml(inputFilePath, outputFilePath string) error {
	// Open the input file
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer inputFile.Close()

	// Decode the input file's JSON data
	var data interface{}
	err = json.NewDecoder(inputFile).Decode(&data)
	if err != nil {
		return fmt.Errorf("error decoding input file: %w", err)
	}

	// Marshal the Go object into BSON data
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling json data to yaml: %w", err)
	}

	// Write YAML data to the destination file
	err = os.WriteFile(outputFilePath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error writing yaml data to output file: %w", err)
	}

	// Get the permissions of the source file
	srcInfo, err := os.Stat(inputFilePath)
	if err != nil {
		return fmt.Errorf("error reading stats for input json file: %w", err)
	}
	srcMode := srcInfo.Mode()

	err = os.Chmod(outputFilePath, srcMode)
	if err != nil {
		return fmt.Errorf("error chmod for output yaml file: %v", err)
	}
	return nil
}
