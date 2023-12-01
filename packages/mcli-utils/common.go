package mcliutils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// TiifFunc- returns funcTrue or funcFalse result as interface{}, error tulip accordinly by the ifCase bool param
// args ...interface{} is optional args for true and false functions

// TiifFunc is a conditional function executor that takes a boolean 'ifCase' as a condition,
// two functions 'funcTrue' and 'funcFalse' to execute based on the condition, and optional
// arguments 'args' that are passed to the chosen function.
//
// Parameters:
//   - ifCase: The boolean condition that determines which function to execute.
//   - funcTrue: The function to execute if 'ifCase' is true.
//   - funcFalse: The function to execute if 'ifCase' is false.
//   - args: Optional arguments that are passed to the chosen function.
//
// Returns:
//   - interface{}: The result of executing the chosen function.
//   - error: An error, if one occurred during the execution of the chosen function.
//
// Example Usage:
//
//	result, err := TiifFunc(condition, func(args ...interface{}) (interface{}, error) {
//		return true, nil
//	}, func(args ...interface{}) (interface{}, error) {
//		return args[0].(bool) || args[1].(bool) || args[1].(bool), nil
//	}, true, true, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result)
func TiifFunc(ifCase bool, funcTrue func(args ...interface{}) (interface{}, error),
	funcFalse func(args ...interface{}) (interface{}, error), args ...interface{}) (interface{}, error) {

	if ifCase {
		return funcTrue(args)
	}
	return funcFalse(args)
}

// Tiif returns one of two values based on the boolean condition 'ifCase'. If 'ifCase' is true,
// it returns 'trueValue'; otherwise, it returns 'falseValue'.
//
// Parameters:
//   - ifCase: The boolean condition that determines which value to return.
//   - trueValue: The value to return if 'ifCase' is true.
//   - falseValue: The value to return if 'ifCase' is false.
//
// Returns:
//   - interface{}: The chosen value based on the condition.
//
// Example Usage:
//
//	result := Tiif(condition, trueResult, falseResult)
//	fmt.Println(result)
func Tiif(ifCase bool, trueValue interface{}, falseValue interface{}) interface{} {
	if ifCase {
		return trueValue
	}
	return falseValue
}

// StructToMapStringValues converts a struct to a map containing only string values.
//
// It iterates over the fields of the input struct, checks if they are of string type,
// and adds them to the resulting map.
//
// Parameters:
//   - input: The input struct to be converted.
//
// Returns:
//   - map[string]string: A map containing field names as keys and their string values.
//
// Example Usage:
//
//	type Person struct {
//		Name   string
//		Age    int
//		Email  string
//		Active bool
//	}
//
//	func main() {
//		p := Person{
//			Name:   "John Doe",
//			Age:    30,
//			Email:  "john@example.com",
//			Active: true,
//		}
//
//		result := structToMapStringValues(p)
//
//		fmt.Println(result)
//	}
//
// This example will output: map[Email:john@example.com Name:John Doe]
func StructToMapStringValues(input interface{}) map[string]*string {
	result := make(map[string]*string)
	v := reflect.ValueOf(input)

	if v.Kind() == reflect.Struct {
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)

			if field.Kind() == reflect.String {
				stringField := field.String()
				result[fieldType.Name] = &stringField
			}
		}
	}

	return result
}

func PrintAsTable(data map[string][]string, columnDivider string) {

	// Find the maximum length for each column
	columnWidths := make(map[string]int)
	var keys []string
	var rowCount int = 0
	for key, colValues := range data {
		// Add the key to the list if it's not already present
		if _, exists := columnWidths[key]; !exists {
			keys = append(keys, key)
		}
		if len(colValues) > rowCount {
			rowCount = len(colValues)
		}
		// Update the column width if the current value is wider
		for _, value := range colValues {
			if len(value) > columnWidths[key] {
				columnWidths[key] = len(value)
			}
		}
	}

	// Sort the keys to maintain consistent order
	sort.Strings(keys)

	// Calculate the total width of the header
	totalHeaderWidth := 0
	for _, key := range keys {
		totalHeaderWidth += columnWidths[key] + len(columnDivider)
	}

	// Print header
	headerDivider := strings.Repeat(" ", totalHeaderWidth-1) // Subtract 1 to match the last column
	for _, key := range keys {
		fmt.Printf("%-*s%s", columnWidths[key]+len(columnDivider), key, columnDivider)
	}
	fmt.Println() // Move to the next line after the header
	_ = headerDivider
	fmt.Println(headerDivider)

	// Print data rows
	for i := 0; i < rowCount; i++ {
		for _, key := range keys {
			fmt.Printf("%-*s%s", columnWidths[key]+len(columnDivider), data[key][i], columnDivider)
		}
		fmt.Println() // Move to the next line after each row
	}
}
