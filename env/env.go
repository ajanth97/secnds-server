package env

import (
	"os"
	"strconv"

	"github.com/labstack/gommon/log"
)

func Get(envName string) string {
	value, exists := os.LookupEnv(envName)
	if !exists {
		log.Fatalf("Enviroment variable %s not found", envName)
	}
	return value
}

func GetInt(envName string) int {
	stringValue := Get(envName)
	intValue, err := strconv.Atoi(stringValue)
	if err != nil {
		log.Fatalf("Enviroment variable %s cannot be converted to int", envName)
	}
	return intValue
}

func GetByte(envName string) []byte {
	stringValue := Get(envName)
	byteValue := []byte(stringValue)
	return byteValue
}
