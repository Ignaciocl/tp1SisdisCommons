package configloader

import (
	"fmt"
	"io"
	"os"
)

func GetConfigFileAsBytes(filepath string) ([]byte, error) {
	configFile, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %s", err)
	}

	configFileBytes, err := io.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	return configFileBytes, nil
}
