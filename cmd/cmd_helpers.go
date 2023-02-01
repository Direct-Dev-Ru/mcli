package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

func GetStringParam(param string, cmd *cobra.Command, fromConfig string) (string, error) {

	paramValue, err := cmd.Flags().GetString(param)
	isParamSet := cmd.Flags().Lookup(param).Changed
	if !isParamSet && len(fromConfig) > 0 {
		paramValue = fromConfig
	}
	return paramValue, err
}

func IsPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Created with chatGTP assistance
// The function pathExists takes in two arguments:
// path is the path that you want to check if it exists or not.
// create is a bool value, if it is true, the function will create the directory if it doesn't exist.
// It uses the os.Stat() function to check if the path exists, and the os.IsNotExist() function to check if the error is because the path does not exist. If the path exists, it uses os.Stat() again to check whether it is a directory or a file. If it doesn't exist and you passed create as true, it uses os.MkdirAll() to create the directory.
// It returns a tuple with three values:
// bool value that tells whether the path exists or not
// error if there is any
// string value that tells whether the path is file or directory.
func IsPathExistsAndCreate(pathParam string, create bool) (bool, string, error) {

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
			// fist we must make make decision either it is dir or a file
			pathToAnalyse := strings.TrimSpace(pathParam)
			// pathToAnalyse = strings.TrimLeft(pathToAnalyse, ".")
			pathToAnalyse = strings.ReplaceAll(pathToAnalyse, `\`, "/")
			lastPart := path.Base(pathToAnalyse)
			withoutLastPart := path.Dir(pathToAnalyse)
			pathToCreate := pathToAnalyse
			itemType := "directory"
			// if there are extension - it is file and we should create only dir path
			if strings.Contains(lastPart, ".") && !(strings.HasSuffix(lastPart, ".d") && GlobalMap["OS"] == "linux") {
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
