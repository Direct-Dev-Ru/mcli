package mcliutils

import (
	"os"
	"strconv"
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
