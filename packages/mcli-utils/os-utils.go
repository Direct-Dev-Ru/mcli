package mcliutils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// GetIntEnv retrieves the integer value of an environment variable specified by the 'key'.
// If the environment variable is not set, its value is an empty string, or if it cannot
// be converted to an integer, it returns the provided 'defaultValue' instead.
//
// Parameters:
//   - key: The name of the environment variable to retrieve.
//   - defaultValue: The default value to return if the environment variable is not set,
//     its value is empty, or if it cannot be converted to an integer.
//
// Returns:
//   - int: The integer value of the environment variable if it is set, not empty, and
//     can be successfully converted to an integer; otherwise, it returns the
//     'defaultValue'.
//
// Example Usage:
// Get the integer value of the "MAX_CONNECTIONS" environment variable, or use 10 if it's not set,
// empty, or not a valid integer.
//
//	maxConnections := GetIntEnv("MAX_CONNECTIONS", 10)
func GetIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); len(value) == 0 {
		return defaultValue
	} else {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		} else {
			return defaultValue
		}
	}
}

// GetStrEnv retrieves the value of an environment variable specified by the 'key'.
// If the environment variable is not set or its value is an empty string, it returns
// the provided 'defaultValue' instead.
//
// Parameters:
//   - key: The name of the environment variable to retrieve.
//   - defaultValue: The default value to return if the environment variable is not set
//     or its value is empty.
//
// Returns:
//   - string: The value of the environment variable if it is set and not empty, or the
//     'defaultValue' if the environment variable is not set or empty.
//
// Example Usage:
// Get the value of the "API_KEY" environment variable, or use "defaultApiKey" if it's not set or empty.
//
//	apiKey := GetStrEnv("API_KEY", "defaultApiKey")
func GetStrEnv(key, defaultValue string) string {
	if value := os.Getenv(key); len(value) == 0 {
		return defaultValue
	} else {
		return value
	}
}

// IsExistsAndCreate checks if a file or directory exists at the specified path.
// If it exists, it returns true along with the type (file/directory).
// If it doesn't exist and 'create' is set to true, it creates the necessary
// file or directory. If 'asFile' is true, it creates a file; otherwise, it creates
// a directory.
//
// Parameters:
//   - pathParam: The path to check/create.
//   - create: Indicates whether to create the file/directory if it doesn't exist.
//   - asFile: Indicates whether to create a file (if true) or a directory (if false).
//
// Returns:
//   - bool: Indicates if the file/directory exists or was successfully created.
//   - string: The type of the item (file/directory).
//   - error: Returns an error if any occurred during the process.
//
// Example Usage:
//
//	exists, itemType, err := IsExistsAndCreate("/path/to/directory", true, false)
//	if err != nil {
//	  log.Fatal(err)
//	}
//	if exists {
//	  fmt.Printf("%s already exists as a %s\n", pathParam, itemType)
//	} else {
//	  fmt.Printf("%s created as a %s\n", pathParam, itemType)
//	}
func IsExistsAndCreate(pathParam string, create, asFile bool) (bool, string, error) {

	_, err := os.Stat(pathParam)
	if err == nil {
		// path exists - returning now
		fileInfo, _ := os.Stat(pathParam)
		if fileInfo.IsDir() {
			return true, "directory", nil
		} else {
			return true, "file", nil
		}
	}

	if os.IsNotExist(err) && create {

		pathToAnalyse := strings.TrimSpace(pathParam)
		pathToAnalyse = strings.ReplaceAll(pathToAnalyse, `\`, "/")
		itemType := "directory"
		fileName := path.Base(pathToAnalyse)
		pathToCreate := pathToAnalyse
		if asFile {
			withoutLastPart := path.Dir(pathToAnalyse)
			pathToCreate = withoutLastPart
		}

		// create directory
		err := os.MkdirAll(pathToCreate, os.ModePerm)
		if err != nil {
			return false, "", err
		}
		if asFile {
			itemType = "file"
			// create empty file
			err = os.WriteFile(filepath.Join(pathToCreate, fileName), []byte{}, os.ModePerm)
			if err != nil {
				return false, "", err
			}
		}
		return true, itemType, nil
	}
	return false, "", nil
}

// RunExternalCmdsPiped runs a series of external commands in a pipeline, passing the
// output of each command as input to the next. It returns the combined standard output
// of the last command and any error encountered during the execution of the commands.
//
// Parameters:
//   - stdinStr: The standard input string to be passed to the first command in the pipeline.
//   - errorPrefix: A prefix to be used in error messages. If empty, a default error message
//     will be generated.
//   - commands: A slice of command argument slices, where each inner slice represents a
//     command and its arguments.
//
// Returns:
//   - string: The combined standard output of the last command in the pipeline.
//   - error: An error, if one occurred during the execution of the commands.
//
// Example Usage:
//
// Run a series of commands in a pipeline and get the result
//
//	out, err := RunExternalCmdsPiped("input data", "pipeline error", [][]string{
//	    {"command1", "arg1", "arg2"},
//	    {"command2", "arg3"},
//	    {"command3"},
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(out)
func RunExternalCmdsPiped(stdinStr, errorPrefix string, commands [][]string) (string, error) {
	var err error
	if len(errorPrefix) == 0 {
		errorPrefix = fmt.Sprintf("error occured in %v commands", "pipe of")
	}
	if len(commands) < 2 {
		if err != nil {
			return "", fmt.Errorf("%v: %v ", errorPrefix, "at least two commands are required")
		}
	}
	var outBuf, errBuf bytes.Buffer
	// cmd.Stdout = &outBuf
	// cmd.Stderr = &errBuf

	var cmd []*exec.Cmd

	// Create the command objects
	for _, c := range commands {
		cmd = append(cmd, exec.Command(c[0], c[1:]...))
	}

	// Connect the commands in a pipeline
	for i := 0; i < len(cmd)-1; i++ {
		currCmd := cmd[i]
		if len(stdinStr) > 0 && i == 0 {
			currCmd.Stdin = strings.NewReader(stdinStr)
		}
		nextCmd := cmd[i+1]

		pipe, err := currCmd.StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("%v: error creating pipe: %w ", errorPrefix, err)
		}
		nextCmd.Stdin = pipe
	}

	// Set the last command's stdout to os.Stdout
	lastCmd := cmd[len(cmd)-1]
	lastCmd.Stdout = &outBuf
	lastCmd.Stderr = &errBuf

	// Start the commands in reverse order
	for i := len(cmd) - 1; i >= 0; i-- {
		err = cmd[i].Start()
		if err != nil {
			return "", fmt.Errorf("%v: error starting pipe: %w ", errorPrefix, err)
		}
	}

	// Wait for the commands to finish
	for _, c := range cmd {
		err = c.Wait()
		if err != nil {
			return "", fmt.Errorf("%v: %v < details: (%v) >", errorPrefix, err, errBuf.String())
		}
	}
	if len(errBuf.String()) > 0 {
		return "", fmt.Errorf("%v: %v < details: (%v) >", errorPrefix, err, errBuf.String())
	}
	return outBuf.String(), nil
}

// RunExternalCmd runs an external command and returns its standard output.
// It optionally takes a string to be used as the command's standard input.
// If the command exits with a non-zero status, an error is returned.
//
// Parameters:
//   - stdinString: String to be used as the command's standard input.
//   - errorPrefix: Prefix to be used in case of an error message.
//   - commandName: The name of the command to execute.
//   - commandArgs: Additional arguments for the command.
//
// Returns:
//   - string: The standard output of the command.
//   - error: An error if the command exits with a non-zero status.
//
// Example usage:
//
//	out, err := RunExternalCmd("", "error occurred while running external command", "ls", "-lah")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(out)
func RunExternalCmd(stdinString, errorPrefix string, commandName string,
	commandArgs ...string) (string, error) {
	cmd := exec.Command(commandName, commandArgs...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if len(stdinString) > 0 {
		cmd.Stdin = strings.NewReader(stdinString)
	}
	err := cmd.Run()
	if len(errorPrefix) == 0 {
		errorPrefix = fmt.Sprintf("error occured in %v command", commandName)
	}
	if err != nil {
		return "", fmt.Errorf("%v: %v < details: (%v) >", errorPrefix, err, errBuf.String())
	}
	return outBuf.String(), nil
}
