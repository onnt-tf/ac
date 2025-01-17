package util

import (
	"encoding/json"
	"os"
)

// ReadJSON reads a JSON file and unmarshals it into the provided interface.
func ReadJSON(filePath string, v interface{}) error {
	byteValue, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(byteValue, v)
}
