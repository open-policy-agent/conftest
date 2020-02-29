package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GetConfigurations parses and returns the configurations given in the file list
func GetConfigurations(ctx context.Context, input string, fileList []string) (map[string]interface{}, error) {
	var fileConfigs []ConfigDoc
	for _, fileName := range fileList {
		var err error
		var config io.ReadCloser

		config, err = getConfig(fileName)
		if err != nil {
			return nil, fmt.Errorf("get config: %w", err)
		}

		configDoc := ConfigDoc{
			ReadCloser: config,
			Filepath:   fileName,
		}

		fileConfigs = append(fileConfigs, configDoc)
	}

	unmarshaledConfigs, err := BulkUnmarshal(fileConfigs, input)
	if err != nil {
		return nil, fmt.Errorf("bulk unmarshal: %w", err)
	}

	return unmarshaledConfigs, nil
}

func getConfig(fileName string) (io.ReadCloser, error) {
	if fileName == "-" {
		config := ioutil.NopCloser(bufio.NewReader(os.Stdin))
		return config, nil
	}

	filePath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("get abs: %w", err)
	}

	config, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return config, nil
}
