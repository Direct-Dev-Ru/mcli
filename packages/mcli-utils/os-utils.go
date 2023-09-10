package mcliutils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
)

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
func GetStrEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); len(value) == 0 {
		return defaultValue
	} else {
		return value
	}
}

func IsExistsAndCreate(pathParam string, create bool) (bool, string, error) {

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

	if os.IsNotExist(err) {
		// path does not exist
		if create {
			// fist we must make decision either it is dir or a file
			pathToAnalyse := strings.TrimSpace(pathParam)
			// pathToAnalyse = strings.TrimPrefix(pathToAnalyse, ".")
			pathToAnalyse = strings.ReplaceAll(pathToAnalyse, `\`, "/")
			lastPart := path.Base(pathToAnalyse)
			withoutLastPart := path.Dir(pathToAnalyse)
			pathToCreate := pathToAnalyse
			itemType := "directory"
			// if there are extension - it is file and we should create only dir path
			if strings.Contains(lastPart, ".") && !(strings.HasSuffix(lastPart, ".d") && runtime.GOOS == "linux") {
				pathToCreate = withoutLastPart
				itemType = "file"
			}
			// create directory
			// println(pathToAnalyse, lastPart, withoutLastPart, pathToCreate)
			err := os.MkdirAll(pathToCreate, os.ModePerm)
			if err != nil {
				return false, "", err
			}
			return true, itemType, nil
		}
		return false, "", nil
	}
	return false, "", nil
}

func RunExternalCmdsPiped(stdinStr, errorPrefix string, commands [][]string) (string, error) {
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
	var err error

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

func RunExternalCmd(stdinString, errorPrefix string, commandName string,
	commandArgs ...string) (string, error) {
	// Apply the Kubernetes manifest using the 'kubectl' command
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
