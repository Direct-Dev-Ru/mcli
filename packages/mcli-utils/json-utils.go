package mcliutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
)

func PrettyPrintMap(inputMap map[string]interface{}) string {
	prettyJSON, err := json.MarshalIndent(inputMap, "", "  ")
	if err != nil {
		return ""
	}
	return string(prettyJSON)
}

func PrettyJsonEncode(data interface{}, out io.Writer) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	return nil
}
func JsonEncode(data interface{}, out io.Writer) error {
	enc := json.NewEncoder(out)

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
func JsonEncodeToString(data interface{}) (string, error) {

	var buffer bytes.Buffer
	err := JsonEncode(data, &buffer)

	return buffer.String(), err
}

func YamlEncodeToString(data interface{}) (string, error) {

	yamlData, err := yaml.Marshal(data)

	return string(yamlData), err

}

func InterfaceToYamlString(data interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

func PrintSliceAsTable(data interface{}, columnLimit, nSpaces int) {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		fmt.Println("Input is not a slice")
		return
	}

	numRows := val.Len()
	if numRows == 0 {
		fmt.Println("No data to print")
		return
	}

	elemType := val.Index(0).Type()

	// Get field names
	var fieldNames []string
	columnWidths := make([]int, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		fieldNames = append(fieldNames, elemType.Field(i).Name)
		columnWidths[i] = len(elemType.Field(i).Name)
	}

	// Calculate maximum column widths
	for i := 0; i < numRows; i++ {
		elem := val.Index(i)
		for j := 0; j < elemType.NumField(); j++ {
			fieldValue := elem.Field(j).Interface()
			fieldLength := len(fmt.Sprintf("%v", fieldValue))
			if fieldLength > columnWidths[j] {
				columnWidths[j] = fieldLength
			}
		}
	}
	separatorSlice := make([]string, 0)
	// Print table header
	for i, fieldName := range fieldNames {
		fmt.Printf("%-*.*s ", columnWidths[i]+nSpaces, columnWidths[i]+nSpaces, fieldName)
		separatorSlice = append(separatorSlice, strings.Repeat("-", len(fieldName)))
	}
	fmt.Println()

	for i, fieldName := range separatorSlice {
		fmt.Printf("%-*.*s ", columnWidths[i]+nSpaces, columnWidths[i]+nSpaces, fieldName)
	}
	fmt.Println()

	// fmt.Println(strings.Repeat("-", columnWidths))
	// Print table rows
	for i := 0; i < numRows; i++ {
		elem := val.Index(i)
		for j := 0; j < elemType.NumField(); j++ {
			fieldValue := elem.Field(j).Interface()
			fmt.Printf("%-*.*s ", columnWidths[j]+nSpaces, columnWidths[j]+nSpaces, fmt.Sprintf("%v", fieldValue))
		}
		fmt.Println()
	}

	// fmt.Println(strings.Repeat("-", columnLimit*len(columnWidths)))
}

func YamlStringToInterface(yamlData []byte) (interface{}, error) {
	var data interface{}
	err := yaml.Unmarshal(yamlData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
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
