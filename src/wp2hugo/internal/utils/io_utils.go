package utils

import (
	"fmt"
	"os"
)

func CreateDirIfNotExist(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating directory: %s", err)
	}
	return nil
}
