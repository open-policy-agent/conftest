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

// ConfigDoc is an input document to be checked
type ConfigDoc struct {
	ReadCloser io.ReadCloser
	Filepath   string
	Parser     Parser
}

type CustomConfigManager interface {
	GetConfigurations(ctx context.Context, input string, fileList []string) (map[string]interface{}, error)
}

type ConfigManager struct{}

// GetConfigurations parses and returns the configurations given in the file list
func (c *ConfigManager) GetConfigurations(ctx context.Context, input string, fileList []string) (map[string]interface{}, error) {
	var fileConfigs []ConfigDoc
	for _, fileName := range fileList {
		var config io.ReadCloser

		config, err := c.getConfig(fileName)
		if err != nil {
			return nil, fmt.Errorf("get config: %w", err)
		}

		fileType := c.getFileType(fileName, input)
		parser, err := GetParser(fileType)
		if err != nil {
			return nil, fmt.Errorf("get parser: %w", err)
		}

		configDoc := ConfigDoc{
			ReadCloser: config,
			Filepath:   fileName,
			Parser:     parser,
		}

		fileConfigs = append(fileConfigs, configDoc)
	}

	unmarshaledConfigs, err := c.bulkUnmarshal(fileConfigs)
	if err != nil {
		return nil, fmt.Errorf("bulk unmarshal: %w", err)
	}

	return unmarshaledConfigs, nil
}

func (c *ConfigManager) bulkUnmarshal(configList []ConfigDoc) (map[string]interface{}, error) {
	configContents := make(map[string]interface{})
	for _, config := range configList {
		contents, err := ioutil.ReadAll(config.ReadCloser)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}

		var singleContent interface{}
		if err := config.Parser.Unmarshal(contents, &singleContent); err != nil {
			return nil, fmt.Errorf("parser unmarshal: %w", err)
		}

		configContents[config.Filepath] = singleContent
		config.ReadCloser.Close()
	}

	return configContents, nil
}

func (c *ConfigManager) getConfig(fileName string) (io.ReadCloser, error) {
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

func (c *ConfigManager) getFileType(fileName string, input string) string {
	if input != "" {
		return input
	}

	if fileName == "-" {
		return "yaml"
	}

	if filepath.Ext(fileName) == "" {
		return filepath.Base(fileName)
	}

	fileExtension := filepath.Ext(fileName)

	return fileExtension[1:]
}
