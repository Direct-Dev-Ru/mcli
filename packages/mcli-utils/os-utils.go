package mcliutils

import (
	"os"
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
