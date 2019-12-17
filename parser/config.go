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

var (
	supporedFileTypes = []string{
		"toml",
		"tf",
		"hcl",
		"hcl2",
		"cue",
		"ini",
		"yaml",
		"yml",
		"json",
		"edn",
		"xml",
		"Dockerfile",
	}
)

// GetConfigurations parses and returns the configurations given in the file list
func GetConfigurations(ctx context.Context, input string, fileList []string) (map[string]interface{}, error) {
	var configFiles []ConfigDoc
	var fileType string

	for _, fileName := range fileList {
		var err error
		var config io.ReadCloser

		fileType, err = getFileType(input, fileName)
		if err != nil {
			return nil, fmt.Errorf("get file type: %w", err)
		}

		config, err = getConfig(fileName)
		if err != nil {
			return nil, fmt.Errorf("get config: %w", err)
		}

		configFiles = append(configFiles, ConfigDoc{
			ReadCloser: config,
			Filepath:   fileName,
		})
	}

	configManager, err := NewConfigManager(fileType)
	if err != nil {
		return nil, fmt.Errorf("create config manager: %w", err)
	}

	configurations, err := configManager.BulkUnmarshal(configFiles)
	if err != nil {
		return nil, fmt.Errorf("bulk unmarshal: %w", err)
	}

	return configurations, nil

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
		if !isSupported(inputFileType) {
			return "", fmt.Errorf("unsupported file type")
		}

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

		if !isSupported(fileType) {
			return "", fmt.Errorf("unsupported file type")
		}

		return fileType, nil
	}

	return "", fmt.Errorf("unsupported file type")
}

func isSupported(fileType string) bool {
	for _, t := range supporedFileTypes {
		if fileType == t {
			return true
		}
	}

	return false
}
