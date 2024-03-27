package utils

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func GetYAML(input any) ([]byte, error) {
	output := bytes.NewBuffer(make([]byte, 0, 1024*10))
	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	err := encoder.Encode(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling to YAML: %s", err)
	}
	return output.Bytes(), nil
}

func CreateDirIfNotExist(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating directory: %s", err)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
