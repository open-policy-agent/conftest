package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GetConfigurations parses and returns the configurations given in the file list
func GetConfigurations(ctx context.Context, input string, fileList []string) (map[string]interface{}, error) {
	totalCfgs := make(map[string]interface{})
	fileSchema := make(map[string][]ConfigDoc)

	for _, fileName := range fileList {
		var err error
		var cfg io.ReadCloser

		fileType, err := getFileType(input, fileName)
		if err != nil {
			return nil, fmt.Errorf("get file type: %w", err)
		}

		cfg, err = getConfig(fileName)
		if err != nil {
			return nil, fmt.Errorf("get config: %w", err)
		}

		fileSchema[fileType] = append(fileSchema[fileType], ConfigDoc{
			ReadCloser: cfg,
			Filepath:   fileName,
		})
	}

	for fileType, cfgFiles := range fileSchema {
		cfgManager, err := NewConfigManager(fileType)
		if err != nil {
			return nil, fmt.Errorf("create config manager: %w", err)
		}

		cfgs, err := cfgManager.BulkUnmarshal(cfgFiles)
		if err != nil {
			return nil, fmt.Errorf("bulk unmarshal: %w", err)
		}

		// concatenate configurations
		for k, v := range cfgs {
			totalCfgs[k] = v
		}
	}

	return totalCfgs, nil

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

func getFileType(inputFileType, fileName string) (string, error) {
	if inputFileType != "" {
		return inputFileType, nil
	}

	if fileName == "-" && inputFileType == "" {
		return "yaml", nil
	}

	if fileName != "-" {
		fileType := ""
		if strings.Contains(fileName, ".") {
			fileType = strings.TrimPrefix(filepath.Ext(fileName), ".")
		} else {
			ss := strings.SplitAfter(fileName, "/")
			fileType = ss[len(ss)-1]
		}

		return fileType, nil
	}

	return "", fmt.Errorf("unsupported file type")
}
