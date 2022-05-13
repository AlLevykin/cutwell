package store

import (
	"encoding/json"
	"os"
)

func FileToMap(fileName string) map[string]string {
	var m map[string]string
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		m = make(map[string]string)
		return m
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&m); err != nil {
		m = make(map[string]string)
		return m
	}
	return m
}

func MapToFile(m map[string]string, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	return encoder.Encode(&m)
}
