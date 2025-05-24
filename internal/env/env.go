package env

import (
	"os"
	"strconv"
)

func GetString(key string, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return value
}

func GetInt(key string, defaultValue int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return valueInt
}
