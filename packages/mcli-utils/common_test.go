package mcliutils

import (
	"reflect"
	"testing"
)

func TestStructToMapStringValues(t *testing.T) {
	p := struct {
		Name   string
		Age    int
		Email  string
		Active bool
	}{
		Name:   "John Doe",
		Age:    30,
		Email:  "john@example.com",
		Active: true,
	}
	ptrP := &p

	result, _ := StructToMapStringValues(*ptrP)

	expected := map[string]string{
		"Name":  "John Doe",
		"Email": "john@example.com",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result is incorrect. Got %v, want %v", result, expected)
	}
}
