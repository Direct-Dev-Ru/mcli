package mcliutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsExistsAndCreate(t *testing.T) {
	// Test case 1: Path dont exists and it's a directory - creating it and then checking again
	dirPath := "/tmp/test_dir"
	os.Remove(dirPath)

	exists, itemType, err := IsExistsAndCreate(dirPath, true, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists || itemType != "directory" {
		t.Errorf("Expected directory, got %s", itemType)
	}
	exists, itemType, err = IsExistsAndCreate(dirPath, false, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists || itemType != "directory" {
		t.Errorf("Expected directory, got %s", itemType)
	}

	// Test case 2: Path exists and it's a file
	filePath := "/tmp/test_dir/test_file.txt"
	os.Remove(filePath)
	_, err = os.Create(filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	exists, itemType, err = IsExistsAndCreate(filePath, false, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists || itemType != "file" {
		t.Errorf("Expected file, got %s", itemType)
	}

	// Test case 3: Path does not exist and create is false
	nonexistentPath := "nonexistent_dir"
	os.Remove(nonexistentPath)
	exists, itemType, err = IsExistsAndCreate(nonexistentPath, false, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if exists && itemType == "directory" {
		t.Errorf("Expected false as not existing directory, got %v", exists)
	}

	// Test case 4: Path to file not exists and create is true
	filePath = filepath.Join(dirPath, "test_noexisting_file.txt")
	os.Remove(dirPath)
	os.Remove(filePath)
	exists, itemType, err = IsExistsAndCreate(filePath, true, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists || itemType != "file" {
		t.Errorf("Expected file creating, got %v %s", exists, itemType)
	}
	exists, itemType, err = IsExistsAndCreate(filePath, false, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists || itemType != "file" {
		t.Errorf("Expected file existing after creating, got %v %s", exists, itemType)
	}

	// Clean up created files and directories
	os.Remove(dirPath)
	os.Remove(filePath)
	os.Remove(nonexistentPath)
}

func TestRunExternalCmd(t *testing.T) {
	stdinString := "input string"
	errorPrefix := "error occurred while running external command"
	commandName := "echo"
	commandArgs := []string{"-n", "test"}

	output, err := RunExternalCmd(stdinString, errorPrefix, commandName, commandArgs...)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedOutput := "test"
	if output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, output)
	}
}
