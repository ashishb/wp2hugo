package utils

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func GetYAML(input any) ([]byte, error) {
	output := bytes.NewBuffer(make([]byte, 0, 1024*10))
	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	err := encoder.Encode(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling to YAML: %w", err)
	}
	return output.Bytes(), nil
}

func CreateDirIfNotExist(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating directory: %w", err)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DirExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
